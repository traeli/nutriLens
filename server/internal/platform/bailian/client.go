package bailian

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrNotConfigured = errors.New("bailian speech recognition is not configured")

type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

type completionRequest struct {
	Model      string              `json:"model"`
	Messages   []completionMessage `json:"messages"`
	Stream     bool                `json:"stream"`
	ASROptions asrOptions          `json:"asr_options"`
}

type completionMessage struct {
	Role    string              `json:"role"`
	Content []completionContent `json:"content"`
}

type completionContent struct {
	Type       string     `json:"type"`
	InputAudio inputAudio `json:"input_audio"`
}

type inputAudio struct {
	Data string `json:"data"`
}

type asrOptions struct {
	EnableITN bool `json:"enable_itn"`
}

type completionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		apiKey:     strings.TrimSpace(apiKey),
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		model:      strings.TrimSpace(model),
		httpClient: &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *Client) Transcribe(ctx context.Context, audio []byte, mediaType string) (string, error) {
	if c.apiKey == "" || c.baseURL == "" || c.model == "" {
		return "", ErrNotConfigured
	}
	payload := completionRequest{
		Model: c.model,
		Messages: []completionMessage{{
			Role: "user",
			Content: []completionContent{{
				Type:       "input_audio",
				InputAudio: inputAudio{Data: "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(audio)},
			}},
		}},
		Stream:     false,
		ASROptions: asrOptions{EnableITN: true},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode bailian request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create bailian request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call bailian: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read bailian response: %w", err)
	}
	var result completionResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", fmt.Errorf("decode bailian response with status %d: %w", response.StatusCode, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("bailian returned status %d code %q", response.StatusCode, result.Error.Code)
	}
	if len(result.Choices) == 0 {
		return "", errors.New("bailian response contains no transcription")
	}
	text := strings.TrimSpace(result.Choices[0].Message.Content)
	if text == "" {
		return "", errors.New("bailian returned an empty transcription")
	}
	return text, nil
}
