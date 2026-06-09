package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/auth"
	"github.com/cwr0401/f3moon/internal/game"
	"github.com/cwr0401/f3moon/internal/ws"
)

// WSHandler WebSocket接口处理器
type WSHandler struct {
	hub    *ws.Hub
	jwtMgr *auth.JWTManager
}

// NewWSHandler 创建WebSocket处理器
func NewWSHandler(hub *ws.Hub, jwtMgr *auth.JWTManager) *WSHandler {
	return &WSHandler{hub: hub, jwtMgr: jwtMgr}
}

// HandleWebSocket 处理WebSocket连接
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	claims, err := h.jwtMgr.Parse(tokenStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	gameID := c.Query("game_id")

	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(h.hub, conn, claims.UserID, gameID)
	h.hub.Register(client)

	// 事件处理器: 将WebSocket消息转发给游戏状态机
	handler := func(evt game.GameEvent) {
		// TODO: 根据gameID找到对应的StateMachine并处理事件
	}

	go client.WritePump()
	go client.ReadPump(handler)
}
