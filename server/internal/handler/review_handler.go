package handler

import (
	"net/http"
	"strconv"

	"shijibu/internal/middleware"
	"shijibu/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateReviewFromVisit(c *gin.Context) {
	visitID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	view, err := h.Reviews.CreateFromVisit(c.Request.Context(), middleware.UserID(c), visitID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, view)
}

func (h *Handler) MyReviews(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.Reviews.ListMine(middleware.UserID(c), c.Query("status"), limit)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"items": items})
}

func (h *Handler) ReviewStatus(c *gin.Context) {
	reviewID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	view, err := h.Reviews.GetOwned(middleware.UserID(c), reviewID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{
		"review_id": view.Review.ID, "publish_status": view.Review.PublishStatus,
		"risk_level": view.Review.RiskLevel, "submitted_at": view.Review.SubmittedAt,
		"published_at": view.Review.PublishedAt,
	})
}
