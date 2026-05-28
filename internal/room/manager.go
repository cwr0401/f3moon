package room

import (
	"fmt"
	"sync"

	"github.com/cwr0401/f3moon/internal/model"
)

// Manager 房间管理器
type Manager struct {
	mu     sync.RWMutex
	rooms  map[string]*Room
}

// NewManager 创建房间管理器
func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

// CreateRoom 创建房间
func (m *Manager) CreateRoom(name string, mode model.GameMode, ownerID string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("room_%d", len(m.rooms)+1)
	room := NewRoom(id, name, mode, ownerID)
	m.rooms[id] = room
	return room
}

// GetRoom 获取房间
func (m *Manager) GetRoom(id string) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rooms[id]
}

// ListRooms 列出所有房间
func (m *Manager) ListRooms() []*Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rooms := make([]*Room, 0, len(m.rooms))
	for _, room := range m.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// RemoveRoom 删除房间
func (m *Manager) RemoveRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

// JoinRoom 加入房间
func (m *Manager) JoinRoom(roomID string, player *RoomPlayer) (*Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}
	if !room.AddPlayer(player) {
		return nil, fmt.Errorf("room is full")
	}
	return room, nil
}

// LeaveRoom 离开房间
func (m *Manager) LeaveRoom(roomID, playerID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}
	room.RemovePlayer(playerID)
	return nil
}
