package room

import (
	"fmt"
	"sync"
	"time"

	"github.com/cwr0401/f3moon/internal/model"
	"github.com/cwr0401/f3moon/internal/zone"
	"github.com/google/uuid"
)

// Repository 房间持久化接口
type Repository interface {
	CreateRoom(room *Room) error
	UpdateRoom(room *Room) error
	CloseRoom(roomID string, closedAt time.Time) error
	UpsertScore(roomID, playerID string, score int) error
	GetUserRoom(userID string) (string, error)
	SetUserRoom(userID, roomID string) error
	RemoveUserRoom(userID string) error
	IsRoomActive(roomID string) (bool, error)
	CountActiveRoomsByZone(zoneID string) (int, error)
}

// Manager 房间管理器
type Manager struct {
	mu          sync.RWMutex
	rooms       map[string]*Room
	repo        Repository
	zoneManager *zone.Manager
}

// NewManager 创建房间管理器
func NewManager(repo Repository, zoneManager *zone.Manager) *Manager {
	return &Manager{
		rooms:       make(map[string]*Room),
		repo:        repo,
		zoneManager: zoneManager,
	}
}

// checkUserNotInRoom 检查用户是否不在任何房间中，如果关联的房间已失效则自动清理
func (m *Manager) checkUserNotInRoom(userID string) error {
	currentRoomID, err := m.repo.GetUserRoom(userID)
	if err != nil {
		return fmt.Errorf("check user room: %w", err)
	}
	if currentRoomID == "" {
		return nil
	}
	// 先检查内存中是否存在该房间
	if _, ok := m.rooms[currentRoomID]; ok {
		return fmt.Errorf("you are already in room %s", currentRoomID)
	}
	// 内存中不存在，检查数据库中房间是否仍活跃
	active, err := m.repo.IsRoomActive(currentRoomID)
	if err != nil {
		return fmt.Errorf("check room active: %w", err)
	}
	if active {
		return fmt.Errorf("you are already in room %s", currentRoomID)
	}
	// 房间已失效（已关闭或不存在），清理过期关联
	_ = m.repo.RemoveUserRoom(userID)
	return nil
}

// CreateRoom 创建房间
func (m *Manager) CreateRoom(zoneID, name string, mode model.GameMode, ownerID string, maxRounds int) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查用户是否已经在其他房间
	if err := m.checkUserNotInRoom(ownerID); err != nil {
		return nil, err
	}

	// 检查游戏区是否可以创建房间
	if !m.zoneManager.CanCreateRoomInZone(zoneID) {
		return nil, fmt.Errorf("cannot create room in zone %s: zone is full", zoneID)
	}

	id := uuid.NewString()
	room := NewRoom(id, zoneID, name, mode, ownerID, maxRounds)
	if err := m.repo.CreateRoom(room); err != nil {
		return nil, fmt.Errorf("persist created room %s: %w", id, err)
	}

	// 增加游戏区房间计数
	if err := m.zoneManager.IncrementZoneRoomCount(zoneID); err != nil {
		// 回滚房间创建
		_ = m.repo.CloseRoom(id, time.Now().UTC())
		return nil, fmt.Errorf("failed to update zone room count: %w", err)
	}

	m.rooms[id] = room
	return room, nil
}

// CreateRoomAndJoin 创建房间并让房主直接加入
func (m *Manager) CreateRoomAndJoin(zoneID, name string, mode model.GameMode, owner *RoomPlayer, maxRounds int) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查用户是否已经在其他房间
	if err := m.checkUserNotInRoom(owner.ID); err != nil {
		return nil, err
	}

	// 检查游戏区是否可以创建房间
	if !m.zoneManager.CanCreateRoomInZone(zoneID) {
		return nil, fmt.Errorf("cannot create room in zone %s: zone is full", zoneID)
	}

	id := uuid.NewString()
	room := NewRoom(id, zoneID, name, mode, owner.ID, maxRounds)

	// 房主加入房间
	if !room.AddPlayer(owner) {
		return nil, fmt.Errorf("cannot add owner to room")
	}

	if err := m.repo.CreateRoom(room); err != nil {
		return nil, fmt.Errorf("persist created room %s: %w", id, err)
	}

	// 设置房主的房间关联
	if err := m.repo.SetUserRoom(owner.ID, id); err != nil {
		_ = m.repo.CloseRoom(id, time.Now().UTC())
		return nil, fmt.Errorf("set owner room: %w", err)
	}

	// 增加游戏区房间计数
	if err := m.zoneManager.IncrementZoneRoomCount(zoneID); err != nil {
		_ = m.repo.RemoveUserRoom(owner.ID)
		_ = m.repo.CloseRoom(id, time.Now().UTC())
		return nil, fmt.Errorf("failed to update zone room count: %w", err)
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

// ListRoomsByZone 列出指定游戏区的房间
func (m *Manager) ListRoomsByZone(zoneID string) []*Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rooms := make([]*Room, 0)
	for _, room := range m.rooms {
		if room.ZoneID == zoneID {
			rooms = append(rooms, room)
		}
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

	// 检查用户是否已经在其他房间
	if err := m.checkUserNotInRoom(player.ID); err != nil {
		return nil, err
	}

	room, ok := m.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}
	if !room.AddPlayer(player) {
		return nil, fmt.Errorf("room is full")
	}

	// 更新用户房间关联
	if err := m.repo.SetUserRoom(player.ID, roomID); err != nil {
		// 回滚添加玩家操作
		room.RemovePlayer(player.ID)
		return nil, fmt.Errorf("set user room: %w", err)
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

	// 检查是否有正在进行的游戏
	room.mu.RLock()
	if room.Status == RoomPlaying {
		room.mu.RUnlock()
		return fmt.Errorf("cannot leave room while game is playing")
	}
	isOwner := room.Owner == playerID
	room.mu.RUnlock()

	// 如果是房主，不能直接离开，需要关闭房间
	if isOwner {
		return fmt.Errorf("room owner cannot leave, use close room instead")
	}

	room.RemovePlayer(playerID)

	// 移除用户房间关联
	if err := m.repo.RemoveUserRoom(playerID); err != nil {
		return fmt.Errorf("remove user room: %w", err)
	}

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
func (m *Manager) CloseRoom(roomID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	// 检查是否是房主
	room.mu.RLock()
	if room.Owner != userID {
		room.mu.RUnlock()
		return fmt.Errorf("only room owner can close room")
	}
	// 检查是否有正在进行的游戏
	if room.Status == RoomPlaying {
		room.mu.RUnlock()
		return fmt.Errorf("cannot close room while game is playing")
	}
	// 收集所有玩家ID以便清除关联
	var playerIDs []string
	for _, p := range room.Players {
		if p != nil {
			playerIDs = append(playerIDs, p.ID)
		}
	}
	room.mu.RUnlock()

	closedAt := room.Close()
	if err := m.repo.CloseRoom(roomID, closedAt); err != nil {
		return fmt.Errorf("persist closed room %s: %w", roomID, err)
	}

	// 清除所有玩家的房间关联
	for _, pid := range playerIDs {
		_ = m.repo.RemoveUserRoom(pid)
	}

	// 减少游戏区房间计数
	if room.ZoneID != "" {
		_ = m.zoneManager.DecrementZoneRoomCount(room.ZoneID)
	}

	// 从内存中移除房间
	delete(m.rooms, roomID)

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
