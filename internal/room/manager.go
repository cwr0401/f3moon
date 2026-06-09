package room

import (
	"fmt"
	"sync"
	"time"

	"github.com/cwr0401/f3moon/internal/model"
	"github.com/google/uuid"
)

// Repository 房间持久化接口
type Repository interface {
	CreateRoom(room *Room) error
	UpdateRoom(room *Room) error
	CloseRoom(roomID string, closedAt time.Time) error
	UpsertScore(roomID, playerID string, score int) error
}

// Manager 房间管理器
type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
	repo  Repository
}

// NewManager 创建房间管理器
func NewManager(repo Repository) *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
		repo:  repo,
	}
}

// CreateRoom 创建房间
func (m *Manager) CreateRoom(name string, mode model.GameMode, ownerID string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.NewString()
	room := NewRoom(id, name, mode, ownerID)
	if err := m.repo.CreateRoom(room); err != nil {
		return nil, fmt.Errorf("persist created room %s: %w", id, err)
	}

	m.rooms[id] = room
	return room, nil
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
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}
	if !room.AddPlayer(player) {
		return nil, fmt.Errorf("room is full")
	}

	if err := m.repo.UpdateRoom(room); err != nil {
		return nil, fmt.Errorf("persist joined room %s: %w", roomID, err)
	}
	return room, nil
}

// LeaveRoom 离开房间
func (m *Manager) LeaveRoom(roomID, playerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}
	room.RemovePlayer(playerID)

	if err := m.repo.UpdateRoom(room); err != nil {
		return fmt.Errorf("persist left room %s: %w", roomID, err)
	}
	return nil
}

// SetReady 设置玩家准备状态
func (m *Manager) SetReady(roomID, playerID string, ready bool) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}
	room.SetReady(playerID, ready)

	if err := m.repo.UpdateRoom(room); err != nil {
		return nil, fmt.Errorf("persist ready room %s: %w", roomID, err)
	}
	return room, nil
}

// AddAIPlayer 添加AI玩家并持久化房间更新时间
func (m *Manager) AddAIPlayer(roomID string) (*RoomPlayer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}
	player := room.AddAIPlayer()
	if player == nil {
		return nil, fmt.Errorf("room is full")
	}

	if err := m.repo.UpdateRoom(room); err != nil {
		return nil, fmt.Errorf("persist ai room %s: %w", roomID, err)
	}
	return player, nil
}

// StartRoomGame 校验并将房间切换为游戏中状态
func (m *Manager) StartRoomGame(roomID, ownerID string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}
	if room.Owner != ownerID {
		return nil, fmt.Errorf("only room owner can start game")
	}
	if !room.IsFull() {
		return nil, fmt.Errorf("room is not full")
	}
	if !room.AllReady() {
		return nil, fmt.Errorf("not all players ready")
	}

	room.SetStatus(RoomPlaying)
	if err := m.repo.UpdateRoom(room); err != nil {
		return nil, fmt.Errorf("persist started room %s: %w", roomID, err)
	}
	return room, nil
}

// CloseRoom 关闭房间并记录关闭时间
func (m *Manager) CloseRoom(roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	closedAt := room.Close()
	if err := m.repo.CloseRoom(roomID, closedAt); err != nil {
		return fmt.Errorf("persist closed room %s: %w", roomID, err)
	}
	return nil
}

// UpdateScore 更新玩家积分
func (m *Manager) UpdateScore(roomID, playerID string, score int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	room.SetScore(playerID, score)
	if err := m.repo.UpsertScore(roomID, playerID, score); err != nil {
		return fmt.Errorf("persist room %s score for player %s: %w", roomID, playerID, err)
	}
	return nil
}
