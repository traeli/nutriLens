package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	platformauth "shijibu/internal/platform/auth"

	"github.com/gin-gonic/gin"
)

type stubAccountStatusChecker struct {
	active bool
	err    error
}

func (s stubAccountStatusChecker) IsAccountActive(context.Context, uint) (bool, error) {
	return s.active, s.err
}

func TestAuthenticateRejectsDeletedAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := platformauth.NewTokenManager("01234567890123456789012345678901", time.Hour)
	token, err := tokens.Issue(42)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	engine := gin.New()
	engine.Use(Authenticate(tokens, stubAccountStatusChecker{active: false}))
	engine.GET("/private", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), "账号已注销或不可用") {
		t.Fatalf("GET /private = %d %s, want deleted-account unauthorized response", response.Code, response.Body.String())
	}
}

func TestAuthenticateAllowsActiveAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := platformauth.NewTokenManager("01234567890123456789012345678901", time.Hour)
	token, err := tokens.Issue(42)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	engine := gin.New()
	engine.Use(Authenticate(tokens, stubAccountStatusChecker{active: true}))
	engine.GET("/private", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("GET /private = %d %s, want 204", response.Code, response.Body.String())
	}
}
