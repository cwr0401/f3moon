package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/game"
	"github.com/cwr0401/f3moon/internal/ws"
)

// WSHandler WebSocket接口处理器
type WSHandler struct {
	hub *ws.Hub
}

// NewWSHandler 创建WebSocket处理器
func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// HandleWebSocket 处理WebSocket连接
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	playerID := c.Query("player_id")
	gameID := c.Query("game_id")

	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(h.hub, conn, playerID, gameID)
	h.hub.Register(client)

	// 事件处理器: 将WebSocket消息转发给游戏状态机
	handler := func(evt game.GameEvent) {
		// TODO: 根据gameID找到对应的StateMachine并处理事件
	}

	go client.WritePump()
	go client.ReadPump(handler)
}
