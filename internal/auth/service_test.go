package auth

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// memoryRepo 内存实现，用于测试
type memoryRepo struct {
	mu     sync.RWMutex
	users  map[string]*User            // id -> user
	byMail map[string]*User            // email -> user
	tokens map[string]*VerificationToken // token -> token
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		users:  make(map[string]*User),
		byMail: make(map[string]*User),
		tokens: make(map[string]*VerificationToken),
	}
}

func (r *memoryRepo) CreateUser(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byMail[user.Email]; ok {
		return errors.New("duplicate email")
	}
	r.users[user.ID] = user
	r.byMail[user.Email] = user
	return nil
}

func (r *memoryRepo) FindUserByEmail(email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byMail[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (r *memoryRepo) FindUserByID(id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (r *memoryRepo) MarkUserVerified(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return ErrUserNotFound
	}
	u.Verified = true
	return nil
}

func (r *memoryRepo) CreateToken(token *VerificationToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token.Token] = token
	return nil
}

func (r *memoryRepo) FindToken(token string) (*VerificationToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tokens[token]
	if !ok {
		return nil, ErrTokenNotFound
	}
	return t, nil
}

func (r *memoryRepo) ClaimToken(token string, usedAt time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tokens[token]
	if !ok {
		return 0, ErrTokenNotFound
	}
	if t.UsedAt != nil {
		return 0, nil // already claimed
	}
	t.UsedAt = &usedAt
	return 1, nil
}

// mockMailer 记录调用
type mockMailer struct {
	lastTo       string
	lastURL      string
	shouldError  bool
}

func (m *mockMailer) SendVerification(to, nickname, url string) error {
	m.lastTo = to
	m.lastURL = url
	if m.shouldError {
		return errors.New("send failed")
	}
	return nil
}

// newTestService 创建测试用 service
func newTestService() (*Service, *memoryRepo, *mockMailer) {
	repo := newMemoryRepo()
	mailer := &mockMailer{}
	jwt, _ := NewJWTManager("test-secret", 24*time.Hour)
	svc := NewService(repo, mailer, jwt, "http://localhost:8080")
	return svc, repo, mailer
}

func TestRegister_Success(t *testing.T) {
	svc, _, mailer := newTestService()

	result, err := svc.Register(RegisterInput{
		Email:    "a@b.com",
		Password: "Pass1234",
		Nickname: "老王",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if result.UserID == "" {
		t.Fatal("user ID is empty")
	}
	if mailer.lastTo != "a@b.com" {
		t.Errorf("mailer.lastTo = %q, want %q", mailer.lastTo, "a@b.com")
	}
	if mailer.lastURL == "" {
		t.Fatal("mailer.lastURL is empty")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _, _ := newTestService()

	_, _ = svc.Register(RegisterInput{Email: "a@b.com", Password: "Pass1234"})
	_, err := svc.Register(RegisterInput{Email: "a@b.com", Password: "Pass5678"})
	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("err = %v, want ErrEmailTaken", err)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.Register(RegisterInput{Email: "not-an-email", Password: "Pass1234"})
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("err = %v, want ErrInvalidEmail", err)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc, _, _ := newTestService()
	tests := []struct {
		name string
		pwd  string
	}{
		{"short", "Ab1"},
		{"no digit", "abcdefgh"},
		{"no letter", "12345678"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Register(RegisterInput{Email: "a@b.com", Password: tt.pwd})
			if !errors.Is(err, ErrWeakPassword) {
				t.Errorf("err = %v, want ErrWeakPassword", err)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	svc, repo, _ := newTestService()
	result, _ := svc.Register(RegisterInput{Email: "a@b.com", Password: "Pass1234", Nickname: "老王"})

	// 手动标记为已验证
	_ = repo.MarkUserVerified(result.UserID)

	login, err := svc.Login(LoginInput{Email: "a@b.com", Password: "Pass1234"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if login.Token == "" {
		t.Fatal("token is empty")
	}
	if login.User.Email != "a@b.com" {
		t.Errorf("email = %q", login.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, repo, _ := newTestService()
	result, _ := svc.Register(RegisterInput{Email: "a@b.com", Password: "Pass1234"})
	_ = repo.MarkUserVerified(result.UserID)

	_, err := svc.Login(LoginInput{Email: "a@b.com", Password: "Wrong123"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_NotVerified(t *testing.T) {
	svc, _, _ := newTestService()
	_, _ = svc.Register(RegisterInput{Email: "a@b.com", Password: "Pass1234"})

	_, err := svc.Login(LoginInput{Email: "a@b.com", Password: "Pass1234"})
	if !errors.Is(err, ErrEmailNotVerified) {
		t.Errorf("err = %v, want ErrEmailNotVerified", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.Login(LoginInput{Email: "nobody@b.com", Password: "Pass1234"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestVerifyEmail_Success(t *testing.T) {
	svc, _, mailer := newTestService()
	_, _ = svc.Register(RegisterInput{Email: "a@b.com", Password: "Pass1234"})

	// 提取验证 token
	token := extractTokenFromURL(mailer.lastURL)

	user, err := svc.VerifyEmail(token)
	if err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}
	if !user.Verified {
		t.Fatal("user should be verified")
	}

	// 二次使用应报错
	_, err = svc.VerifyEmail(token)
	if !errors.Is(err, ErrTokenUsed) {
		t.Errorf("second use: err = %v, want ErrTokenUsed", err)
	}
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.VerifyEmail("nonexistent")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyEmail_ExpiredToken(t *testing.T) {
	svc, _, mailer := newTestService()
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return baseTime }
	_, _ = svc.Register(RegisterInput{Email: "a@b.com", Password: "Pass1234"})

	token := extractTokenFromURL(mailer.lastURL)

	// 模拟过期：token TTL 为 24h，将 now 往后推 25 小时
	baseTime = baseTime.Add(25 * time.Hour)

	_, err := svc.VerifyEmail(token)
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("err = %v, want ErrTokenExpired", err)
	}
}

func TestResendVerification_Throttled(t *testing.T) {
	svc, _, _ := newTestService()
	_, _ = svc.Register(RegisterInput{Email: "a@b.com", Password: "Pass1234"})

	err := svc.ResendVerification("a@b.com")
	if err != nil {
		t.Fatalf("first resend: %v", err)
	}

	err = svc.ResendVerification("a@b.com")
	if !errors.Is(err, ErrResendThrottled) {
		t.Errorf("second resend: err = %v, want ErrResendThrottled", err)
	}
}

// extractTokenFromURL 从 verify URL 提取 token 参数
func extractTokenFromURL(rawURL string) string {
	// 形如 http://localhost:8080/api/v1/auth/verify?token=xxx
	for i := len(rawURL) - 1; i >= 0; i-- {
		if rawURL[i] == '=' {
			return rawURL[i+1:]
		}
	}
	return ""
}
