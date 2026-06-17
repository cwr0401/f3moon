package room

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// RoomRecord 房间持久化模型
type RoomRecord struct {
	ID           string     `gorm:"primaryKey;size:36" json:"id"`
	ZoneID       string     `gorm:"size:36;index" json:"zone_id"`
	Name         string     `gorm:"size:128;not null" json:"name"`
	Mode         int        `gorm:"not null" json:"mode"`
	Status       int        `gorm:"not null" json:"status"`
	Owner        string     `gorm:"size:36;not null" json:"owner"`
	MaxPlayers   int        `gorm:"not null" json:"max_players"`
	MaxRounds    int        `gorm:"not null;default:8" json:"max_rounds"`
	CurrentRound int        `gorm:"not null;default:0" json:"current_round"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
}

// TableName 表名
func (RoomRecord) TableName() string { return "rooms" }

// RoomScore 房间积分记录
type RoomScore struct {
	RoomID    string    `gorm:"primaryKey;size:36" json:"room_id"`
	PlayerID  string    `gorm:"primaryKey;size:36" json:"player_id"`
	Score     int       `gorm:"not null;default:0" json:"score"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 表名
func (RoomScore) TableName() string { return "room_scores" }

// UserRoomRecord 用户房间关联记录 - 确保一个用户只能在一个房间中
type UserRoomRecord struct {
	UserID    string    `gorm:"primaryKey;size:36" json:"user_id"`
	RoomID    string    `gorm:"size:36;index;not null" json:"room_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 表名
func (UserRoomRecord) TableName() string { return "user_rooms" }

// gormRepository GORM 实现的房间存储
type gormRepository struct {
	db *gorm.DB
}

// NewGORMRepository 创建 GORM 房间存储
func NewGORMRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateRoom(room *Room) error {
	record := toRoomRecord(room)
	if err := r.db.Create(record).Error; err != nil {
		return fmt.Errorf("create room: %w", err)
	}
	return nil
}

func (r *gormRepository) UpdateRoom(room *Room) error {
	record := toRoomRecord(room)
	result := r.db.Model(&RoomRecord{}).Where("id = ?", record.ID).Updates(map[string]any{
		"name":          record.Name,
		"mode":          record.Mode,
		"status":        record.Status,
		"owner":         record.Owner,
		"max_players":   record.MaxPlayers,
		"max_rounds":    record.MaxRounds,
		"current_round": record.CurrentRound,
		"updated_at":    record.UpdatedAt,
	})
	if result.Error != nil {
		return fmt.Errorf("update room %s: %w", record.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update room %s: not found", record.ID)
	}
	return nil
}

func (r *gormRepository) CloseRoom(roomID string, closedAt time.Time) error {
	result := r.db.Model(&RoomRecord{}).Where("id = ?", roomID).Updates(map[string]any{
		"status":     int(RoomFinished),
		"closed_at":  closedAt,
		"updated_at": closedAt,
	})
	if result.Error != nil {
		return fmt.Errorf("close room %s: %w", roomID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("close room %s: not found", roomID)
	}
	return nil
}

func (r *gormRepository) UpsertScore(roomID, playerID string, score int) error {
	now := time.Now().UTC()
	result := r.db.Where("room_id = ? AND player_id = ?", roomID, playerID).
		Assign(map[string]any{
			"score":      score,
			"updated_at": now,
		}).
		FirstOrCreate(&RoomScore{
			RoomID:    roomID,
			PlayerID:  playerID,
			Score:     score,
			CreatedAt: now,
		})
	if result.Error != nil {
		return fmt.Errorf("upsert score: %w", result.Error)
	}
	return nil
}

func (r *gormRepository) GetUserRoom(userID string) (string, error) {
	var record UserRoomRecord
	result := r.db.Where("user_id = ?", userID).First(&record)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", fmt.Errorf("get user room: %w", result.Error)
	}
	return record.RoomID, nil
}

func (r *gormRepository) SetUserRoom(userID, roomID string) error {
	now := time.Now().UTC()
	result := r.db.Where("user_id = ?", userID).
		Assign(map[string]any{
			"room_id":    roomID,
			"updated_at": now,
		}).
		FirstOrCreate(&UserRoomRecord{
			UserID:    userID,
			RoomID:    roomID,
			CreatedAt: now,
			UpdatedAt: now,
		})
	if result.Error != nil {
		return fmt.Errorf("set user room: %w", result.Error)
	}
	return nil
}

func (r *gormRepository) RemoveUserRoom(userID string) error {
	result := r.db.Where("user_id = ?", userID).Delete(&UserRoomRecord{})
	if result.Error != nil {
		return fmt.Errorf("remove user room: %w", result.Error)
	}
	return nil
}

func (r *gormRepository) IsRoomActive(roomID string) (bool, error) {
	var count int64
	result := r.db.Model(&RoomRecord{}).Where("id = ? AND status != ?", roomID, int(RoomFinished)).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("check room active: %w", result.Error)
	}
	return count > 0, nil
}

func (r *gormRepository) CountActiveRoomsByZone(zoneID string) (int, error) {
	var count int64
	result := r.db.Model(&RoomRecord{}).Where("zone_id = ? AND status != ?", zoneID, int(RoomFinished)).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("count active rooms by zone %s: %w", zoneID, result.Error)
	}
	return int(count), nil
}

// toRoomRecord 将内存 Room 转为持久化记录
func toRoomRecord(r *Room) *RoomRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &RoomRecord{
		ID:           r.ID,
		ZoneID:       r.ZoneID,
		Name:         r.Name,
		Mode:         int(r.Mode),
		Status:       int(r.Status),
		Owner:        r.Owner,
		MaxPlayers:   r.MaxPlayers,
		MaxRounds:    r.MaxRounds,
		CurrentRound: r.CurrentRound,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
		ClosedAt:     r.ClosedAt,
	}
}
