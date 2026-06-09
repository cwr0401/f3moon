package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/auth"
)

const (
	// HeaderKey Authorization header
	HeaderKey = "Authorization"
	// ContextKeyUserID 存入 gin.Context 的 user_id key
	ContextKeyUserID = "user_id"
	// ContextKeyEmail 存入 gin.Context 的 email key
	ContextKeyEmail = "email"
	// ContextKeyNickname 存入 gin.Context 的 nickname key
	ContextKeyNickname = "nickname"
)

// JWTAuth 返回 JWT 校验中间件
func JWTAuth(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			return
		}

		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyNickname, claims.Nickname)
		c.Next()
	}
}

// extractToken 从 Header 或 Query 提取 Bearer token
func extractToken(c *gin.Context) string {
	// 1. Authorization: Bearer xxx
	h := c.GetHeader(HeaderKey)
	if h != "" {
		parts := strings.SplitN(h, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
		// 非 Bearer 格式的 Authorization header，不回退
		return ""
	}
	// 2. ?token=xxx (WebSocket 等不便加 header 的场景)
	if t := c.Query("token"); t != "" {
		return t
	}
	return ""
}
