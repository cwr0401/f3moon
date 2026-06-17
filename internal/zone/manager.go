package zone

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Manager 游戏区管理器
type Manager struct {
	mu    sync.RWMutex
	zones map[string]*GameZone
	repo  Repository
}

// NewManager 创建游戏区管理器
func NewManager(repo Repository) *Manager {
	return &Manager{
		zones: make(map[string]*GameZone),
		repo:  repo,
	}
}

// LoadFromDB 从数据库加载所有游戏区
func (m *Manager) LoadFromDB() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zones, err := m.repo.ListZones()
	if err != nil {
		return fmt.Errorf("load zones from db: %w", err)
	}

	for _, zone := range zones {
		m.zones[zone.ID] = zone
	}
	return nil
}

// RebuildRoomCounts 根据外部计数函数重建每个游戏区的活跃房间数。
// 用于服务启动时同步内存计数与数据库实际状态，避免重启后 RoomCount 全部为 0
// 而被错误的容量限制阻塞或允许超额创建。
func (m *Manager) RebuildRoomCounts(counter func(zoneID string) (int, error)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, z := range m.zones {
		count, err := counter(z.ID)
		if err != nil {
			return fmt.Errorf("count active rooms for zone %s: %w", z.ID, err)
		}
		z.mu.Lock()
		z.RoomCount = count
		z.mu.Unlock()
	}
	return nil
}

// CreateZone 创建游戏区
func (m *Manager) CreateZone(name, description string, maxRooms int) (*GameZone, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查名称是否已存在
	existing, err := m.repo.GetZoneByName(name)
	if err != nil {
		return nil, fmt.Errorf("check zone name: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("zone name already exists: %s", name)
	}

	id := uuid.NewString()
	zone := NewGameZone(id, name, description, maxRooms)

	if err := m.repo.CreateZone(zone); err != nil {
		return nil, fmt.Errorf("persist zone: %w", err)
	}

	m.zones[id] = zone
	return zone, nil
}

// GetZone 获取游戏区
func (m *Manager) GetZone(id string) *GameZone {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.zones[id]
}

// ListZones 列出所有游戏区
func (m *Manager) ListZones() []*GameZone {
	m.mu.RLock()
	defer m.mu.RUnlock()

	zones := make([]*GameZone, 0, len(m.zones))
	for _, zone := range m.zones {
		zones = append(zones, zone)
	}
	return zones
}

// UpdateZone 更新游戏区
func (m *Manager) UpdateZone(id, name, description string, maxRooms int) (*GameZone, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	zone, ok := m.zones[id]
	if !ok {
		return nil, fmt.Errorf("zone not found: %s", id)
	}

	// 如果改了名称，检查新名称是否已被占用
	if name != zone.Name {
		existing, err := m.repo.GetZoneByName(name)
		if err != nil {
			return nil, fmt.Errorf("check zone name: %w", err)
		}
		if existing != nil {
			return nil, fmt.Errorf("zone name already exists: %s", name)
		}
	}

	zone.Update(name, description, maxRooms)

	if err := m.repo.UpdateZone(zone); err != nil {
		return nil, fmt.Errorf("persist updated zone: %w", err)
	}

	return zone, nil
}

// DeleteZone 删除游戏区
func (m *Manager) DeleteZone(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zone, ok := m.zones[id]
	if !ok {
		return fmt.Errorf("zone not found: %s", id)
	}

	if zone.GetRoomCount() > 0 {
		return fmt.Errorf("cannot delete zone with active rooms")
	}

	if err := m.repo.DeleteZone(id); err != nil {
		return fmt.Errorf("delete zone from db: %w", err)
	}

	delete(m.zones, id)
	return nil
}

// CanCreateRoomInZone 检查是否可以在指定游戏区创建房间
func (m *Manager) CanCreateRoomInZone(zoneID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	zone, ok := m.zones[zoneID]
	if !ok {
		return false
	}
	return zone.CanCreateRoom()
}

// IncrementZoneRoomCount 增加游戏区的房间计数
func (m *Manager) IncrementZoneRoomCount(zoneID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zone, ok := m.zones[zoneID]
	if !ok {
		return fmt.Errorf("zone not found: %s", zoneID)
	}

	zone.IncrementRoomCount()
	if err := m.repo.UpdateZone(zone); err != nil {
		zone.DecrementRoomCount() // 回滚
		return fmt.Errorf("persist zone room count: %w", err)
	}

	return nil
}

// DecrementZoneRoomCount 减少游戏区的房间计数
func (m *Manager) DecrementZoneRoomCount(zoneID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zone, ok := m.zones[zoneID]
	if !ok {
		return fmt.Errorf("zone not found: %s", zoneID)
	}

	zone.DecrementRoomCount()
	if err := m.repo.UpdateZone(zone); err != nil {
		zone.IncrementRoomCount() // 回滚
		return fmt.Errorf("persist zone room count: %w", err)
	}

	return nil
}

// InitializeDefaultZones 初始化默认游戏区
func (m *Manager) InitializeDefaultZones() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	defaultZones := []struct {
		name        string
		description string
		maxRooms    int
	}{
		{"新手区", "适合新手玩家的游戏区", 10},
		{"进阶区", "有一定经验的玩家", 20},
		{"高手区", "高手玩家专区", 15},
	}

	for _, dz := range defaultZones {
		existing, err := m.repo.GetZoneByName(dz.name)
		if err != nil {
			return fmt.Errorf("check default zone %s: %w", dz.name, err)
		}
		if existing != nil {
			continue // 已存在，跳过
		}

		id := uuid.NewString()
		zone := NewGameZone(id, dz.name, dz.description, dz.maxRooms)
		if err := m.repo.CreateZone(zone); err != nil {
			return fmt.Errorf("create default zone %s: %w", dz.name, err)
		}
		m.zones[id] = zone
	}

	return nil
}
