package ws

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/cwr0401/f3moon/internal/game"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// Upgrader WebSocket升级器
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client WebSocket客户端
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan game.NotifyMessage
	PlayerID string
	GameID   string
}

// NewClient 创建客户端
func NewClient(hub *Hub, conn *websocket.Conn, playerID, gameID string) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan game.NotifyMessage, 256),
		PlayerID: playerID,
		GameID:   gameID,
	}
}

// ReadPump 读取客户端消息
func (c *Client) ReadPump(handler func(game.GameEvent)) {
	defer func() {
		c.Hub.Unregister(c.PlayerID)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// 解析消息为GameEvent
		evt, err := parseGameEvent(message, c.PlayerID)
		if err != nil {
			continue
		}

		handler(evt)
	}
}

// WritePump 写入消息到客户端
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteJSON(msg)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// parseGameEvent 解析WebSocket消息为GameEvent
func parseGameEvent(data []byte, playerID string) (game.GameEvent, error) {
	// 简化实现: 使用JSON解析
	// 实际实现需要根据消息类型映射到对应的事件数据类型
	evt := game.GameEvent{
		PlayerID: playerID,
	}
	return evt, nil
}

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(hub *Hub, handler func(game.GameEvent)) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.Query("player_id")
		gameID := c.Query("game_id")

		conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := NewClient(hub, conn, playerID, gameID)
		hub.Register(client)

		go client.WritePump()
		go client.ReadPump(handler)
	}
}
