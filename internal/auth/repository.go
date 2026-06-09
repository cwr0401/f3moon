package auth

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ErrUserNotFound 用户不存在
var ErrUserNotFound = errors.New("user not found")

// ErrTokenNotFound 验证 token 不存在
var ErrTokenNotFound = errors.New("token not found")

// Repository 用户与验证 token 存储接口
type Repository interface {
	CreateUser(user *User) error
	FindUserByEmail(email string) (*User, error)
	FindUserByID(id string) (*User, error)
	MarkUserVerified(id string) error

	CreateToken(token *VerificationToken) error
	FindToken(token string) (*VerificationToken, error)
	// ClaimToken 原子地标记 token 为已使用（WHERE used_at IS NULL），返回受影响行数
	ClaimToken(token string, usedAt time.Time) (int64, error)
}

// gormRepository GORM 实现
type gormRepository struct {
	db *gorm.DB
}

// NewGORMRepository 创建 GORM repository
func NewGORMRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateUser(user *User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *gormRepository) FindUserByEmail(email string) (*User, error) {
	var u User
	err := r.db.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

func (r *gormRepository) FindUserByID(id string) (*User, error) {
	var u User
	err := r.db.Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

func (r *gormRepository) MarkUserVerified(id string) error {
	res := r.db.Model(&User{}).Where("id = ?", id).Update("verified", true)
	if res.Error != nil {
		return fmt.Errorf("mark user verified: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *gormRepository) CreateToken(token *VerificationToken) error {
	if err := r.db.Create(token).Error; err != nil {
		return fmt.Errorf("create token: %w", err)
	}
	return nil
}

func (r *gormRepository) FindToken(token string) (*VerificationToken, error) {
	var t VerificationToken
	err := r.db.Where("token = ?", token).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find token: %w", err)
	}
	return &t, nil
}

func (r *gormRepository) ClaimToken(token string, usedAt time.Time) (int64, error) {
	res := r.db.Model(&VerificationToken{}).
		Where("token = ? AND used_at IS NULL", token).
		Update("used_at", usedAt)
	if res.Error != nil {
		return 0, fmt.Errorf("claim token: %w", res.Error)
	}
	return res.RowsAffected, nil
}
