package service

import (
	"strings"
	"time"

	"shijibu/internal/model"

	"gorm.io/gorm"
)

type NutritionService struct{ db *gorm.DB }

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

func NewNutritionService(db *gorm.DB) *NutritionService { return &NutritionService{db: db} }

func (s *NutritionService) Create(userID uint, input NutritionInput) (*model.NutritionRecord, error) {
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
	item := &model.NutritionRecord{
		UserID: userID, MealPeriod: input.MealPeriod, EatenAt: input.EatenAt, SourceType: "manual",
		Description: input.Description, Foods: input.Foods, Calories: input.Calories, ProteinGrams: input.ProteinGrams,
		FatGrams: input.FatGrams, CarbohydrateGrams: input.CarbohydrateGrams,
	}
	return item, s.db.Create(item).Error
}

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
