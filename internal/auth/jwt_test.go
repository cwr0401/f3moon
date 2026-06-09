package auth

import (
	"testing"
	"time"
)

func TestJWTManager_IssueAndParse(t *testing.T) {
	mgr, err := NewJWTManager("test-secret", 1*time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager: %v", err)
	}

	result, err := mgr.Issue("user-123", "a@b.com", "老王")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if result.Token == "" {
		t.Fatal("token is empty")
	}

	claims, err := mgr.Parse(result.Token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Email != "a@b.com" {
		t.Errorf("Email = %q, want %q", claims.Email, "a@b.com")
	}
}

func TestJWTManager_Expired(t *testing.T) {
	mgr, _ := NewJWTManager("test-secret", -1*time.Nanosecond)
	// 负 TTL 会被默认为 24h，手动改
	mgr.ttl = -1 * time.Nanosecond

	result, err := mgr.Issue("user-1", "a@b.com", "")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	_, err = mgr.Parse(result.Token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	mgr, _ := NewJWTManager("test-secret", 1*time.Hour)

	_, err := mgr.Parse("not-a-jwt")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestJWTManager_WrongSecret(t *testing.T) {
	mgr1, _ := NewJWTManager("secret-a", 1*time.Hour)
	mgr2, _ := NewJWTManager("secret-b", 1*time.Hour)

	result, _ := mgr1.Issue("user-1", "a@b.com", "")
	_, err := mgr2.Parse(result.Token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestJWTManager_EmptySecret(t *testing.T) {
	_, err := NewJWTManager("", 1*time.Hour)
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}
