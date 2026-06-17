package zone

import (
	"sync"
	"time"
)

// GameZone 游戏区
type GameZone struct {
	mu          sync.RWMutex
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MaxRooms    int       `json:"max_rooms"`
	RoomCount   int       `json:"room_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewGameZone 创建游戏区
func NewGameZone(id, name, description string, maxRooms int) *GameZone {
	now := time.Now().UTC()
	return &GameZone{
		ID:          id,
		Name:        name,
		Description: description,
		MaxRooms:    maxRooms,
		RoomCount:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CanCreateRoom 检查是否可以在该游戏区创建房间
func (z *GameZone) CanCreateRoom() bool {
	z.mu.RLock()
	defer z.mu.RUnlock()
	return z.RoomCount < z.MaxRooms
}

// IncrementRoomCount 增加房间计数
func (z *GameZone) IncrementRoomCount() {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.RoomCount++
	z.UpdatedAt = time.Now().UTC()
}

// DecrementRoomCount 减少房间计数
func (z *GameZone) DecrementRoomCount() {
	z.mu.Lock()
	defer z.mu.Unlock()
	if z.RoomCount > 0 {
		z.RoomCount--
	}
	z.UpdatedAt = time.Now().UTC()
}

// Update 更新游戏区信息
func (z *GameZone) Update(name, description string, maxRooms int) {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.Name = name
	z.Description = description
	z.MaxRooms = maxRooms
	z.UpdatedAt = time.Now().UTC()
}

// GetRoomCount 获取当前房间数
func (z *GameZone) GetRoomCount() int {
	z.mu.RLock()
	defer z.mu.RUnlock()
	return z.RoomCount
}
