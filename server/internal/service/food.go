package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"nutrilens/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FoodService struct {
	db       *gorm.DB
	DeepSeek *DeepSeekService
}

func NewFoodService(db *gorm.DB, ds *DeepSeekService) *FoodService {
	return &FoodService{db: db, DeepSeek: ds}
}

// DB returns the underlying gorm.DB for direct use (user queries, privacy, etc.).
func (s *FoodService) DB() *gorm.DB {
	return s.db
}

// logAICall records every AI call to the AICallLog table.
func (s *FoodService) logAICall(userID uint, callType, input string, isFood bool, aiResp string) {
	log := model.AICallLog{
		UserID:     userID,
		CallType:   callType,
		Input:      input,
		IsFood:     isFood,
		AIResponse: aiResp,
	}
	s.db.Create(&log)
}

// AnalyzeImage analyzes a food image and saves the record.
func (s *FoodService) AnalyzeImage(ctx context.Context, userID uint, mealType int, imageBytes []byte, imageURL string) (*model.FoodRecord, error) {
	user := s.GetUser(userID)

	result, err := s.DeepSeek.AnalyzeByImage(ctx, imageBytes, user)
	if err != nil {
		return nil, err
	}

	defer func() {
		aiJSON, _ := json.Marshal(result)
		s.logAICall(userID, "analyze_image", imageURL, result.IsFood, string(aiJSON))
	}()

	if !result.IsFood {
		return nil, fmt.Errorf("not_food")
	}

	record := &model.FoodRecord{
		GroupID:      uuid.New().String()[:8],
		UserID:       userID,
		MealType:     mealType,
		FoodName:     result.FoodName,
		Unit:         result.Unit,
		UnitAmount:   result.UnitAmount,
		Calories:     result.Calories,
		ImageURL:     imageURL,
		AISuggestion: result.Suggestion,
		Nutrients:    result.Nutrients,
	}
	if err := s.db.Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

// AnalyzeText analyzes a text description and saves one record per food item.
func (s *FoodService) AnalyzeText(ctx context.Context, userID uint, mealType int, description string) ([]model.FoodRecord, string, error) {
	user := s.GetUser(userID)

	result, err := s.DeepSeek.AnalyzeByText(ctx, description, user)
	if err != nil {
		return nil, "", err
	}

	defer func() {
		aiJSON, _ := json.Marshal(result)
		s.logAICall(userID, "analyze_text", description, result.IsFood, string(aiJSON))
	}()

	if !result.IsFood {
		return nil, "", nil
	}

	// If multi-food result
	if len(result.Foods) > 0 {
		groupID := uuid.New().String()[:8]
		var records []model.FoodRecord
		for _, item := range result.Foods {
			record := model.FoodRecord{
				GroupID:      groupID,
				UserID:       userID,
				MealType:     mealType,
				FoodName:     item.Name,
				Unit:         item.Unit,
				UnitAmount:   item.UnitAmount,
				Calories:     item.Calories,
				Description:  description,
				AISuggestion: result.Suggestion,
				Nutrients:    item.Nutrients,
			}
			if err := s.db.Create(&record).Error; err != nil {
				return records, result.Suggestion, err
			}
			records = append(records, record)
		}
		return records, result.Suggestion, nil
	}

	// Fallback: single food result
	groupID := uuid.New().String()[:8]
	record := model.FoodRecord{
		GroupID:      groupID,
		UserID:       userID,
		MealType:     mealType,
		FoodName:     result.FoodName,
		Unit:         result.Unit,
		UnitAmount:   result.UnitAmount,
		Calories:     result.Calories,
		Description:  description,
		AISuggestion: result.Suggestion,
		Nutrients:    result.Nutrients,
	}
	if err := s.db.Create(&record).Error; err != nil {
		return nil, result.Suggestion, err
	}
	return []model.FoodRecord{record}, result.Suggestion, nil
}

// GetUser returns a user by ID.
func (s *FoodService) GetUser(userID uint) model.User {
	var user model.User
	s.db.First(&user, userID)
	return user
}

// ListRecords returns food records for a user, optionally filtered by date and meal type.
func (s *FoodService) ListRecords(userID uint, date string, mealType int) ([]model.FoodRecord, error) {
	query := s.db.Where("user_id = ?", userID)
	if date != "" {
		start, err := time.Parse("2006-01-02", date)
		if err == nil {
			end := start.Add(24 * time.Hour)
			query = query.Where("created_at >= ? AND created_at < ?", start, end)
		}
	}
	if mealType > 0 {
		query = query.Where("meal_type = ?", mealType)
	}

	var records []model.FoodRecord
	err := query.Order("created_at DESC").Find(&records).Error
	return records, err
}

// GetRecord returns a single food record.
func (s *FoodService) GetRecord(userID uint, id string) (*model.FoodRecord, error) {
	var record model.FoodRecord
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// DeleteRecord deletes a food record.
func (s *FoodService) DeleteRecord(userID uint, id string) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.FoodRecord{}).Error
}

// DailySummary returns the aggregated nutrition data for a given date.
func (s *FoodService) DailySummary(userID uint, date string) (map[string]interface{}, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	start, _ := time.Parse("2006-01-02", date)
	end := start.Add(24 * time.Hour)

	var records []model.FoodRecord
	s.db.Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).Find(&records)

	totalCalories := 0.0
	totalNutrients := model.JSONMap{}
	for _, r := range records {
		totalCalories += r.Calories
		for k, v := range r.Nutrients {
			totalNutrients[k] += v
		}
	}

	return map[string]interface{}{
		"date":            date,
		"total_calories":  totalCalories,
		"total_nutrients": totalNutrients,
		"meal_count":      len(records),
		"records":         records,
	}, nil
}

// MonthlySummary returns aggregated data for a whole month plus per-day breakdown.
func (s *FoodService) MonthlySummary(userID uint, month string) (map[string]interface{}, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, fmt.Errorf("invalid month format, use YYYY-MM")
	}
	end := start.AddDate(0, 1, 0)

	var records []model.FoodRecord
	s.db.Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Order("created_at ASC").Find(&records)

	totalCalories := 0.0
	totalNutrients := model.JSONMap{}
	for _, r := range records {
		totalCalories += r.Calories
		for k, v := range r.Nutrients {
			totalNutrients[k] += v
		}
	}

	daysInMonth := int(end.Sub(start).Hours() / 24)
	daySet := map[string]bool{}
	for _, r := range records {
		daySet[r.CreatedAt.Format("2006-01-02")] = true
	}
	daysWithRecords := len(daySet)

	avgCalories := 0.0
	if daysWithRecords > 0 {
		avgCalories = totalCalories / float64(daysWithRecords)
	}

	return map[string]interface{}{
		"month":             month,
		"total_calories":    totalCalories,
		"avg_calories":      avgCalories,
		"total_nutrients":   totalNutrients,
		"days_in_month":     daysInMonth,
		"days_with_records": daysWithRecords,
		"records":           records,
	}, nil
}
