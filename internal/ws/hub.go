package ws

import (
	"sync"

	"github.com/cwr0401/f3moon/internal/game"
)

// Hub WebSocket连接管理器
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client // playerID -> Client
}

// NewHub 创建Hub
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.PlayerID] = client
}

// Unregister 注销客户端
func (h *Hub) Unregister(playerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if client, ok := h.clients[playerID]; ok {
		close(client.Send)
		delete(h.clients, playerID)
	}
}

// Broadcast 广播消息
func (h *Hub) Broadcast(msg game.NotifyMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 如果指定了目标玩家, 只发送给该玩家
	if msg.Player != "" {
		if client, ok := h.clients[msg.Player]; ok {
			select {
			case client.Send <- msg:
			default:
				// 发送失败, 跳过
			}
		}
		return
	}

	// 广播给所有客户端
	for _, client := range h.clients {
		select {
		case client.Send <- msg:
		default:
			// 发送失败, 跳过
		}
	}
}

// SendToPlayer 发送消息给指定玩家
func (h *Hub) SendToPlayer(playerID string, msg game.NotifyMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.clients[playerID]; ok {
		select {
		case client.Send <- msg:
		default:
		}
	}
}

// IsOnline 检查玩家是否在线
func (h *Hub) IsOnline(playerID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[playerID]
	return ok
}

// OnlineCount 在线玩家数
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
