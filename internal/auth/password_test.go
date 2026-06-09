package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("Pass1234")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" {
		t.Fatal("hash is empty")
	}
	if hash == "Pass1234" {
		t.Fatal("hash equals plain password")
	}
}

func TestVerifyPassword(t *testing.T) {
	hash, _ := HashPassword("Pass1234")

	tests := []struct {
		name  string
		hash  string
		plain string
		want  bool
	}{
		{"correct", hash, "Pass1234", true},
		{"wrong password", hash, "Wrong123", false},
		{"empty hash", "", "Pass1234", false},
		{"empty plain", hash, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyPassword(tt.hash, tt.plain); got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashPassword_Empty(t *testing.T) {
	_, err := HashPassword("")
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}
