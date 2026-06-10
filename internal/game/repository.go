package game

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Uint8Slice 是一个可以存储到数据库的 []uint8 类型
type Uint8Slice []uint8

// Value 实现 driver.Valuer 接口
func (u Uint8Slice) Value() (driver.Value, error) {
	if u == nil {
		return nil, nil
	}
	return json.Marshal(u)
}

// Scan 实现 sql.Scanner 接口
func (u *Uint8Slice) Scan(value interface{}) error {
	if value == nil {
		*u = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Uint8Slice: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, u)
}

// GameDeckRecord 游戏牌栈持久化模型
type GameDeckRecord struct {
	ID             string    `gorm:"primaryKey;size:36" json:"id"`
	RoomID         string    `gorm:"size:36;not null;index:idx_room_id" json:"room_id"`
	Seed           int64     `gorm:"not null;default:0" json:"seed"`
	ShuffleStack   Uint8Slice `gorm:"type:json" json:"shuffle_stack"` // 洗牌后的栈
	CutStack       Uint8Slice `gorm:"type:json" json:"cut_stack"`     // 切牌后的栈
	StackTop       int       `gorm:"not null;default:0" json:"stack_top"`
	StackBottom    int       `gorm:"not null;default:111" json:"stack_bottom"`
	Shuffled       bool      `gorm:"not null;default:false" json:"shuffled"`
	CutPosition    int       `gorm:"not null;default:0" json:"cut_position"`
	CutFinished    bool      `gorm:"not null;default:false" json:"cut_finished"`
	DealerID       string    `gorm:"size:36" json:"dealer_id"`
	Player1ID      string    `gorm:"size:36" json:"player1_id"`
	Player2ID      string    `gorm:"size:36" json:"player2_id"`
	DealerHand     Uint8Slice `gorm:"type:json" json:"dealer_hand"`
	Player1Hand    Uint8Slice `gorm:"type:json" json:"player1_hand"`
	Player2Hand    Uint8Slice `gorm:"type:json" json:"player2_hand"`
	DealerFinished bool      `gorm:"not null;default:false" json:"dealer_finished"`
	DealFinished   bool      `gorm:"not null;default:false" json:"deal_finished"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 表名
func (GameDeckRecord) TableName() string { return "game_decks" }

// Repository 游戏存储接口
type Repository interface {
	CreateGameDeck(deck *GameDeckRecord) error
	GetGameDeck(id string) (*GameDeckRecord, error)
	GetGameDeckByRoomID(roomID string) (*GameDeckRecord, error)
	UpdateGameDeck(deck *GameDeckRecord) error
}

// gormRepository GORM 实现的游戏存储
type gormRepository struct {
	db *gorm.DB
}

// NewGORMRepository 创建 GORM 游戏存储
func NewGORMRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateGameDeck(deck *GameDeckRecord) error {
	if err := r.db.Create(deck).Error; err != nil {
		return fmt.Errorf("create game deck: %w", err)
	}
	return nil
}

func (r *gormRepository) GetGameDeck(id string) (*GameDeckRecord, error) {
	var deck GameDeckRecord
	if err := r.db.First(&deck, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get game deck %s: %w", id, err)
	}
	return &deck, nil
}

func (r *gormRepository) GetGameDeckByRoomID(roomID string) (*GameDeckRecord, error) {
	var deck GameDeckRecord
	if err := r.db.First(&deck, "room_id = ?", roomID).Error; err != nil {
		return nil, fmt.Errorf("get game deck by room %s: %w", roomID, err)
	}
	return &deck, nil
}

func (r *gormRepository) UpdateGameDeck(deck *GameDeckRecord) error {
	result := r.db.Model(&GameDeckRecord{}).Where("id = ?", deck.ID).Updates(map[string]any{
		"shuffle_stack":   deck.ShuffleStack,
		"cut_stack":       deck.CutStack,
		"stack_top":       deck.StackTop,
		"stack_bottom":    deck.StackBottom,
		"shuffled":        deck.Shuffled,
		"cut_position":    deck.CutPosition,
		"cut_finished":    deck.CutFinished,
		"dealer_id":       deck.DealerID,
		"player1_id":      deck.Player1ID,
		"player2_id":      deck.Player2ID,
		"dealer_hand":     deck.DealerHand,
		"player1_hand":    deck.Player1Hand,
		"player2_hand":    deck.Player2Hand,
		"dealer_finished": deck.DealerFinished,
		"deal_finished":   deck.DealFinished,
		"updated_at":      deck.UpdatedAt,
	})
	if result.Error != nil {
		return fmt.Errorf("update game deck %s: %w", deck.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update game deck %s: not found", deck.ID)
	}
	return nil
}
