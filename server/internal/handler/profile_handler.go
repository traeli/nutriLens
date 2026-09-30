package handler

import (
	"net/http"

	"shijibu/internal/middleware"
	"shijibu/internal/platform/httpx"
	"shijibu/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetProfile(c *gin.Context) {
	profile, err := h.Accounts.Get(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, profile)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var input struct {
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatar_url"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请求参数不正确")
		return
	}
	profile, err := h.Accounts.Update(middleware.UserID(c), input.Nickname, input.AvatarURL)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, profile)
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	if err := h.Accounts.DeleteAccount(c.Request.Context(), middleware.UserID(c)); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) AcceptAgreement(c *gin.Context) {
	var input struct {
		AgreementType string `json:"agreement_type" binding:"required"`
		Version       string `json:"version" binding:"required"`
		ClientVersion string `json:"client_version"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "协议类型和版本不能为空")
		return
	}
	err := h.Accounts.AcceptAgreement(middleware.UserID(c), input.AgreementType, input.Version, "", input.ClientVersion)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"accepted": true})
}

func (h *Handler) CurrentAgreements(c *gin.Context) {
	items := make([]gin.H, 0, len(service.RequiredAgreementTypes()))
	for _, agreementType := range service.RequiredAgreementTypes() {
		items = append(items, gin.H{"agreement_type": agreementType, "version": service.CurrentAgreementVersion})
	}
	httpx.OK(c, http.StatusOK, gin.H{"items": items})
}
