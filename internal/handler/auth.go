package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/auth"
	"github.com/cwr0401/f3moon/internal/middleware"
)

// AuthHandler 认证接口处理器
type AuthHandler struct {
	svc *auth.Service
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register 注册
// @Summary 用户注册
// @Description 注册新用户并发送验证邮件
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body object true "注册信息" example({"email":"user@example.com","password":"secret123","nickname":"玩家1"})
// @Success 200 {object} object "{ \"user_id\": \"string\", \"message\": \"验证邮件已发送，请查收邮箱\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 409 {object} object "{ \"error\": \"email already taken\" }"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.Register(auth.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, auth.ErrEmailTaken) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  result.UserID,
		"message":  "验证邮件已发送，请查收邮箱",
	})
}

// Login 登录
// @Summary 用户登录
// @Description 使用邮箱和密码登录，返回 JWT token
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body object true "登录信息" example({"email":"user@example.com","password":"secret123"})
// @Success 200 {object} object "{ \"token\": \"string\", \"expires_at\": \"string\", \"user\": {} }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 401 {object} object "{ \"error\": \"invalid credentials\" }"
// @Failure 403 {object} object "{ \"error\": \"email not verified\" }"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.Login(auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, auth.ErrEmailNotVerified):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      result.Token,
		"expires_at": result.ExpiresAt,
		"user":       result.User,
	})
}

// Verify 邮箱验证
// @Summary 验证邮箱
// @Description 通过邮件中的 token 验证用户邮箱
// @Tags 认证
// @Produce html
// @Param token query string true "验证令牌"
// @Success 200 {string} string "验证成功页面"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 410 {object} object "{ \"error\": \"token already used\" }"
// @Router /auth/verify [get]
func (h *AuthHandler) Verify(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}

	user, err := h.svc.VerifyEmail(token)
	if err != nil {
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, auth.ErrInvalidToken), errors.Is(err, auth.ErrTokenExpired):
			status = http.StatusBadRequest
		case errors.Is(err, auth.ErrTokenUsed):
			status = http.StatusGone
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "verified.html", gin.H{
		"nickname": user.Nickname,
		"email":    user.Email,
	})
}

// Resend 重发验证邮件
// @Summary 重发验证邮件
// @Description 重新发送邮箱验证邮件（有频率限制）
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body object true "邮箱" example({"email":"user@example.com"})
// @Success 200 {object} object "{ \"message\": \"如果该邮箱已注册且未验证，验证邮件已发送\" }"
// @Failure 400 {object} object "{ \"error\": \"string\" }"
// @Failure 429 {object} object "{ \"error\": \"resend throttled\" }"
// @Router /auth/resend [post]
func (h *AuthHandler) Resend(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.svc.ResendVerification(req.Email)
	if err != nil {
		if errors.Is(err, auth.ErrResendThrottled) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "如果该邮箱已注册且未验证，验证邮件已发送"})
}

// Me 获取当前用户
// @Summary 获取当前用户信息
// @Description 返回当前已认证用户的信息
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "用户信息"
// @Failure 404 {object} object "{ \"error\": \"user not found\" }"
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	user, err := h.svc.CurrentUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
