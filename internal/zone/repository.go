package zone

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// GameZoneRecord 游戏区持久化模型
type GameZoneRecord struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Name        string    `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:512" json:"description"`
	MaxRooms    int       `gorm:"not null" json:"max_rooms"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 表名
func (GameZoneRecord) TableName() string { return "game_zones" }

// Repository 游戏区持久化接口
type Repository interface {
	CreateZone(zone *GameZone) error
	UpdateZone(zone *GameZone) error
	DeleteZone(zoneID string) error
	GetZone(zoneID string) (*GameZone, error)
	GetZoneByName(name string) (*GameZone, error)
	ListZones() ([]*GameZone, error)
}

// gormRepository GORM 实现的游戏区存储
type gormRepository struct {
	db *gorm.DB
}

// NewGORMRepository 创建 GORM 游戏区存储
func NewGORMRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateZone(zone *GameZone) error {
	record := toZoneRecord(zone)
	if err := r.db.Create(record).Error; err != nil {
		return fmt.Errorf("create zone: %w", err)
	}
	return nil
}

func (r *gormRepository) UpdateZone(zone *GameZone) error {
	record := toZoneRecord(zone)
	result := r.db.Model(&GameZoneRecord{}).Where("id = ?", record.ID).Updates(map[string]any{
		"name":        record.Name,
		"description": record.Description,
		"max_rooms":   record.MaxRooms,
		"updated_at":  record.UpdatedAt,
	})
	if result.Error != nil {
		return fmt.Errorf("update zone %s: %w", record.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update zone %s: not found", record.ID)
	}
	return nil
}

func (r *gormRepository) DeleteZone(zoneID string) error {
	result := r.db.Delete(&GameZoneRecord{}, "id = ?", zoneID)
	if result.Error != nil {
		return fmt.Errorf("delete zone %s: %w", zoneID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete zone %s: not found", zoneID)
	}
	return nil
}

func (r *gormRepository) GetZone(zoneID string) (*GameZone, error) {
	var record GameZoneRecord
	if err := r.db.Where("id = ?", zoneID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get zone %s: %w", zoneID, err)
	}
	return fromZoneRecord(&record), nil
}

func (r *gormRepository) GetZoneByName(name string) (*GameZone, error) {
	var record GameZoneRecord
	if err := r.db.Where("name = ?", name).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get zone by name %s: %w", name, err)
	}
	return fromZoneRecord(&record), nil
}

func (r *gormRepository) ListZones() ([]*GameZone, error) {
	var records []GameZoneRecord
	if err := r.db.Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list zones: %w", err)
	}

	zones := make([]*GameZone, len(records))
	for i := range records {
		zones[i] = fromZoneRecord(&records[i])
	}
	return zones, nil
}

// toZoneRecord 将内存 GameZone 转为持久化记录
func toZoneRecord(z *GameZone) *GameZoneRecord {
	z.mu.RLock()
	defer z.mu.RUnlock()

	return &GameZoneRecord{
		ID:          z.ID,
		Name:        z.Name,
		Description: z.Description,
		MaxRooms:    z.MaxRooms,
		CreatedAt:   z.CreatedAt,
		UpdatedAt:   z.UpdatedAt,
	}
}

// fromZoneRecord 将持久化记录转为内存 GameZone
func fromZoneRecord(record *GameZoneRecord) *GameZone {
	return &GameZone{
		ID:          record.ID,
		Name:        record.Name,
		Description: record.Description,
		MaxRooms:    record.MaxRooms,
		RoomCount:   0, // 这个值在 Manager 中动态维护
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}
