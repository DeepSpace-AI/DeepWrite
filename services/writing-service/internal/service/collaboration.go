package service

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/deepwrite/writing-service/internal/repository"
	"github.com/gorilla/websocket"
)

type CollaborationMessage struct {
	Type      string          `json:"type"`
	DocumentID string         `json:"document_id"`
	UserID    string          `json:"user_id"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp int64           `json:"timestamp"`
}

type Client struct {
	ID         string
	UserID     string
	DocumentID string
	Conn       *websocket.Conn
	Hub        *CollaborationHub
	Send       chan []byte
}

type CollaborationHub struct {
	clients    map[string]map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *CollaborationMessage
	mu         sync.RWMutex
	contentRepo repository.ContentRepository
}

func NewCollaborationHub(contentRepo repository.ContentRepository) *CollaborationHub {
	return &CollaborationHub{
		clients:     make(map[string]map[string]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *CollaborationMessage, 256),
		contentRepo: contentRepo,
	}
}

func (h *CollaborationHub) Register(client *Client) {
	h.register <- client
}

func (h *CollaborationHub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *CollaborationHub) Broadcast(msg *CollaborationMessage) {
	h.broadcast <- msg
}

func (h *CollaborationHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.DocumentID] == nil {
				h.clients[client.DocumentID] = make(map[string]*Client)
			}
			h.clients[client.DocumentID][client.ID] = client
			h.mu.Unlock()
			log.Printf("User %s joined document %s", client.UserID, client.DocumentID)
			h.notifyUserJoined(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if docClients, ok := h.clients[client.DocumentID]; ok {
				if _, ok := docClients[client.ID]; ok {
					delete(docClients, client.ID)
					close(client.Send)
					if len(docClients) == 0 {
						delete(h.clients, client.DocumentID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("User %s left document %s", client.UserID, client.DocumentID)
			h.notifyUserLeft(client)

		case msg := <-h.broadcast:
			h.mu.RLock()
			docClients := h.clients[msg.DocumentID]
			h.mu.RUnlock()

			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Failed to marshal message: %v", err)
				continue
			}

			for _, client := range docClients {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					h.mu.Lock()
					delete(h.clients[msg.DocumentID], client.ID)
					h.mu.Unlock()
				}
			}
		}
	}
}

func (h *CollaborationHub) notifyUserJoined(client *Client) {
	msg := &CollaborationMessage{
		Type:       "user_joined",
		DocumentID: client.DocumentID,
		UserID:     client.UserID,
		Payload:    json.RawMessage(`{"message": "user joined"}`),
		Timestamp:  time.Now().Unix(),
	}
	h.broadcastToOthers(client, msg)
}

func (h *CollaborationHub) notifyUserLeft(client *Client) {
	msg := &CollaborationMessage{
		Type:       "user_left",
		DocumentID: client.DocumentID,
		UserID:     client.UserID,
		Payload:    json.RawMessage(`{"message": "user left"}`),
		Timestamp:  time.Now().Unix(),
	}
	h.broadcastToOthers(client, msg)
}

func (h *CollaborationHub) broadcastToOthers(sender *Client, msg *CollaborationMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients[sender.DocumentID] {
		if client.ID != sender.ID {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

func (h *CollaborationHub) GetActiveUsers(documentID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var users []string
	for _, client := range h.clients[documentID] {
		users = append(users, client.UserID)
	}
	return users
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg CollaborationMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		msg.UserID = c.UserID
		msg.DocumentID = c.DocumentID
		msg.Timestamp = time.Now().Unix()

		if msg.Type == "content_update" {
			var payload struct {
				Content map[string]interface{} `json:"content"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				go c.Hub.persistContent(c.DocumentID, payload.Content)
			}
		}

		c.Hub.Broadcast(&msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *CollaborationHub) persistContent(documentID string, content map[string]interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existing, err := h.contentRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		log.Printf("Failed to get content: %v", err)
		return
	}

	if existing != nil {
		existing.Content = content
		existing.UpdatedAt = time.Now()
		if err := h.contentRepo.Update(ctx, documentID, existing); err != nil {
			log.Printf("Failed to update content: %v", err)
		}
	}
}
