package service

import (
	"context"
	"fmt"
	"strings"
)

// Base64 adds roughly one third to the request size. Seven MiB keeps the
// encoded audio below Bailian's ten MiB synchronous-input limit.
const MaxSpeechAudioBytes int64 = 7 << 20

type SpeechRecognizer interface {
	Transcribe(ctx context.Context, audio []byte, mediaType string) (string, error)
}

type SpeechService struct {
	recognizer SpeechRecognizer
}

func NewSpeechService(recognizer SpeechRecognizer) *SpeechService {
	return &SpeechService{recognizer: recognizer}
}

func (s *SpeechService) Transcribe(ctx context.Context, audio []byte, mediaType string) (string, error) {
	if s == nil || s.recognizer == nil {
		return "", ErrUnavailable
	}
	if len(audio) == 0 || int64(len(audio)) > MaxSpeechAudioBytes || !supportedSpeechMediaType(mediaType) {
		return "", ErrInvalidInput
	}
	text, err := s.recognizer.Transcribe(ctx, audio, mediaType)
	if err != nil {
		return "", fmt.Errorf("recognize speech: %w: %w", ErrUnavailable, err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ErrUnavailable
	}
	return text, nil
}

func supportedSpeechMediaType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "audio/aac", "audio/amr", "audio/flac", "audio/mpeg", "audio/ogg", "audio/opus", "audio/wav", "audio/webm", "audio/x-ms-wma", "audio/x-wav":
		return true
	default:
		return false
	}
}
