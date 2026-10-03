package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	model "shijibu/internal/model/pgsql"
	"shijibu/internal/platform/llm"

	"gorm.io/gorm"
)

var ErrNutritionAnalysisNotConfigured = errors.New("nutrition analysis is not configured")

type NutritionAnalyzer interface {
	AnalyzeNutrition(ctx context.Context, description string) (string, error)
}

type NutritionService struct {
	db       *gorm.DB
	analyzer NutritionAnalyzer
}

type nutritionAnalysis struct {
	Foods             []nutritionFood `json:"foods"`
	Calories          float64         `json:"calories"`
	ProteinGrams      float64         `json:"protein_grams"`
	FatGrams          float64         `json:"fat_grams"`
	CarbohydrateGrams float64         `json:"carbohydrate_grams"`
	Advice            string          `json:"advice"`
}

type nutritionFood struct {
	Name              string  `json:"name"`
	Amount            string  `json:"amount"`
	Calories          float64 `json:"calories"`
	ProteinGrams      float64 `json:"protein_grams"`
	FatGrams          float64 `json:"fat_grams"`
	CarbohydrateGrams float64 `json:"carbohydrate_grams"`
}

type NutritionInput struct {
	MealPeriod        string
	EatenAt           time.Time
	Description       string
	Foods             model.JSONDocument
	Calories          *float64
	ProteinGrams      *float64
	FatGrams          *float64
	CarbohydrateGrams *float64
}

type NutritionSummary struct {
	MealCount         int64   `json:"meal_count"`
	Calories          float64 `json:"calories"`
	ProteinGrams      float64 `json:"protein_grams"`
	FatGrams          float64 `json:"fat_grams"`
	CarbohydrateGrams float64 `json:"carbohydrate_grams"`
}

func NewNutritionService(db *gorm.DB, analyzer NutritionAnalyzer) *NutritionService {
	return &NutritionService{db: db, analyzer: analyzer}
}

func (s *NutritionService) Create(ctx context.Context, userID uint, input NutritionInput) (*model.NutritionRecord, error) {
	input.MealPeriod, input.Description = strings.TrimSpace(input.MealPeriod), strings.TrimSpace(input.Description)
	if input.EatenAt.IsZero() {
		input.EatenAt = time.Now()
	}
	if len([]rune(input.Description)) > 1000 || !validMealPeriod(input.MealPeriod) || invalidNutritionNumber(input.Calories) ||
		invalidNutritionNumber(input.ProteinGrams) || invalidNutritionNumber(input.FatGrams) || invalidNutritionNumber(input.CarbohydrateGrams) {
		return nil, ErrInvalidInput
	}
	if len(input.Foods) == 0 {
		input.Foods = model.JSONDocument("[]")
	}
	sourceType := "manual"
	advice := ""
	if shouldAnalyzeNutrition(input) {
		if input.Description == "" {
			return nil, ErrInvalidInput
		}
		analysis, err := s.analyze(ctx, input.Description)
		if err != nil {
			return nil, err
		}
		foods, err := json.Marshal(analysis.Foods)
		if err != nil {
			return nil, fmt.Errorf("encode nutrition foods: %w", err)
		}
		input.Foods = model.JSONDocument(foods)
		input.Calories = float64Pointer(analysis.Calories)
		input.ProteinGrams = float64Pointer(analysis.ProteinGrams)
		input.FatGrams = float64Pointer(analysis.FatGrams)
		input.CarbohydrateGrams = float64Pointer(analysis.CarbohydrateGrams)
		advice = strings.TrimSpace(analysis.Advice)
		sourceType = "llm_text"
	}
	item := &model.NutritionRecord{
		UserID: userID, MealPeriod: input.MealPeriod, EatenAt: input.EatenAt, SourceType: sourceType,
		Description: input.Description, Foods: input.Foods, Calories: input.Calories, ProteinGrams: input.ProteinGrams,
		FatGrams: input.FatGrams, CarbohydrateGrams: input.CarbohydrateGrams, Advice: advice,
	}
	return item, s.db.Create(item).Error
}

func (s *NutritionService) analyze(ctx context.Context, description string) (*nutritionAnalysis, error) {
	if s == nil || s.analyzer == nil {
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, ErrNutritionAnalysisNotConfigured)
	}
	content, err := s.analyzer.AnalyzeNutrition(ctx, description)
	if errors.Is(err, llm.ErrNotConfigured) {
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, ErrNutritionAnalysisNotConfigured)
	}
	if err != nil {
		return nil, fmt.Errorf("analyze nutrition: %w: %w", ErrUnavailable, err)
	}
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	var result nutritionAnalysis
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil || decoder.Decode(&struct{}{}) != io.EOF || !validNutritionAnalysis(result) {
		return nil, fmt.Errorf("invalid nutrition model result: %w", ErrUnavailable)
	}
	return &result, nil
}

func shouldAnalyzeNutrition(input NutritionInput) bool {
	return input.Calories == nil && input.ProteinGrams == nil && input.FatGrams == nil && input.CarbohydrateGrams == nil
}

func validNutritionAnalysis(value nutritionAnalysis) bool {
	if len(value.Foods) == 0 || invalidNutritionValue(value.Calories) || invalidNutritionValue(value.ProteinGrams) ||
		invalidNutritionValue(value.FatGrams) || invalidNutritionValue(value.CarbohydrateGrams) ||
		strings.TrimSpace(value.Advice) == "" || len([]rune(strings.TrimSpace(value.Advice))) > 80 {
		return false
	}
	for _, food := range value.Foods {
		if strings.TrimSpace(food.Name) == "" || strings.TrimSpace(food.Amount) == "" ||
			invalidNutritionValue(food.Calories) || invalidNutritionValue(food.ProteinGrams) ||
			invalidNutritionValue(food.FatGrams) || invalidNutritionValue(food.CarbohydrateGrams) {
			return false
		}
	}
	return true
}

func invalidNutritionValue(value float64) bool { return value < 0 || value > 100000 }

func float64Pointer(value float64) *float64 { return &value }

func (s *NutritionService) List(userID uint, mealPeriod string, limit int) ([]model.NutritionRecord, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	db := s.db.Where("user_id = ?", userID)
	if mealPeriod = strings.TrimSpace(mealPeriod); mealPeriod != "" {
		if !validMealPeriod(mealPeriod) {
			return nil, ErrInvalidInput
		}
		db = db.Where("meal_period = ?", mealPeriod)
	}
	var items []model.NutritionRecord
	return items, db.Order("eaten_at DESC, id DESC").Limit(limit).Find(&items).Error
}

func (s *NutritionService) Delete(userID, id uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.NutritionRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *NutritionService) Summary(userID uint) (*NutritionSummary, error) {
	var summary NutritionSummary
	err := s.db.Model(&model.NutritionRecord{}).Where("user_id = ? AND deleted_at IS NULL", userID).
		Select("COUNT(*) AS meal_count, COALESCE(SUM(calories), 0) AS calories, COALESCE(SUM(protein_grams), 0) AS protein_grams, COALESCE(SUM(fat_grams), 0) AS fat_grams, COALESCE(SUM(carbohydrate_grams), 0) AS carbohydrate_grams").
		Scan(&summary).Error
	return &summary, err
}

func validMealPeriod(value string) bool {
	return value == "" || value == "breakfast" || value == "lunch" || value == "dinner" || value == "snack"
}

func invalidNutritionNumber(value *float64) bool {
	return value != nil && (*value < 0 || *value > 100000)
}
