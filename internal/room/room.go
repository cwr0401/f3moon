package room

import (
	"sync"

	"github.com/cwr0401/f3moon/internal/model"
)

// RoomStatus 房间状态
type RoomStatus int

const (
	RoomWaiting RoomStatus = iota
	RoomPlaying
	RoomFinished
)

// Room 房间
type Room struct {
	mu       sync.RWMutex
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Mode     model.GameMode     `json:"mode"`
	Status   RoomStatus         `json:"status"`
	Players  [4]*RoomPlayer     `json:"players"`
	Owner    string             `json:"owner"`
	MaxPlayers int              `json:"max_players"`
}

// RoomPlayer 房间中的玩家
type RoomPlayer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	IsAI   bool   `json:"is_ai"`
	Seat   int    `json:"seat"` // 座位号 0-3
	Ready  bool   `json:"ready"`
}

// NewRoom 创建房间
func NewRoom(id, name string, mode model.GameMode, ownerID string) *Room {
	maxPlayers := 4
	if mode == model.GameMode3Player {
		maxPlayers = 3
	}
	return &Room{
		ID:         id,
		Name:       name,
		Mode:       mode,
		Status:     RoomWaiting,
		Owner:      ownerID,
		MaxPlayers: maxPlayers,
	}
}

// AddPlayer 添加玩家
func (r *Room) AddPlayer(player *RoomPlayer) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != RoomWaiting {
		return false
	}

	for i := 0; i < r.MaxPlayers; i++ {
		if r.Players[i] == nil {
			player.Seat = i
			r.Players[i] = player
			return true
		}
	}
	return false
}

// RemovePlayer 移除玩家
func (r *Room) RemovePlayer(playerID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.MaxPlayers; i++ {
		if r.Players[i] != nil && r.Players[i].ID == playerID {
			r.Players[i] = nil
			return true
		}
	}
	return false
}

// PlayerCount 当前玩家数量
func (r *Room) PlayerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for i := 0; i < r.MaxPlayers; i++ {
		if r.Players[i] != nil {
			count++
		}
	}
	return count
}

// IsFull 房间是否已满
func (r *Room) IsFull() bool {
	return r.PlayerCount() >= r.MaxPlayers
}

// AllReady 所有玩家是否就绪
func (r *Room) AllReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i := 0; i < r.MaxPlayers; i++ {
		if r.Players[i] == nil || !r.Players[i].Ready {
			return false
		}
	}
	return true
}

// SetReady 设置玩家就绪状态
func (r *Room) SetReady(playerID string, ready bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.MaxPlayers; i++ {
		if r.Players[i] != nil && r.Players[i].ID == playerID {
			r.Players[i].Ready = ready
			return
		}
	}
}

// AddAIPlayer 添加AI玩家
func (r *Room) AddAIPlayer() *RoomPlayer {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.MaxPlayers; i++ {
		if r.Players[i] == nil {
			player := &RoomPlayer{
				ID:    "ai_" + r.ID + "_" + string(rune('A'+i)),
				Name:  "AI-" + string(rune('A'+i)),
				IsAI:  true,
				Seat:  i,
				Ready: true,
			}
			r.Players[i] = player
			return player
		}
	}
	return nil
}

// ToModelPlayers 将房间玩家转换为模型玩家
func (r *Room) ToModelPlayers() [4]*model.Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var players [4]*model.Player
	roles := []model.PlayerRole{
		model.RoleDealer,
		model.RoleIdle1,
		model.RoleIdle2,
		model.RoleRest,
	}

	for i := 0; i < 4; i++ {
		if r.Players[i] != nil {
			players[i] = model.NewPlayer(
				r.Players[i].ID,
				r.Players[i].Name,
				roles[i],
				r.Players[i].IsAI,
			)
		}
	}

	return players
}
