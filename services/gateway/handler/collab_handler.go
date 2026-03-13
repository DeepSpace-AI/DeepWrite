package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/deepwrite/serivces/gateway/models/document"
	"github.com/deepwrite/serivces/gateway/models/user"
	"github.com/deepwrite/serivces/gateway/pkg/collab"
	"github.com/deepwrite/serivces/gateway/pkg/config"
	gatewaylogger "github.com/deepwrite/serivces/gateway/pkg/logger"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type CollabHandler struct{}

type collabClient struct {
	Conn         *websocket.Conn
	SessionID    string
	UserID       string
	Role         string
	JoinedAt     time.Time
	LastSeenAt   time.Time
	BytesIn      int64
	UpdateCount  int64
	AckLatencies []int64
	Send         chan []byte
	DocumentID   string
	WorkspaceID  string
	DisplayName  string
}

type awarenessEntry struct {
	ClientID  uint64
	Clock     uint64
	StateJSON []byte
}

type collabRoom struct {
	DocumentID        string
	WorkspaceID       string
	Cfg               config.CollabConfig
	Mu                sync.Mutex
	Clients           map[*collabClient]struct{}
	AwarenessByClient map[*collabClient][]byte
	State             document.CollabState
	Updates           []document.CollabUpdate
	Pending           []document.CollabUpdate
	Flushing          bool
	Closed            bool
	FlushTicker       *time.Ticker
	StopCh            chan struct{}
}

type collabManager struct {
	Mu    sync.Mutex
	Rooms map[string]*collabRoom
}

var (
	globalCollabManager = &collabManager{Rooms: map[string]*collabRoom{}}
	upgrader            = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// @Summary      签发协作令牌
// @Description  为当前用户签发一个短期单次使用的协作连接令牌
// @Tags         Document
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "文档ID"
// @Success      200 {object} response.Response "签发成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      403 {object} response.Response "无权限"
// @Failure      404 {object} response.Response "文档不存在"
// @Router       /documents/{id}/collab-token [post]
func (h *CollabHandler) IssueToken(c *gin.Context) {
	cfg := config.GetGlobalConfig()
	if !cfg.Collab.Enabled {
		response.Failed(c, response.ErrorBadRequestCode, "collaboration is disabled")
		return
	}

	documentID := strings.TrimSpace(c.Param("id"))
	if documentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "document id is required")
		return
	}

	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	doc, err := document.GetByID(c.Request.Context(), documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "文档不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取文档失败")
		return
	}

	_, role, err := getWorkspaceRole(c.Request.Context(), doc.WorkspaceID, userID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取工作区信息失败")
		return
	}
	if !canViewWorkspace(role) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限访问文档")
		return
	}

	rawToken, issued, err := document.IssueCollabToken(c.Request.Context(), doc, userID, role, time.Duration(cfg.Collab.TokenTTLSeconds)*time.Second)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "签发协作令牌失败")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"token":      rawToken,
		"expires_at": issued.ExpiresAt,
		"ws_path":    cfg.Collab.WSPath,
		"read_only":  !canEditWorkspace(role),
	})
}

// @Summary      同步协作落地内容
// @Description  客户端定期将当前编辑器 JSON 同步到 documents.content_json
// @Tags         Document
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "文档ID"
// @Param        request body request.SyncCollabContentRequest true "内容同步参数"
// @Success      200 {object} response.Response "同步成功"
// @Failure      400 {object} response.Response "请求参数错误"
// @Failure      403 {object} response.Response "无权限"
// @Router       /documents/{id}/collab/content [post]
func (h *CollabHandler) SyncContent(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	documentID := strings.TrimSpace(c.Param("id"))
	if documentID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "document id is required")
		return
	}

	var req request.SyncCollabContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "Invalid request: "+err.Error())
		return
	}

	doc, err := document.GetByID(c.Request.Context(), documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Failed(c, 404, "文档不存在")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "获取文档失败")
		return
	}

	_, role, err := getWorkspaceRole(c.Request.Context(), doc.WorkspaceID, userID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "获取工作区信息失败")
		return
	}
	if !canEditWorkspace(role) {
		response.Failed(c, response.ErrorForbiddenCode, "无权限编辑文档")
		return
	}

	contentJSON, err := toDatatypesJSON(req.ContentJSON)
	if err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "content_json must be a valid JSON object")
		return
	}

	updated, err := document.UpdateLiveContent(c.Request.Context(), documentID, req.Title, contentJSON)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "同步文档内容失败")
		return
	}

	response.Success(c, response.SuccessCode, updated)
}

func (h *CollabHandler) ConnectWS(c *gin.Context) {
	cfg := config.GetGlobalConfig()
	if !cfg.Collab.Enabled {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 400, "message": "collaboration is disabled"})
		return
	}

	documentID := strings.TrimSpace(c.Param("id"))
	rawToken := strings.TrimSpace(c.Query("token"))
	if documentID == "" || rawToken == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 400, "message": "document id and token are required"})
		return
	}

	issued, err := document.ConsumeCollabToken(c.Request.Context(), rawToken, documentID, cfg.Collab.TokenSingleUse)
	if err != nil {
		gatewaylogger.S().Warnw("collab ws token rejected", "document_id", documentID, "err", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "invalid collab token"})
		return
	}

	viewer, err := user.GetUserByID(c.Request.Context(), issued.UserID)
	if err != nil {
		gatewaylogger.S().Warnw("collab ws user lookup failed", "document_id", documentID, "user_id", issued.UserID, "err", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "invalid collab user"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		gatewaylogger.S().Warnw("collab ws upgrade failed", "document_id", documentID, "user_id", issued.UserID, "err", err)
		return
	}

	client := &collabClient{
		Conn:        conn,
		SessionID:   issued.ID,
		UserID:      issued.UserID,
		Role:        issued.Role,
		JoinedAt:    time.Now(),
		LastSeenAt:  time.Now(),
		Send:        make(chan []byte, 256),
		DocumentID:  issued.DocumentID,
		WorkspaceID: issued.WorkspaceID,
		DisplayName: resolveCollabDisplayName(viewer),
	}

	room, err := globalCollabManager.getOrCreateRoom(c.Request.Context(), issued.DocumentID, issued.WorkspaceID, cfg.Collab)
	if err != nil {
		gatewaylogger.S().Errorw("collab room init failed", "document_id", issued.DocumentID, "workspace_id", issued.WorkspaceID, "err", err)
		_ = conn.Close()
		return
	}
	room.addClient(client)
	gatewaylogger.S().Infow("collab ws connected", "document_id", client.DocumentID, "workspace_id", client.WorkspaceID, "session_id", client.SessionID, "user_id", client.UserID, "role", client.Role)

	// 连接建立后主动请求客户端发送当前状态，并回放已持久化更新。
	gatewaylogger.S().Infow("collab ws enqueue sync step1", "document_id", client.DocumentID, "session_id", client.SessionID)
	client.Send <- collab.BuildSyncFrame(collab.SyncStep1, nil)
	gatewaylogger.S().Infow("collab ws replay updates", "document_id", client.DocumentID, "session_id", client.SessionID)
	room.replayUpdates(client)

	go h.writePump(client, cfg.Collab)
	h.readPump(room, client, cfg.Collab)
}

func (h *CollabHandler) readPump(room *collabRoom, client *collabClient, cfg config.CollabConfig) {
	defer func() {
		gatewaylogger.S().Infow("collab ws disconnected", "document_id", client.DocumentID, "session_id", client.SessionID, "user_id", client.UserID, "updates", client.UpdateCount, "bytes_in", client.BytesIn)
		room.removeClient(client)
		h.persistAudit(room, client)
		close(client.Send)
		_ = client.Conn.Close()
	}()

	_ = client.Conn.SetReadDeadline(time.Now().Add(time.Duration(cfg.HeartbeatTimeoutSeconds) * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.LastSeenAt = time.Now()
		_ = client.Conn.SetReadDeadline(time.Now().Add(time.Duration(cfg.HeartbeatTimeoutSeconds) * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			gatewaylogger.S().Debugw("collab ws read end", "document_id", client.DocumentID, "session_id", client.SessionID, "err", err)
			return
		}
		client.LastSeenAt = time.Now()

		started := time.Now()
		messageType, subtype, payload, err := collab.ParseIncomingFrame(message)
		if err != nil {
			gatewaylogger.S().Warnw("collab frame parse failed", "document_id", client.DocumentID, "session_id", client.SessionID, "bytes", len(message), "err", err)
			continue
		}

		gatewaylogger.S().Infow("collab frame received", "document_id", client.DocumentID, "session_id", client.SessionID, "user_id", client.UserID, "message_type", messageType, "subtype", subtype, "payload_bytes", len(payload))

		switch messageType {
		case collab.MessageSync:
			switch subtype {
			case collab.SyncStep1:
				// 简化实现：由客户端先发完整状态到服务端；这里返回空 step2，后续靠 update 广播收敛。
				gatewaylogger.S().Infow("collab sync step1 received", "document_id", client.DocumentID, "session_id", client.SessionID)
				client.Send <- collab.BuildSyncFrame(collab.SyncStep2, []byte{})
			case collab.SyncStep2, collab.SyncUpdate:
				if !canEditWorkspace(client.Role) {
					gatewaylogger.S().Warnw("collab update dropped by role", "document_id", client.DocumentID, "session_id", client.SessionID, "user_id", client.UserID, "role", client.Role, "subtype", subtype, "payload_bytes", len(payload))
					continue
				}
				room.acceptUpdate(client, subtype, payload)
			}
		case collab.MessageAwareness:
			room.acceptAwareness(client, payload)
		}

		latencyMS := time.Since(started).Milliseconds()
		client.AckLatencies = append(client.AckLatencies, latencyMS)
		if len(client.AckLatencies) > 300 {
			client.AckLatencies = client.AckLatencies[len(client.AckLatencies)-300:]
		}
	}
}

func (h *CollabHandler) writePump(client *collabClient, cfg config.CollabConfig) {
	ticker := time.NewTicker(time.Duration(cfg.HeartbeatPingSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-client.Send:
			if !ok {
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			messageType, subtype, payload, parseErr := collab.ParseIncomingFrame(msg)
			if parseErr != nil {
				gatewaylogger.S().Warnw("collab ws outbound frame parse failed", "document_id", client.DocumentID, "session_id", client.SessionID, "bytes", len(msg), "err", parseErr)
			} else {
				gatewaylogger.S().Infow("collab frame sent", "document_id", client.DocumentID, "session_id", client.SessionID, "user_id", client.UserID, "message_type", messageType, "subtype", subtype, "payload_bytes", len(payload))
			}
			if err := client.Conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				gatewaylogger.S().Debugw("collab ws write failed", "document_id", client.DocumentID, "session_id", client.SessionID, "err", err)
				return
			}
		case <-ticker.C:
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				gatewaylogger.S().Debugw("collab ws ping failed", "document_id", client.DocumentID, "session_id", client.SessionID, "err", err)
				return
			}
		}
	}
}

func (h *CollabHandler) persistAudit(room *collabRoom, client *collabClient) {
	room.Mu.Lock()
	flushSeq := room.State.FlushSeq
	room.Mu.Unlock()

	audit := document.CollabAudit{
		DocumentID:    client.DocumentID,
		WorkspaceID:   client.WorkspaceID,
		SessionID:     client.SessionID,
		UserID:        client.UserID,
		JoinedAt:      client.JoinedAt,
		LeftAt:        time.Now(),
		FlushSeq:      flushSeq,
		UpdateCount:   client.UpdateCount,
		BytesIn:       client.BytesIn,
		AckLatencyP95: percentile95(client.AckLatencies),
	}
	if err := document.CreateCollabAudit(context.Background(), audit); err != nil {
		gatewaylogger.S().Warnw("collab audit persist failed", "document_id", client.DocumentID, "err", err)
	}
}

func percentile95(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]int64, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	idx := int(float64(len(sorted)-1) * 0.95)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func (m *collabManager) getOrCreateRoom(ctx context.Context, documentID, workspaceID string, cfg config.CollabConfig) (*collabRoom, error) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if room, ok := m.Rooms[documentID]; ok {
		return room, nil
	}

	state, err := document.GetOrCreateCollabState(ctx, documentID, workspaceID)
	if err != nil {
		return nil, err
	}

	updates, err := document.ListCollabUpdates(ctx, documentID, 10000)
	if err != nil {
		return nil, err
	}

	room := &collabRoom{
		DocumentID:        documentID,
		WorkspaceID:       workspaceID,
		Cfg:               cfg,
		Clients:           map[*collabClient]struct{}{},
		AwarenessByClient: map[*collabClient][]byte{},
		State:             state,
		Updates:           updates,
		Pending:           make([]document.CollabUpdate, 0, 64),
		FlushTicker:       time.NewTicker(time.Duration(cfg.FlushIntervalMS) * time.Millisecond),
		StopCh:            make(chan struct{}),
	}
	if len(updates) > 0 {
		room.State.LastSeq = updates[len(updates)-1].Seq
	}

	go room.flushLoop()
	m.Rooms[documentID] = room
	return room, nil
}

func (r *collabRoom) flushLoop() {
	for {
		select {
		case <-r.FlushTicker.C:
			r.flush(false)
		case <-r.StopCh:
			r.FlushTicker.Stop()
			return
		}
	}
}

func (r *collabRoom) addClient(client *collabClient) {
	r.Mu.Lock()
	awarenessFrames := make([][]byte, 0, len(r.AwarenessByClient))
	for _, payload := range r.AwarenessByClient {
		awarenessFrames = append(awarenessFrames, append([]byte(nil), payload...))
	}
	r.Clients[client] = struct{}{}
	r.Mu.Unlock()

	for _, payload := range awarenessFrames {
		select {
		case client.Send <- collab.BuildAwarenessFrame(payload):
		default:
		}
	}
}

func (r *collabRoom) removeClient(client *collabClient) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	delete(r.Clients, client)
	delete(r.AwarenessByClient, client)
}

func (r *collabRoom) replayUpdates(client *collabClient) {
	r.Mu.Lock()
	updates := make([]document.CollabUpdate, len(r.Updates))
	copy(updates, r.Updates)
	r.Mu.Unlock()

	for _, update := range updates {
		client.Send <- collab.BuildSyncFrame(collab.SyncUpdate, update.Payload)
	}
}

func (r *collabRoom) acceptUpdate(client *collabClient, subtype uint64, payload []byte) {
	if len(payload) == 0 {
		return
	}

	if subtype != collab.SyncStep2 && subtype != collab.SyncUpdate {
		return
	}

	now := time.Now()

	r.Mu.Lock()
	r.State.LastSeq++
	item := document.CollabUpdate{
		DocumentID:  r.DocumentID,
		WorkspaceID: r.WorkspaceID,
		SessionID:   client.SessionID,
		UserID:      client.UserID,
		Seq:         r.State.LastSeq,
		Payload:     append([]byte(nil), payload...),
		Kind:        "sync",
		CreatedAt:   now,
	}
	r.Pending = append(r.Pending, item)
	r.Updates = append(r.Updates, item)
	r.State.PendingUpdateCount++
	r.State.SnapshotUpdateCount++
	pendingCount := r.State.PendingUpdateCount
	r.Mu.Unlock()

	frame := collab.BuildSyncFrame(collab.SyncUpdate, payload)
	recipients := r.broadcast(client, frame, false)
	gatewaylogger.S().Debugw("collab update accepted", "document_id", r.DocumentID, "session_id", client.SessionID, "user_id", client.UserID, "seq", item.Seq, "payload_bytes", len(payload), "recipients", recipients)

	client.UpdateCount++
	client.BytesIn += int64(len(payload))

	if pendingCount >= int64(r.Cfg.FlushMaxUpdates) {
		go r.flush(true)
	}
}

func (r *collabRoom) acceptAwareness(client *collabClient, payload []byte) {
	canonicalPayload, err := buildCanonicalAwarenessPayload(client, payload)
	if err != nil {
		return
	}

	r.Mu.Lock()
	r.AwarenessByClient[client] = append([]byte(nil), canonicalPayload...)
	r.Mu.Unlock()

	recipients := r.broadcast(client, collab.BuildAwarenessFrame(canonicalPayload), false)
	gatewaylogger.S().Debugw("collab awareness accepted", "document_id", r.DocumentID, "session_id", client.SessionID, "user_id", client.UserID, "payload_bytes", len(canonicalPayload), "recipients", recipients)
}

func resolveCollabDisplayName(u user.User) string {
	name := strings.TrimSpace(u.UserProfile.DisplayName)
	if name != "" {
		return name
	}

	email := strings.TrimSpace(u.Email)
	if email == "" {
		return strings.TrimSpace(u.ID)
	}

	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 0 {
		return email
	}
	local := strings.TrimSpace(parts[0])
	if local != "" {
		return local
	}

	return email
}

func buildCanonicalAwarenessPayload(client *collabClient, payload []byte) ([]byte, error) {
	entries, err := parseAwarenessPayload(payload)
	if err != nil {
		return nil, err
	}

	for i := range entries {
		state := map[string]any{}
		if len(entries[i].StateJSON) > 0 && string(entries[i].StateJSON) != "null" {
			_ = json.Unmarshal(entries[i].StateJSON, &state)
		}

		userState := map[string]any{}
		if existing, ok := state["user"].(map[string]any); ok {
			for k, v := range existing {
				userState[k] = v
			}
		}

		userState["id"] = client.UserID
		userState["name"] = client.DisplayName
		state["user"] = userState

		encoded, marshalErr := json.Marshal(state)
		if marshalErr != nil {
			return nil, marshalErr
		}
		entries[i].StateJSON = encoded
	}

	return encodeAwarenessPayload(entries), nil
}

func parseAwarenessPayload(payload []byte) ([]awarenessEntry, error) {
	idx := 0
	count, err := collab.DecodeVarUint(payload, &idx)
	if err != nil {
		return nil, err
	}

	entries := make([]awarenessEntry, 0, count)
	for i := uint64(0); i < count; i++ {
		clientID, err := collab.DecodeVarUint(payload, &idx)
		if err != nil {
			return nil, err
		}
		clock, err := collab.DecodeVarUint(payload, &idx)
		if err != nil {
			return nil, err
		}
		stateLen, err := collab.DecodeVarUint(payload, &idx)
		if err != nil {
			return nil, err
		}

		end := idx + int(stateLen)
		if end < idx || end > len(payload) {
			return nil, errors.New("invalid awareness payload")
		}

		stateJSON := append([]byte(nil), payload[idx:end]...)
		idx = end

		entries = append(entries, awarenessEntry{
			ClientID:  clientID,
			Clock:     clock,
			StateJSON: stateJSON,
		})
	}

	return entries, nil
}

func encodeAwarenessPayload(entries []awarenessEntry) []byte {
	var buf bytes.Buffer
	buf.Write(collab.EncodeVarUint(uint64(len(entries))))
	for _, entry := range entries {
		buf.Write(collab.EncodeVarUint(entry.ClientID))
		buf.Write(collab.EncodeVarUint(entry.Clock))
		buf.Write(collab.EncodeVarUint(uint64(len(entry.StateJSON))))
		buf.Write(entry.StateJSON)
	}
	return buf.Bytes()
}

func (r *collabRoom) broadcast(sender *collabClient, frame []byte, includeSender bool) int {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	recipients := 0
	for client := range r.Clients {
		if !includeSender && client == sender {
			continue
		}
		select {
		case client.Send <- frame:
			recipients++
		default:
		}
	}
	return recipients
}

func (r *collabRoom) flush(force bool) {
	r.Mu.Lock()
	if r.Flushing {
		r.Mu.Unlock()
		return
	}

	if len(r.Pending) == 0 {
		r.Mu.Unlock()
		return
	}

	if !force && r.State.PendingUpdateCount < int64(r.Cfg.FlushMaxUpdates) {
		// 定时触发时也允许刷盘，force=false 时不提前返回。
	}

	r.Flushing = true
	batch := make([]document.CollabUpdate, len(r.Pending))
	copy(batch, r.Pending)
	r.Pending = r.Pending[:0]
	r.State.PendingUpdateCount = 0
	r.State.FlushSeq++
	flushSeq := r.State.FlushSeq
	snapshotCount := r.State.SnapshotUpdateCount
	r.Mu.Unlock()

	persistErr := retry(5, func() error {
		return document.SaveCollabUpdates(context.Background(), batch)
	})
	if persistErr != nil {
		gatewaylogger.S().Errorw("collab flush failed", "document_id", r.DocumentID, "flush_seq", flushSeq, "err", persistErr)
		r.Mu.Lock()
		r.Pending = append(batch, r.Pending...)
		r.State.PendingUpdateCount += int64(len(batch))
		r.Flushing = false
		r.Mu.Unlock()
		return
	}

	now := time.Now()
	r.Mu.Lock()
	r.State.LastFlushedAt = &now
	if shouldSnapshot(r.Cfg, r.State.LastSnapshotAt, snapshotCount) {
		r.State.SnapshotUpdateCount = 0
		r.State.LastSnapshotAt = &now
		go func(documentID string) {
			_, _, err := document.SaveVersion(context.Background(), document.SaveVersionInput{
				DocumentID: documentID,
				Source:     document.VersionSourceSnapshot,
				Snapshot:   true,
				Summary:    "periodic collab snapshot",
			})
			if err != nil {
				gatewaylogger.S().Warnw("collab snapshot version failed", "document_id", documentID, "err", err)
			}
		}(r.DocumentID)
	}
	state := r.State
	r.Flushing = false
	r.Mu.Unlock()

	if err := document.UpsertCollabState(context.Background(), state); err != nil {
		gatewaylogger.S().Warnw("collab state persist failed", "document_id", r.DocumentID, "err", err)
	}
}

func shouldSnapshot(cfg config.CollabConfig, lastSnapshotAt *time.Time, snapshotUpdates int64) bool {
	if snapshotUpdates >= int64(cfg.SnapshotMaxUpdates) {
		return true
	}
	if lastSnapshotAt == nil {
		return true
	}
	return time.Since(*lastSnapshotAt) >= time.Duration(cfg.SnapshotIntervalSeconds)*time.Second
}

func retry(max int, fn func() error) error {
	var err error
	delay := 200 * time.Millisecond
	for i := 0; i < max; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
		delay *= 2
	}
	return err
}
