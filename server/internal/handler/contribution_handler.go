package handler

import (
	"net/http"

	"shijibu/internal/middleware"
	"shijibu/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Contribution(c *gin.Context) {
	result, err := h.Contributions.Get(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

func (h *Handler) Badges(c *gin.Context) {
	items, err := h.Contributions.Badges(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"badges": items})
}
