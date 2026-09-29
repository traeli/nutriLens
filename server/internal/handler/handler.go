package handler

import (
	"errors"
	"net/http"

	"shijibu/internal/platform/httpx"
	"shijibu/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Accounts      *service.AccountService
	Discovery     *service.DiscoveryService
	Records       *service.RecordService
	Contributions *service.ContributionService
	Speech        *service.SpeechService
	Engagement    *service.EngagementService
	Nutrition     *service.NutritionService
	Media         *service.MediaService
	Verification  *service.VerificationService
}

func New(accounts *service.AccountService, discovery *service.DiscoveryService, records *service.RecordService, contributions *service.ContributionService, speech *service.SpeechService, engagement *service.EngagementService, nutrition *service.NutritionService, media *service.MediaService, verification *service.VerificationService) *Handler {
	return &Handler{Accounts: accounts, Discovery: discovery, Records: records, Contributions: contributions, Speech: speech, Engagement: engagement, Nutrition: nutrition, Media: media, Verification: verification}
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请求参数不正确")
	case errors.Is(err, service.ErrNotFound):
		httpx.Error(c, http.StatusNotFound, "NOT_FOUND", "未找到对应内容")
	case errors.Is(err, service.ErrConflict):
		httpx.Error(c, http.StatusConflict, "STATE_CONFLICT", "当前状态不允许此操作")
	case errors.Is(err, service.ErrForbidden):
		httpx.Error(c, http.StatusForbidden, "FORBIDDEN", "无权执行此操作")
	case errors.Is(err, service.ErrUnavailable):
		httpx.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "语音识别暂时不可用")
	case errors.Is(err, service.ErrRateLimited):
		httpx.Error(c, http.StatusTooManyRequests, "RATE_LIMITED", "操作过于频繁，请稍后再试")
	case errors.Is(err, service.ErrDisabled):
		httpx.Error(c, http.StatusServiceUnavailable, "FEATURE_DISABLED", "该功能暂时关闭")
	default:
		httpx.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时不可用")
	}
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	var input struct {
		ID uint `uri:"id" binding:"required"`
	}
	if name != "id" || c.ShouldBindUri(&input) != nil || input.ID == 0 {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "ID 不正确")
		return 0, false
	}
	return input.ID, true
}
