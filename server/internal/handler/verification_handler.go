package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"shijibu/internal/middleware"
	"shijibu/internal/platform/httpx"
)

func (h *Handler) PublisherVerificationStatus(c *gin.Context) {
	item, err := h.Verification.Status(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, item)
}

func (h *Handler) VerifyPublisherPhone(c *gin.Context) {
	var input struct {
		Code string `json:"code" binding:"required"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "手机号授权凭证不能为空")
		return
	}
	item, err := h.Verification.VerifyPhone(c.Request.Context(), middleware.UserID(c), input.Code)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, item)
}
