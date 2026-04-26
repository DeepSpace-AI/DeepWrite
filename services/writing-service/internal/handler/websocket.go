package handler

import (
	"net/http"
	"time"

	"github.com/deepwrite/writing-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub       *service.CollaborationHub
	jwtSecret string
}

func NewWebSocketHandler(hub *service.CollaborationHub, jwtSecret string) *WebSocketHandler {
	return &WebSocketHandler{
		hub:       hub,
		jwtSecret: jwtSecret,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document ID is required"})
		return
	}

	tokenString := c.Query("token")
	if tokenString == "" {
		tokenString = c.GetHeader("Authorization")
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upgrade connection"})
		return
	}

	client := &service.Client{
		ID:         userID + "_" + time.Now().Format("20060102150405"),
		UserID:     userID,
		DocumentID: documentID,
		Conn:       conn,
		Hub:        h.hub,
		Send:       make(chan []byte, 256),
	}

	client.Hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

func (h *WebSocketHandler) GetActiveUsers(c *gin.Context) {
	documentID := c.Param("id")
	users := h.hub.GetActiveUsers(documentID)
	c.JSON(http.StatusOK, gin.H{
		"document_id":  documentID,
		"active_users": users,
		"count":        len(users),
	})
}
