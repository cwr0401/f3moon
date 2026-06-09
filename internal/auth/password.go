package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcrypt 工作因子；cost=12 在普通服务器上约 250ms
const bcryptCost = 12

// HashPassword 对密码做 bcrypt 哈希
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", fmt.Errorf("password is empty")
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}

// VerifyPassword 校验明文密码是否匹配哈希
func VerifyPassword(hash, plain string) bool {
	if hash == "" || plain == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
