package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"shijibu/internal/middleware"
	"shijibu/internal/model"
	"shijibu/internal/platform/httpx"
	"shijibu/internal/service"

	"github.com/gin-gonic/gin"
)

type recordRequest struct {
	PlaceID       uint            `json:"place_id"`
	PlaceName     string          `json:"place_name"`
	CityCode      string          `json:"city_code"`
	VisitDate     string          `json:"visit_date" binding:"required"`
	ConsumerType  string          `json:"consumer_type"`
	Conclusion    string          `json:"conclusion" binding:"required"`
	PriceMin      *float64        `json:"price_min"`
	PriceMax      *float64        `json:"price_max"`
	AverageCost   *float64        `json:"average_cost"`
	WaitMinutes   *int            `json:"wait_minutes"`
	MealPeriod    string          `json:"meal_period"`
	Dishes        json.RawMessage `json:"dishes"`
	Content       string          `json:"content"`
	Visibility    string          `json:"visibility"`
	TagCodes      []string        `json:"tag_codes"`
	ChangeSummary string          `json:"change_summary"`
}

func (h *Handler) CreateRecord(c *gin.Context) {
	input, ok := bindRecordInput(c)
	if !ok {
		return
	}
	view, err := h.Records.Create(middleware.UserID(c), input, c.GetHeader("Idempotency-Key"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, view)
}

func (h *Handler) GetRecord(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	view, err := h.Records.GetOwned(middleware.UserID(c), id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, view)
}

func (h *Handler) UpdateRecord(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	input, ok := bindRecordInput(c)
	if !ok {
		return
	}
	view, err := h.Records.Update(middleware.UserID(c), id, input)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, view)
}

func (h *Handler) DeleteRecord(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.Records.Delete(middleware.UserID(c), id); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) MyRecords(c *gin.Context) {
	page, err := h.Records.ListMine(middleware.UserID(c), c.Query("status"), c.Query("city_code"), c.Query("cursor"), queryLimit(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, page)
}

func (h *Handler) SubmitRecord(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	view, err := h.Records.SubmitPublic(c.Request.Context(), middleware.UserID(c), id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, view)
}

func (h *Handler) RecordReviewStatus(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	view, err := h.Records.GetOwned(middleware.UserID(c), id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{
		"record_id": view.Record.ID, "publish_status": view.Record.PublishStatus,
		"risk_level": view.Record.RiskLevel, "submitted_at": view.Record.SubmittedAt,
		"published_at": view.Record.PublishedAt,
	})
}

func (h *Handler) FootprintSummary(c *gin.Context) {
	result, err := h.Records.FootprintSummary(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

func (h *Handler) FootprintMap(c *gin.Context) {
	items, err := h.Records.FootprintMap(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"points": items})
}

func bindRecordInput(c *gin.Context) (service.RecordInput, bool) {
	var request recordRequest
	if c.ShouldBindJSON(&request) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "记录内容格式不正确")
		return service.RecordInput{}, false
	}
	visitDate, err := time.Parse("2006-01-02", strings.TrimSpace(request.VisitDate))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "到店日期格式应为 YYYY-MM-DD")
		return service.RecordInput{}, false
	}
	dishes := request.Dishes
	if len(dishes) == 0 {
		dishes = json.RawMessage("[]")
	}
	if !json.Valid(dishes) {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "菜品数据格式不正确")
		return service.RecordInput{}, false
	}
	return service.RecordInput{PlaceID: request.PlaceID, PlaceName: request.PlaceName, CityCode: request.CityCode, VisitDate: visitDate, ConsumerType: request.ConsumerType, Conclusion: request.Conclusion, PriceMin: request.PriceMin, PriceMax: request.PriceMax, AverageCost: request.AverageCost, WaitMinutes: request.WaitMinutes, MealPeriod: request.MealPeriod, Dishes: model.JSONDocument(dishes), Content: request.Content, Visibility: request.Visibility, TagCodes: request.TagCodes, ChangeSummary: request.ChangeSummary}, true
}
