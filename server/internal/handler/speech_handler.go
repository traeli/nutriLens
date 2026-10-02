package handler

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"shijibu/internal/platform/httpx"
	"shijibu/internal/service"

	"github.com/gin-gonic/gin"
)

const maxSpeechRequestBytes = service.MaxSpeechAudioBytes + 1<<20

func (h *Handler) TranscribeSpeech(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSpeechRequestBytes)
	header, err := c.FormFile("audio")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请选择需要识别的音频")
			return
		}
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "音频上传失败或文件过大")
		return
	}
	if header.Size <= 0 || header.Size > service.MaxSpeechAudioBytes {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "音频不能为空且不能超过 7MB")
		return
	}
	mediaType := speechMediaType(header)
	if mediaType == "" {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "暂不支持该音频格式")
		return
	}
	file, err := header.Open()
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "无法读取音频")
		return
	}
	defer file.Close()
	audio, err := io.ReadAll(io.LimitReader(file, service.MaxSpeechAudioBytes+1))
	if err != nil || int64(len(audio)) > service.MaxSpeechAudioBytes {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "无法读取音频或文件过大")
		return
	}
	text, err := h.Speech.Transcribe(c.Request.Context(), audio, mediaType)
	if errors.Is(err, service.ErrSpeechNotConfigured) {
		httpx.Error(c, http.StatusServiceUnavailable, "SPEECH_NOT_CONFIGURED", "语音识别服务尚未配置，请联系管理员")
		return
	}
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"text": text})
}

func speechMediaType(header *multipart.FileHeader) string {
	extensionTypes := map[string]string{
		".aac": "audio/aac", ".amr": "audio/amr", ".flac": "audio/flac", ".mp3": "audio/mpeg",
		".mpeg": "audio/mpeg", ".ogg": "audio/ogg", ".opus": "audio/opus", ".wav": "audio/wav",
		".webm": "audio/webm", ".wma": "audio/x-ms-wma",
	}
	if mediaType := extensionTypes[strings.ToLower(filepath.Ext(header.Filename))]; mediaType != "" {
		return mediaType
	}
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(header.Header.Get("Content-Type"), ";")[0]))
	for _, supported := range extensionTypes {
		if mediaType == supported {
			return mediaType
		}
	}
	return ""
}
