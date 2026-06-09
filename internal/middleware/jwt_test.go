package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cwr0401/f3moon/internal/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestJWTAuth_ValidToken(t *testing.T) {
	jwtMgr, _ := auth.NewJWTManager("test-secret", 1*time.Hour)
	result, _ := jwtMgr.Issue("user-1", "a@b.com", "老王")

	r := gin.New()
	r.Use(JWTAuth(jwtMgr))
	r.GET("/test", func(c *gin.Context) {
		uid := c.GetString(ContextKeyUserID)
		c.String(http.StatusOK, uid)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+result.Token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "user-1" {
		t.Errorf("body = %q, want %q", w.Body.String(), "user-1")
	}
}

func TestJWTAuth_MissingToken(t *testing.T) {
	jwtMgr, _ := auth.NewJWTManager("test-secret", 1*time.Hour)

	r := gin.New()
	r.Use(JWTAuth(jwtMgr))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	jwtMgr, _ := auth.NewJWTManager("test-secret", 1*time.Hour)

	r := gin.New()
	r.Use(JWTAuth(jwtMgr))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_QueryToken(t *testing.T) {
	jwtMgr, _ := auth.NewJWTManager("test-secret", 1*time.Hour)
	result, _ := jwtMgr.Issue("user-2", "b@b.com", "")

	r := gin.New()
	r.Use(JWTAuth(jwtMgr))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString(ContextKeyUserID))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?token="+result.Token, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "user-2" {
		t.Errorf("body = %q, want %q", w.Body.String(), "user-2")
	}
}
