package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type SessionEvent struct {
	SessionID string      `json:"session_id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
}

type SessionEventBus struct {
	mu      sync.RWMutex
	clients map[string]map[chan SessionEvent]struct{}
}

var eventBus = &SessionEventBus{
	clients: make(map[string]map[chan SessionEvent]struct{}),
}

func (b *SessionEventBus) Subscribe(userID string) chan SessionEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan SessionEvent, 10)
	if b.clients[userID] == nil {
		b.clients[userID] = make(map[chan SessionEvent]struct{})
	}
	b.clients[userID][ch] = struct{}{}
	log.Printf("[EventBus] Client subscribed for user %s (total clients: %d)", userID, len(b.clients[userID]))
	return ch
}

func (b *SessionEventBus) Unsubscribe(userID string, ch chan SessionEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.clients[userID] != nil {
		delete(b.clients[userID], ch)
		if len(b.clients[userID]) == 0 {
			delete(b.clients, userID)
		}
	}
	close(ch)
	log.Printf("[EventBus] Client unsubscribed for user %s", userID)
}

func (b *SessionEventBus) Publish(userID string, event SessionEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.clients[userID] == nil {
		return
	}

	log.Printf("[EventBus] Publishing event to %d clients for user %s: %+v", len(b.clients[userID]), userID, event)

	for ch := range b.clients[userID] {
		select {
		case ch <- event:
		default:
			log.Printf("[EventBus] Client channel full, skipping event")
		}
	}
}

func PublishSessionEvent(userID, sessionID, eventType string, data interface{}) {
	eventBus.Publish(userID, SessionEvent{
		SessionID: sessionID,
		Type:      eventType,
		Data:      data,
	})
}

type SSEEventHandler struct{}

func (h *SSEEventHandler) Subscribe(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	ch := eventBus.Subscribe(userID)
	defer eventBus.Unsubscribe(userID, ch)

	c.Stream(func(w io.Writer) bool {
		select {
		case event := <-ch:
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("[SSE] Failed to marshal event: %v", err)
				return true
			}
			c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(data)))
			flusher.Flush()
			return true
		case <-c.Request.Context().Done():
			log.Printf("[SSE] Client disconnected for user %s", userID)
			return false
		case <-time.After(30 * time.Second):
			c.Writer.WriteString(": heartbeat\n\n")
			flusher.Flush()
			return true
		}
	})
}
