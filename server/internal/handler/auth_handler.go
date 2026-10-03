package handler

import (
	"net/http"

	"shijibu/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Login(c *gin.Context) {
	var input struct {
		Code string `json:"code" binding:"required"` // 登陆凭证
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "登录凭证不能为空")
		return
	}
	result, err := h.Accounts.Login(c.Request.Context(), input.Code)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required,max=4096"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "刷新凭证不能为空")
		return
	}
	result, err := h.Accounts.Refresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}
