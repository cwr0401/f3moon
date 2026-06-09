package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken token 无效
var ErrInvalidToken = errors.New("invalid token")

// Claims JWT 声明
type Claims struct {
	UserID   string `json:"uid"`
	Email    string `json:"email"`
	Nickname string `json:"nickname,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager 负责签发和校验 JWT
type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(secret string, ttl time.Duration) (*JWTManager, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is empty")
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &JWTManager{secret: []byte(secret), ttl: ttl}, nil
}

// IssueResult 签发结果
type IssueResult struct {
	Token     string
	ExpiresAt time.Time
}

// Issue 为用户签发 token
func (m *JWTManager) Issue(userID, email, nickname string) (*IssueResult, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)
	claims := Claims{
		UserID:   userID,
		Email:    email,
		Nickname: nickname,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tk.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}
	return &IssueResult{Token: signed, ExpiresAt: expiresAt}, nil
}

// Parse 解析并校验 token
func (m *JWTManager) Parse(tokenStr string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	if claims.UserID == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
