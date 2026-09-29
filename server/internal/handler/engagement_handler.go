package handler

import (
	"net/http"

	"shijibu/internal/middleware"
	"shijibu/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

func (h *Handler) FavoritePlace(c *gin.Context) { h.simplePlaceAction(c, h.Engagement.FavoritePlace) }
func (h *Handler) UnfavoritePlace(c *gin.Context) {
	h.simplePlaceAction(c, h.Engagement.UnfavoritePlace)
}

func (h *Handler) simplePlaceAction(c *gin.Context, action func(uint, uint) error) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := action(middleware.UserID(c), id); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) FavoritePlaces(c *gin.Context) {
	items, err := h.Engagement.FavoritePlaces(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"items": items})
}

func (h *Handler) Helpful(c *gin.Context)   { h.simpleExperienceAction(c, h.Engagement.Helpful) }
func (h *Handler) Unhelpful(c *gin.Context) { h.simpleExperienceAction(c, h.Engagement.Unhelpful) }

func (h *Handler) simpleExperienceAction(c *gin.Context, action func(uint, uint) error) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := action(middleware.UserID(c), id); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) MarkOutdated(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		Reason string `json:"reason"`
	}
	if c.ShouldBindJSON(&input) != nil {
		input.Reason = ""
	}
	if err := h.Engagement.MarkOutdated(middleware.UserID(c), id, input.Reason); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateReport(c *gin.Context) {
	var input struct {
		TargetType  string `json:"target_type" binding:"required"`
		TargetID    uint   `json:"target_id" binding:"required"`
		ReasonCode  string `json:"reason_code" binding:"required"`
		Description string `json:"description"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "举报内容格式不正确")
		return
	}
	item, err := h.Engagement.CreateReport(middleware.UserID(c), input.TargetType, input.TargetID, input.ReasonCode, input.Description)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, item)
}

func (h *Handler) MyReports(c *gin.Context) {
	items, err := h.Engagement.Reports(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"items": items})
}

func (h *Handler) CreateAppeal(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		Text string `json:"text" binding:"required"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "申诉内容不能为空")
		return
	}
	item, err := h.Engagement.CreateAppeal(middleware.UserID(c), id, input.Text)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, item)
}

func (h *Handler) MyAppeals(c *gin.Context) {
	items, err := h.Engagement.Appeals(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"items": items})
}

func (h *Handler) FavoriteRoute(c *gin.Context) { h.simpleRouteAction(c, h.Engagement.FavoriteRoute) }
func (h *Handler) UnfavoriteRoute(c *gin.Context) {
	h.simpleRouteAction(c, h.Engagement.UnfavoriteRoute)
}
func (h *Handler) simpleRouteAction(c *gin.Context, action func(uint, uint) error) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := action(middleware.UserID(c), id); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) StartRoute(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	item, err := h.Engagement.StartRoute(middleware.UserID(c), id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, item)
}

func (h *Handler) UpdateJourney(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		CompletedStopIDs []uint `json:"completed_stop_ids"`
		Complete         bool   `json:"complete"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "路线进度格式不正确")
		return
	}
	item, err := h.Engagement.UpdateJourney(middleware.UserID(c), id, input.CompletedStopIDs, input.Complete)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, item)
}

func (h *Handler) MyJourneys(c *gin.Context) {
	items, err := h.Engagement.MyJourneys(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"items": items})
}
