package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"shijibu/internal/middleware"
	model "shijibu/internal/model/pgsql"
	"shijibu/internal/platform/httpx"
	"shijibu/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateNutritionRecord(c *gin.Context) {
	var input struct {
		MealPeriod        string          `json:"meal_period"`
		EatenAt           string          `json:"eaten_at"`
		Description       string          `json:"description"`
		Foods             json.RawMessage `json:"foods"`
		Calories          *float64        `json:"calories"`
		ProteinGrams      *float64        `json:"protein_grams"`
		FatGrams          *float64        `json:"fat_grams"`
		CarbohydrateGrams *float64        `json:"carbohydrate_grams"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "营养记录格式不正确")
		return
	}
	var eatenAt time.Time
	var err error
	if input.EatenAt != "" {
		eatenAt, err = time.Parse(time.RFC3339, input.EatenAt)
		if err != nil {
			httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "用餐时间格式不正确")
			return
		}
	}
	if len(input.Foods) == 0 {
		input.Foods = json.RawMessage("[]")
	}
	if !json.Valid(input.Foods) {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "食物列表格式不正确")
		return
	}
	item, err := h.Nutrition.Create(middleware.UserID(c), service.NutritionInput{MealPeriod: input.MealPeriod, EatenAt: eatenAt, Description: input.Description, Foods: model.JSONDocument(input.Foods), Calories: input.Calories, ProteinGrams: input.ProteinGrams, FatGrams: input.FatGrams, CarbohydrateGrams: input.CarbohydrateGrams})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, item)
}

func (h *Handler) NutritionRecords(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.Nutrition.List(middleware.UserID(c), c.Query("meal_period"), limit)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"items": items})
}

func (h *Handler) NutritionSummary(c *gin.Context) {
	item, err := h.Nutrition.Summary(middleware.UserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, item)
}

func (h *Handler) DeleteNutritionRecord(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.Nutrition.Delete(middleware.UserID(c), id); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
