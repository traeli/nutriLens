package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestClientAnalyzeNutrition(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/chat/completions" || r.Method != http.MethodPost {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		var payload completionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Model != "deepseek-flash" || payload.ResponseFormat["type"] != "json_object" || payload.MaxTokens != 1200 || payload.EnableThinking == nil || *payload.EnableThinking {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		if len(payload.Messages) != 2 || !strings.HasSuffix(payload.Messages[1].Content, "我吃了一个苹果") || !strings.Contains(payload.Messages[0].Content, "protein_grams") {
			t.Fatalf("unexpected messages: %+v", payload.Messages)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"{\"foods\":[{\"name\":\"苹果\"}],\"calories\":95}"}}]}`)),
		}, nil
	})

	client := NewClient("test-key", "https://api.deepseek.com", "deepseek-flash", "disabled")
	client.httpClient = &http.Client{Transport: transport}
	content, err := client.AnalyzeNutrition(context.Background(), " 我吃了一个苹果 ")
	if err != nil {
		t.Fatalf("AnalyzeNutrition() error = %v", err)
	}
	if !strings.Contains(content, `"calories":95`) {
		t.Fatalf("AnalyzeNutrition() = %q", content)
	}
}

func TestClientRequiresConfiguration(t *testing.T) {
	_, err := NewClient("", "", "", "disabled").AnalyzeNutrition(context.Background(), "苹果")
	if err != ErrNotConfigured {
		t.Fatalf("AnalyzeNutrition() error = %v, want ErrNotConfigured", err)
	}
}

func TestThinkingSettingCanUseProviderDefault(t *testing.T) {
	if thinkingSetting("provider_default") != nil {
		t.Fatal("provider_default should omit enable_thinking")
	}
	value := thinkingSetting("disabled")
	if value == nil || *value {
		t.Fatal("disabled should send enable_thinking=false")
	}
}
