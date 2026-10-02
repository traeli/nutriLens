package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	platformauth "shijibu/internal/platform/auth"
	"strings"
	"testing"

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
	tokens := stubTokenAuthenticator{}
	token := "test-access-token"
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
	tokens := stubTokenAuthenticator{}
	token := "test-access-token"
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

type stubTokenAuthenticator struct{ err error }

func (s stubTokenAuthenticator) Authenticate(context.Context, string) (uint, error) { return 42, s.err }

func TestAuthenticateDistinguishesInvalidSessionAndRedisOutage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, item := range []struct {
		err    error
		status int
	}{
		{platformauth.ErrUnauthorized, http.StatusUnauthorized},
		{platformauth.ErrStoreUnavailable, http.StatusServiceUnavailable},
	} {
		engine := gin.New()
		engine.Use(Authenticate(stubTokenAuthenticator{err: item.err}))
		reached := false
		engine.GET("/private", func(c *gin.Context) { reached = true; c.Status(http.StatusNoContent) })
		request := httptest.NewRequest(http.MethodGet, "/private", nil)
		request.Header.Set("Authorization", "Bearer test-token")
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, request)
		if response.Code != item.status || reached {
			t.Fatalf("status=%d reached=%v, want %d and no handler execution", response.Code, reached, item.status)
		}
	}
}
