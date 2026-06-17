package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrEmailTaken 邮箱已被注册
	ErrEmailTaken = errors.New("email already registered")
	// ErrInvalidCredentials 邮箱或密码错误
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrEmailNotVerified 邮箱未验证
	ErrEmailNotVerified = errors.New("email not verified")
	// ErrInvalidEmail 邮箱格式不正确
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrWeakPassword 密码不符合强度要求
	ErrWeakPassword = errors.New("password must be at least 8 characters and contain 3 of: uppercase, lowercase, digit, special character")
	// ErrTokenExpired 验证 token 已过期
	ErrTokenExpired = errors.New("token expired")
	// ErrTokenUsed 验证 token 已被使用
	ErrTokenUsed = errors.New("token already used")
	// ErrResendThrottled 重发过于频繁
	ErrResendThrottled = errors.New("resend throttled, please wait")
	// ErrWrongPassword 原密码错误
	ErrWrongPassword = errors.New("current password is incorrect")
)

const (
	verifyTokenTTL    = 24 * time.Hour
	minPasswordLength = 8
	resendCooldown    = 60 * time.Second
)

// 邮箱格式校验（实用够用版）
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// 密码强度：必须含字母与数字
var (
	reHasLetter = regexp.MustCompile(`[A-Za-z]`)
	reHasDigit  = regexp.MustCompile(`[0-9]`)
)

// RegisterInput 注册输入
type RegisterInput struct {
	Email    string
	Password string
	Nickname string
}

// RegisterResult 注册结果
type RegisterResult struct {
	UserID    string
	VerifyURL string `json:"-"` // 不序列化，防止验证链接泄露
}

// LoginInput 登录输入
type LoginInput struct {
	Email    string
	Password string
}

// LoginResult 登录结果
type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	User      *User
}

// Service 认证业务
type Service struct {
	repo       Repository
	mailer     Mailer
	jwt        *JWTManager
	appBaseURL string

	now func() time.Time // 时钟，方便测试

	resendMu       sync.Mutex
	lastResendByID map[string]time.Time // userID -> last sent
}

// NewService 创建认证服务
func NewService(repo Repository, mailer Mailer, jwt *JWTManager, appBaseURL string) *Service {
	return &Service{
		repo:           repo,
		mailer:         mailer,
		jwt:            jwt,
		appBaseURL:     strings.TrimRight(appBaseURL, "/"),
		now:            time.Now,
		lastResendByID: make(map[string]time.Time),
	}
}

// Register 注册
func (s *Service) Register(in RegisterInput) (*RegisterResult, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	if err := validatePassword(in.Password); err != nil {
		return nil, err
	}

	if _, err := s.repo.FindUserByEmail(email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	hash, err := HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: hash,
		Nickname:     strings.TrimSpace(in.Nickname),
		Verified:     false,
		CreatedAt:    s.now(),
		UpdatedAt:    s.now(),
	}
	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	verifyURL, err := s.issueVerificationToken(user)
	if err != nil {
		// 注册成功但邮件 token 创建失败，记录日志即可（用户可走 resend）
		log.Printf("[auth] issue verification token for %s failed: %v", user.Email, err)
		return &RegisterResult{UserID: user.ID}, nil
	}

	if err := s.mailer.SendVerification(user.Email, user.Nickname, verifyURL); err != nil {
		log.Printf("[auth] send verification email to %s failed: %v", user.Email, err)
	}

	return &RegisterResult{UserID: user.ID, VerifyURL: verifyURL}, nil
}

// Login 登录
func (s *Service) Login(in LoginInput) (*LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || in.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.FindUserByEmail(email)
	if errors.Is(err, ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if !VerifyPassword(user.PasswordHash, in.Password) {
		return nil, ErrInvalidCredentials
	}
	if !user.Verified {
		return nil, ErrEmailNotVerified
	}

	issued, err := s.jwt.Issue(user.ID, user.Email, user.Nickname)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token:     issued.Token,
		ExpiresAt: issued.ExpiresAt,
		User:      user,
	}, nil
}

// VerifyEmail 通过 token 激活邮箱
func (s *Service) VerifyEmail(token string) (*User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrInvalidToken
	}

	t, err := s.repo.FindToken(token)
	if errors.Is(err, ErrTokenNotFound) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	if t.IsUsed() {
		return nil, ErrTokenUsed
	}
	if t.IsExpired(s.now()) {
		return nil, ErrTokenExpired
	}
	if t.Purpose != PurposeVerifyEmail {
		return nil, ErrInvalidToken
	}

	// 原子性 claim：WHERE used_at IS NULL，防止 TOCTOU 竞态
	affected, err := s.repo.ClaimToken(t.Token, s.now())
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		// 被并发请求抢先 claim
		return nil, ErrTokenUsed
	}

	if err := s.repo.MarkUserVerified(t.UserID); err != nil {
		return nil, err
	}

	return s.repo.FindUserByID(t.UserID)
}

// ResendVerification 重发验证邮件
func (s *Service) ResendVerification(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.repo.FindUserByEmail(email)
	if errors.Is(err, ErrUserNotFound) {
		// 不暴露用户是否存在
		return nil
	}
	if err != nil {
		return err
	}
	if user.Verified {
		return nil
	}

	s.resendMu.Lock()
	last, ok := s.lastResendByID[user.ID]
	if ok && s.now().Sub(last) < resendCooldown {
		s.resendMu.Unlock()
		return ErrResendThrottled
	}
	s.lastResendByID[user.ID] = s.now()
	// 顺便清理过期条目，防止内存泄漏
	cutoff := s.now().Add(-resendCooldown)
	for id, t := range s.lastResendByID {
		if t.Before(cutoff) {
			delete(s.lastResendByID, id)
		}
	}
	s.resendMu.Unlock()

	verifyURL, err := s.issueVerificationToken(user)
	if err != nil {
		return err
	}
	if err := s.mailer.SendVerification(user.Email, user.Nickname, verifyURL); err != nil {
		log.Printf("[auth] resend verification email to %s failed: %v", user.Email, err)
	}
	return nil
}

// CurrentUser 通过 userID 查询用户（供 /me 等接口使用）
func (s *Service) CurrentUser(userID string) (*User, error) {
	return s.repo.FindUserByID(userID)
}

// JWT 暴露 JWT 管理器（供中间件复用同一实例）
func (s *Service) JWT() *JWTManager { return s.jwt }

// issueVerificationToken 生成并持久化一次性验证 token，返回完整 URL
func (s *Service) issueVerificationToken(user *User) (string, error) {
	rawToken, err := generateToken(32)
	if err != nil {
		return "", err
	}
	t := &VerificationToken{
		Token:     rawToken,
		UserID:    user.ID,
		Purpose:   PurposeVerifyEmail,
		ExpiresAt: s.now().Add(verifyTokenTTL),
		CreatedAt: s.now(),
	}
	if err := s.repo.CreateToken(t); err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("token", rawToken)
	return fmt.Sprintf("%s/api/v1/auth/verify?%s", s.appBaseURL, q.Encode()), nil
}

func validatePassword(pwd string) error {
	if len(pwd) < minPasswordLength {
		return ErrWeakPassword
	}
	if !reHasLetter.MatchString(pwd) || !reHasDigit.MatchString(pwd) {
		return ErrWeakPassword
	}
	return nil
}

// generateToken 生成 nBytes 字节的随机 hex 字符串
func generateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return hex.EncodeToString(b), nil
}
