package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"nutrilens/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FoodService struct {
	db        *gorm.DB
	AI        *AIProviderService
	cosClient COSObjectGetter
}

// COSObjectGetter downloads objects from COS.
type COSObjectGetter interface {
	GetObject(ctx context.Context, objectKey string) ([]byte, error)
}

func NewFoodService(db *gorm.DB, ai *AIProviderService, cosClient COSObjectGetter) *FoodService {
	return &FoodService{db: db, AI: ai, cosClient: cosClient}
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

// AnalyzeImage analyzes a food image using a two-step flow:
// 1) Vision model identifies food items and quantities
// 2) DeepSeek (text analysis model) provides nutritional analysis and health advice
func (s *FoodService) AnalyzeImage(ctx context.Context, userID uint, mealType int, imageURL string) ([]model.FoodRecord, string, error) {
	user := s.GetUser(userID)

	// Step 1: Identify food items from image via vision model
	identifyResult, err := s.AI.IdentifyImageFood(ctx, imageURL)
	if err != nil {
		return nil, "", err
	}

	log.Printf("[AnalyzeImage] identify result: is_food=%v, foods=%d", identifyResult.IsFood, len(identifyResult.Foods))

	defer func() {
		aiJSON, _ := json.Marshal(identifyResult)
		s.logAICall(userID, "identify_image", imageURL, identifyResult.IsFood, string(aiJSON))
	}()

	if !identifyResult.IsFood || len(identifyResult.Foods) == 0 {
		return nil, "", fmt.Errorf("not_food")
	}

	// Build description from identified foods
	var parts []string
	for _, item := range identifyResult.Foods {
		parts = append(parts, fmt.Sprintf("%s %.2fkg", item.Name, item.UnitAmount))
	}
	description := "我吃了: " + strings.Join(parts, ", ")

	// Step 2: Analyze via text analysis model (DeepSeek) for nutrition + health advice
	result, err := s.AI.AnalyzeByText(ctx, description, user)
	if err != nil {
		return nil, "", err
	}

	defer func() {
		aiJSON, _ := json.Marshal(result)
		s.logAICall(userID, "analyze_image", description, result.IsFood, string(aiJSON))
	}()

	if !result.IsFood {
		return nil, "", fmt.Errorf("not_food")
	}

	// Create FoodRecords (same logic as AnalyzeText)
	groupID := uuid.New().String()[:8]
	var records []model.FoodRecord

	if len(result.Foods) > 0 {
		for _, item := range result.Foods {
			record := model.FoodRecord{
				GroupID:      groupID,
				UserID:       userID,
				MealType:     mealType,
				FoodName:     item.Name,
				Unit:         item.Unit,
				UnitAmount:   item.UnitAmount,
				Calories:     item.Calories,
				ImageURL:     imageURL,
				AISuggestion: result.Suggestion,
				Nutrients:    item.Nutrients,
			}
			if err := s.db.Create(&record).Error; err != nil {
				return records, result.Suggestion, err
			}
			records = append(records, record)
		}
	} else {
		// Fallback: single food result
		record := model.FoodRecord{
			GroupID:      groupID,
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
		if err := s.db.Create(&record).Error; err != nil {
			return nil, result.Suggestion, err
		}
		records = append(records, record)
	}

	return records, result.Suggestion, nil
}

// AnalyzeText analyzes a text description and saves one record per food item.
func (s *FoodService) AnalyzeText(ctx context.Context, userID uint, mealType int, description string) ([]model.FoodRecord, string, error) {
	user := s.GetUser(userID)

	result, err := s.AI.AnalyzeByText(ctx, description, user)
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

// GetDailyAnalysis returns AI-generated daily health suggestion with caching.
// Cache is invalidated when record count or total calories change.
func (s *FoodService) GetDailyAnalysis(ctx context.Context, userID uint, date string) (map[string]interface{}, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	start, _ := time.Parse("2006-01-02", date)
	end := start.Add(24 * time.Hour)

	// Query current day's records
	var records []model.FoodRecord
	s.db.Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).Find(&records)

	currentCount := len(records)
	currentCalories := 0.0
	currentNutrients := model.JSONMap{}
	for _, r := range records {
		currentCalories += r.Calories
		for k, v := range r.Nutrients {
			currentNutrients[k] += v
		}
	}

	// No records: return default message without AI call
	if currentCount == 0 {
		return map[string]interface{}{
			"date":            date,
			"record_count":    0,
			"total_calories":  0,
			"total_nutrients": nil,
			"analysis_text":   "还没有记录食物哦，快去记录今天的饮食吧！记录后我会为你生成个性化健康建议。",
			"is_cached":       false,
		}, nil
	}

	// Check cache
	dateOnly, _ := time.Parse("2006-01-02", date)
	var cached model.UserDailyAnalysis
	cacheHit := s.db.Where("user_id = ? AND date = ?", userID, dateOnly).First(&cached).Error == nil

	if cacheHit && cached.RecordCount == currentCount && cached.TotalCalories == currentCalories {
		return map[string]interface{}{
			"date":            date,
			"record_count":    cached.RecordCount,
			"total_calories":  cached.TotalCalories,
			"total_nutrients": cached.TotalNutrients,
			"analysis_text":   cached.AnalysisText,
			"is_cached":       true,
		}, nil
	}

	// Cache miss: call AI
	user := s.GetUser(userID)

	// Build food list
	var foodItems []string
	for _, r := range records {
		foodItems = append(foodItems, fmt.Sprintf("%s (%.0fg, %.0fkcal)", r.FoodName, r.UnitAmount*1000, r.Calories))
	}
	foodList := strings.Join(foodItems, "\n")

	// Build nutrient summary
	nutrientSummary := fmt.Sprintf(
		"总卡路里: %.0fkcal\n蛋白质: %.1fg\n碳水化合物: %.1fg\n脂肪: %.1fg\n膳食纤维: %.1fg\n糖分: %.1fg",
		currentCalories,
		currentNutrients["protein"],
		currentNutrients["carbs"],
		currentNutrients["fat"],
		currentNutrients["fiber"],
		currentNutrients["sugar"],
	)

	analysisText, err := s.AI.AnalyzeDailySummary(ctx, foodList, nutrientSummary, user)
	if err != nil {
		// Return error but don't cache it
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	// Upsert cache
	newAnalysis := model.UserDailyAnalysis{
		UserID:         userID,
		Date:           dateOnly,
		RecordCount:    currentCount,
		TotalCalories:  currentCalories,
		TotalNutrients: currentNutrients,
		AnalysisText:   analysisText,
	}

	if cacheHit {
		s.db.Model(&model.UserDailyAnalysis{}).Where("user_id = ? AND date = ?", userID, dateOnly).Updates(map[string]interface{}{
			"record_count":    currentCount,
			"total_calories":  currentCalories,
			"total_nutrients": currentNutrients,
			"analysis_text":   analysisText,
		})
	} else {
		s.db.Create(&newAnalysis)
	}

	return map[string]interface{}{
		"date":            date,
		"record_count":    currentCount,
		"total_calories":  currentCalories,
		"total_nutrients": currentNutrients,
		"analysis_text":   analysisText,
		"is_cached":       false,
	}, nil
}
