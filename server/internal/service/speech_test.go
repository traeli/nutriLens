package service

import (
	"context"
	"errors"
	"testing"
)

type speechRecognizerStub struct {
	text string
	err  error
}

func (s speechRecognizerStub) Transcribe(context.Context, []byte, string) (string, error) {
	return s.text, s.err
}

func TestSpeechServiceTranscribe(t *testing.T) {
	service := NewSpeechService(speechRecognizerStub{text: "  这家店值得推荐  "})
	text, err := service.Transcribe(context.Background(), []byte("audio"), "audio/mpeg")
	if err != nil || text != "这家店值得推荐" {
		t.Fatalf("Transcribe() = %q, %v", text, err)
	}
}

func TestSpeechServiceRejectsInvalidAudio(t *testing.T) {
	service := NewSpeechService(speechRecognizerStub{})
	for name, test := range map[string]struct {
		audio     []byte
		mediaType string
	}{
		"empty":       {audio: nil, mediaType: "audio/mpeg"},
		"unsupported": {audio: []byte("audio"), mediaType: "video/mp4"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := service.Transcribe(context.Background(), test.audio, test.mediaType)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Transcribe() error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestSpeechServiceHidesProviderFailure(t *testing.T) {
	service := NewSpeechService(speechRecognizerStub{err: errors.New("provider failed")})
	_, err := service.Transcribe(context.Background(), []byte("audio"), "audio/mpeg")
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Transcribe() error = %v, want ErrUnavailable", err)
	}
}
