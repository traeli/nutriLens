package bailian

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

func TestClientTranscribe(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/compatible-mode/v1/chat/completions" || r.Method != http.MethodPost {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		var payload completionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Model != "qwen3-asr-flash" || !payload.ASROptions.EnableITN {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		data := payload.Messages[0].Content[0].InputAudio.Data
		if !strings.HasPrefix(data, "data:audio/mpeg;base64,") {
			t.Fatalf("audio data URL = %q", data)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":" 人均 80 元，味道推荐。 "}}]}`)),
		}, nil
	})

	client := NewClient("test-key", "https://workspace.example.com/compatible-mode/v1", "qwen3-asr-flash")
	client.httpClient = &http.Client{Transport: transport}
	text, err := client.Transcribe(context.Background(), []byte("audio"), "audio/mpeg")
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}
	if text != "人均 80 元，味道推荐。" {
		t.Fatalf("Transcribe() = %q", text)
	}
}

func TestClientRequiresConfiguration(t *testing.T) {
	_, err := NewClient("", "", "").Transcribe(context.Background(), []byte("audio"), "audio/mpeg")
	if err != ErrNotConfigured {
		t.Fatalf("Transcribe() error = %v, want ErrNotConfigured", err)
	}
}
