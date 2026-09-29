package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"shijibu/internal/handler"
	platformauth "shijibu/internal/platform/auth"

	"github.com/gin-gonic/gin"
)

func TestHealthAndRemovedLegacyRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := platformauth.NewTokenManager("01234567890123456789012345678901", time.Hour)
	engine := New(&handler.Handler{}, tokens, nil)

	health := httptest.NewRecorder()
	engine.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"status":"ok"`) {
		t.Fatalf("GET /health = %d %s", health.Code, health.Body.String())
	}

	legacy := httptest.NewRecorder()
	engine.ServeHTTP(legacy, httptest.NewRequest(http.MethodGet, "/api/v1/food/records", nil))
	if legacy.Code != http.StatusNotFound {
		t.Fatalf("legacy route status = %d, want 404", legacy.Code)
	}
}

func TestHomeSummaryIsPubliclyReachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := platformauth.NewTokenManager("01234567890123456789012345678901", time.Hour)
	engine := New(&handler.Handler{}, tokens, nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/home/summary?city_code=310000", nil))
	if response.Code == http.StatusUnauthorized || response.Code == http.StatusNotFound {
		t.Fatalf("home summary route status = %d, want a registered public route", response.Code)
	}
}

func TestNearbyRouteIsPubliclyReachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := platformauth.NewTokenManager("01234567890123456789012345678901", time.Hour)
	engine := New(&handler.Handler{}, tokens, nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/routes/nearby?city_code=310000&longitude=121.47&latitude=31.23", nil))
	if response.Code == http.StatusUnauthorized || response.Code == http.StatusNotFound {
		t.Fatalf("nearby route status = %d, want a registered public route", response.Code)
	}
}

func TestProtectedRouteRequiresBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := platformauth.NewTokenManager("01234567890123456789012345678901", time.Hour)
	engine := New(&handler.Handler{}, tokens, nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("protected route status = %d, want 401", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"code":"UNAUTHORIZED"`) {
		t.Fatalf("unexpected error response: %s", response.Body.String())
	}
}

func TestSpeechTranscriptionRequiresBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := platformauth.NewTokenManager("01234567890123456789012345678901", time.Hour)
	engine := New(&handler.Handler{}, tokens, nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/speech/transcribe", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("speech route status = %d, want 401", response.Code)
	}
}

func TestSubmitPlaceRequiresBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := platformauth.NewTokenManager("01234567890123456789012345678901", time.Hour)
	engine := New(&handler.Handler{}, tokens, nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/places", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("submit place route status = %d, want 401", response.Code)
	}
}

func TestNewMiniappWriteRoutesRequireBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := platformauth.NewTokenManager("01234567890123456789012345678901", time.Hour)
	engine := New(&handler.Handler{}, tokens, nil)
	cases := []struct{ method, path string }{
		{http.MethodPost, "/api/v1/reports"},
		{http.MethodPost, "/api/v1/records/1/media"},
		{http.MethodPost, "/api/v1/publisher-verification/phone"},
		{http.MethodPost, "/api/v1/nutrition/records"},
		{http.MethodPost, "/api/v1/routes/1/start"},
	}
	for _, item := range cases {
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(item.method, item.path, nil))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s = %d, want 401", item.method, item.path, response.Code)
		}
	}
}
