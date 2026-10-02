package handler

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"shijibu/internal/platform/bailian"
	"shijibu/internal/service"
)

type speechRecognizerStub struct{ err error }

func (s speechRecognizerStub) Transcribe(_ context.Context, audio []byte, mediaType string) (string, error) {
	if string(audio) != "test-audio" || mediaType != "audio/mpeg" {
		return "", errors.New("unexpected audio upload")
	}
	return "这家餐厅值得推荐", s.err
}

func TestTranscribeSpeechMultipart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, item := range []struct {
		name       string
		filename   string
		recognizer service.SpeechRecognizer
		status     int
		expected   string
	}{
		{"mp3 upload", "recording.mp3", speechRecognizerStub{}, 200, "这家餐厅值得推荐"},
		{"missing API key", "recording.mp3", bailian.NewClient("", "https://example.test/v1", "qwen3-asr-flash"), 503, "SPEECH_NOT_CONFIGURED"},
		{"unavailable provider", "recording.mp3", speechRecognizerStub{err: errors.New("private provider details")}, 503, "SERVICE_UNAVAILABLE"},
		{"unsupported format", "recording.txt", speechRecognizerStub{}, 400, "暂不支持该音频格式"},
	} {
		t.Run(item.name, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			file, err := writer.CreateFormFile("audio", item.filename)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := file.Write([]byte("test-audio")); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			h := &Handler{Speech: service.NewSpeechService(item.recognizer)}
			engine := gin.New()
			engine.POST("/speech/transcribe", h.TranscribeSpeech)
			request := httptest.NewRequest(http.MethodPost, "/speech/transcribe", &body)
			request.Header.Set("Content-Type", writer.FormDataContentType())
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)
			if response.Code != item.status || !strings.Contains(response.Body.String(), item.expected) {
				t.Fatalf("response = %d %s", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), "private provider details") {
				t.Fatal("provider details leaked")
			}
		})
	}
}
