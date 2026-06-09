package auth

import "time"

// User 用户
type User struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Nickname     string    `gorm:"size:64" json:"nickname"`
	Verified     bool      `gorm:"default:false;not null" json:"verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 表名
func (User) TableName() string { return "users" }

// TokenPurpose 验证 token 用途
type TokenPurpose string

const (
	PurposeVerifyEmail TokenPurpose = "verify_email"
)

// VerificationToken 邮件验证 / 找回密码等一次性令牌
type VerificationToken struct {
	Token     string       `gorm:"primaryKey;size:64" json:"token"`
	UserID    string       `gorm:"size:36;index;not null" json:"user_id"`
	Purpose   TokenPurpose `gorm:"size:32;not null" json:"purpose"`
	ExpiresAt time.Time    `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time   `json:"used_at,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

// TableName 表名
func (VerificationToken) TableName() string { return "verification_tokens" }

// IsExpired token 是否已过期
func (t *VerificationToken) IsExpired(now time.Time) bool {
	return now.After(t.ExpiresAt)
}

// IsUsed token 是否已被使用
func (t *VerificationToken) IsUsed() bool {
	return t.UsedAt != nil
}
