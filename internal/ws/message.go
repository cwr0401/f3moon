package ws

import "github.com/cwr0401/f3moon/internal/game"

// WSMessage WebSocket消息格式
type WSMessage struct {
	Type    string      `json:"type"`
	GameID  string      `json:"game_id"`
	Player  string      `json:"player"`
	Data    interface{} `json:"data"`
	Seq     int64       `json:"seq"`
}

// ToNotifyMessage 转换为NotifyMessage
func (m *WSMessage) ToNotifyMessage() game.NotifyMessage {
	return game.NotifyMessage{
		Type:   game.NotifyType(m.Type),
		GameID: m.GameID,
		Player: m.Player,
		Data:   m.Data,
	}
}

// FromNotifyMessage 从NotifyMessage创建WSMessage
func FromNotifyMessage(msg game.NotifyMessage, seq int64) WSMessage {
	return WSMessage{
		Type:   string(msg.Type),
		GameID: msg.GameID,
		Player: msg.Player,
		Data:   msg.Data,
		Seq:    seq,
	}
}
