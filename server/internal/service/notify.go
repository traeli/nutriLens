package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"nutrilens/internal/model"

	"gorm.io/gorm"
)

type NotifyConfigProvider interface {
	GetTemplateID() string
	GetDefaultBreakfastTime() string
	GetDefaultLunchTime() string
	GetDefaultDinnerTime() string
}

type NotifyService struct {
	db     *gorm.DB
	wechat *WechatService
	ai     *AIProviderService
	cfg    NotifyConfigProvider
}

func NewNotifyService(db *gorm.DB, wechat *WechatService, ai *AIProviderService, cfg NotifyConfigProvider) *NotifyService {
	return &NotifyService{db: db, wechat: wechat, ai: ai, cfg: cfg}
}

var cstLoc = time.FixedZone("CST", 8*3600)

func mealTypeName(m int) string {
	switch m {
	case 1:
		return "早餐"
	case 2:
		return "午餐"
	case 3:
		return "晚餐"
	default:
		return ""
	}
}

// ==================== User Settings ====================

func (s *NotifyService) GetOrCreateSettings(userID uint) (*model.NotifySetting, error) {
	var setting model.NotifySetting
	if err := s.db.Where("user_id = ?", userID).First(&setting).Error; err == nil {
		return &setting, nil
	}

	// Create with defaults from config
	setting = model.NotifySetting{
		UserID:           userID,
		BreakfastEnabled: true,
		BreakfastTime:    s.cfg.GetDefaultBreakfastTime(),
		LunchEnabled:     true,
		LunchTime:        s.cfg.GetDefaultLunchTime(),
		DinnerEnabled:    true,
		DinnerTime:       s.cfg.GetDefaultDinnerTime(),
	}
	if s.cfg.GetDefaultBreakfastTime() == "" {
		setting.BreakfastTime = "08:00"
	}
	if s.cfg.GetDefaultLunchTime() == "" {
		setting.LunchTime = "12:00"
	}
	if s.cfg.GetDefaultDinnerTime() == "" {
		setting.DinnerTime = "18:00"
	}

	if err := s.db.Create(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

type UpdateNotifySettingsReq struct {
	BreakfastEnabled *bool   `json:"breakfast_enabled"`
	BreakfastTime    *string `json:"breakfast_time"`
	LunchEnabled     *bool   `json:"lunch_enabled"`
	LunchTime        *string `json:"lunch_time"`
	DinnerEnabled    *bool   `json:"dinner_enabled"`
	DinnerTime       *string `json:"dinner_time"`
}

func (s *NotifyService) UpdateSettings(userID uint, req UpdateNotifySettingsReq) error {
	updates := map[string]interface{}{}

	if req.BreakfastEnabled != nil {
		updates["breakfast_enabled"] = *req.BreakfastEnabled
	}
	if req.BreakfastTime != nil {
		if !isValidTime(*req.BreakfastTime) {
			return fmt.Errorf("invalid time format: %s", *req.BreakfastTime)
		}
		updates["breakfast_time"] = *req.BreakfastTime
	}
	if req.LunchEnabled != nil {
		updates["lunch_enabled"] = *req.LunchEnabled
	}
	if req.LunchTime != nil {
		if !isValidTime(*req.LunchTime) {
			return fmt.Errorf("invalid time format: %s", *req.LunchTime)
		}
		updates["lunch_time"] = *req.LunchTime
	}
	if req.DinnerEnabled != nil {
		updates["dinner_enabled"] = *req.DinnerEnabled
	}
	if req.DinnerTime != nil {
		if !isValidTime(*req.DinnerTime) {
			return fmt.Errorf("invalid time format: %s", *req.DinnerTime)
		}
		updates["dinner_time"] = *req.DinnerTime
	}

	if len(updates) == 0 {
		return nil
	}

	return s.db.Model(&model.NotifySetting{}).Where("user_id = ?", userID).Updates(updates).Error
}

func isValidTime(t string) bool {
	_, err := time.Parse("15:04", t)
	return err == nil
}

// ==================== Scheduled Check ====================

// CheckAndNotify is called by the cron scheduler every minute.
func (s *NotifyService) CheckAndNotify(ctx context.Context) error {
	now := time.Now().In(cstLoc)
	currentTime := now.Format("15:04")

	templateID := s.cfg.GetTemplateID()
	if templateID == "" {
		// Template not configured yet, skip silently
		return nil
	}

	var settings []model.NotifySetting
	if err := s.db.Where(
		"(breakfast_enabled = true AND breakfast_time = ?) OR (lunch_enabled = true AND lunch_time = ?) OR (dinner_enabled = true AND dinner_time = ?)",
		currentTime, currentTime, currentTime,
	).Find(&settings).Error; err != nil {
		return fmt.Errorf("query notify settings: %w", err)
	}

	if len(settings) == 0 {
		return nil
	}

	log.Printf("[Notify] %s: found %d users to notify", currentTime, len(settings))

	// Build list of (userID, mealType) pairs
	type notifyJob struct {
		userID   uint
		mealType int
	}

	var jobs []notifyJob
	for _, setting := range settings {
		if setting.BreakfastEnabled && setting.BreakfastTime == currentTime {
			jobs = append(jobs, notifyJob{setting.UserID, 1})
		}
		if setting.LunchEnabled && setting.LunchTime == currentTime {
			jobs = append(jobs, notifyJob{setting.UserID, 2})
		}
		if setting.DinnerEnabled && setting.DinnerTime == currentTime {
			jobs = append(jobs, notifyJob{setting.UserID, 3})
		}
	}

	// Process each job concurrently
	for _, job := range jobs {
		go s.notifyUser(ctx, job.userID, job.mealType, templateID, now)
	}

	return nil
}

// notifyUser handles the full push flow for a single user + meal type.
func (s *NotifyService) notifyUser(ctx context.Context, userID uint, mealType int, templateID string, now time.Time) {
	today := now.Format("2006-01-02")

	// 1. Idempotency guard: check if already sent today
	var existingLog model.NotifyLog
	if err := s.db.Where("user_id = ? AND meal_type = ? AND success = true AND created_at >= ? AND created_at < ?",
		userID, mealType, today+" 00:00:00", today+" 23:59:59",
	).First(&existingLog).Error; err == nil {
		// Already sent successfully today
		return
	}

	// 2. Check if user already has food records for this meal today
	startOfDay, _ := time.ParseInLocation("2006-01-02", today, cstLoc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var foodCount int64
	s.db.Model(&model.FoodRecord{}).Where(
		"user_id = ? AND meal_type = ? AND created_at >= ? AND created_at < ?",
		userID, mealType, startOfDay, endOfDay,
	).Count(&foodCount)

	if foodCount > 0 {
		// User already logged food for this meal, skip
		return
	}

	// 3. Get user info
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		log.Printf("[Notify] user %d not found: %v", userID, err)
		return
	}

	// 4. Get or generate recommendation
	recommendation, err := s.getOrCreateRecommendation(ctx, userID, user, mealType, today, startOfDay, endOfDay)
	if err != nil {
		log.Printf("[Notify] recommendation failed for user %d: %v", userID, err)
		// Fallback: simple message
		recommendation = fmt.Sprintf(`{"recommendations":[{"name":"请打开饭友记查看推荐","reason":"AI推荐暂时不可用","calories":0}],"tip":"记得按时吃饭哦"}`)
	}

	// 5. Build template message data
	templateData := s.buildTemplateData(mealType, recommendation)

	// 6. Send subscribe message
	err = s.wechat.SendSubscribeMessage(ctx, user.OpenID, templateID, "pages/home/home", templateData)

	// 7. Log result
	logEntry := model.NotifyLog{
		UserID:   userID,
		MealType: mealType,
		Content:  recommendation,
		Success:  err == nil,
	}
	if err != nil {
		logEntry.ErrorMsg = err.Error()
		log.Printf("[Notify] send failed for user %d (meal=%d): %v", userID, mealType, err)
	} else {
		log.Printf("[Notify] sent successfully for user %d (meal=%d)", userID, mealType)
	}
	s.db.Create(&logEntry)
}

// getOrCreateRecommendation returns cached or newly generated AI recommendation.
func (s *NotifyService) getOrCreateRecommendation(ctx context.Context, userID uint, user model.User, mealType int, today string, startOfDay, endOfDay time.Time) (string, error) {
	// Check cache first
	var cached model.MealRecommendation
	if err := s.db.Where("user_id = ? AND date = ? AND meal_type = ?", userID, today, mealType).First(&cached).Error; err == nil {
		return cached.Recipient, nil
	}

	// Build context for AI
	userInfo := fmt.Sprintf("身高%.0fcm, 体重%.0fkg, 年龄%d岁, 性别%s",
		user.Height, user.Weight, user.Age, genderStr(user.Gender))

	// Today's food records
	var records []model.FoodRecord
	s.db.Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfDay, endOfDay).
		Order("created_at ASC").Find(&records)

	todayFoods := "暂无记录"
	nutrientSummary := "暂无数据"
	totalCalories := 0.0
	totalNutrients := model.JSONMap{}

	if len(records) > 0 {
		var foodItems []string
		for _, r := range records {
			foodItems = append(foodItems, fmt.Sprintf("- %s (%s) %.0fkcal", r.FoodName, mealTypeName(r.MealType), r.Calories))
			totalCalories += r.Calories
			for k, v := range r.Nutrients {
				totalNutrients[k] += v
			}
		}
		todayFoods = strings.Join(foodItems, "\n")
		nutrientSummary = fmt.Sprintf(
			"总卡路里: %.0fkcal\n蛋白质: %.1fg\n碳水: %.1fg\n脂肪: %.1fg\n膳食纤维: %.1fg\n糖分: %.1fg",
			totalCalories,
			totalNutrients["protein"],
			totalNutrients["carbs"],
			totalNutrients["fat"],
			totalNutrients["fiber"],
			totalNutrients["sugar"],
		)
	}

	// User's preferred dishes (top 10 by weight)
	var dishes []model.Dish
	s.db.Where("is_system = true OR user_id = ?", userID).
		Order("weight DESC").Limit(10).Find(&dishes)

	preferredDishes := "暂无偏好菜品"
	if len(dishes) > 0 {
		var dishItems []string
		for _, d := range dishes {
			dishItems = append(dishItems, fmt.Sprintf("- %s (%s, %.0fkcal, 喜爱度:%d)", d.Name, d.Category, d.Calories, d.Weight))
		}
		preferredDishes = strings.Join(dishItems, "\n")
	}

	// Call AI
	content, err := s.ai.RecommendMeal(ctx, userInfo, todayFoods, nutrientSummary, preferredDishes, mealTypeName(mealType))
	if err != nil {
		return "", err
	}

	// Cache the result
	cache := model.MealRecommendation{
		UserID:    userID,
		Date:      today,
		MealType:  mealType,
		Recipient: content,
	}
	s.db.Create(&cache)

	return content, nil
}

// buildTemplateData maps AI recommendation JSON to WeChat template message keyword slots.
func (s *NotifyService) buildTemplateData(mealType int, aiJSON string) map[string]map[string]string {
	// WeChat template keyword format: {"keyword": {"value": "..."}}
	data := map[string]map[string]string{
		"thing1": {"value": mealTypeName(mealType) + "提醒"},
	}

	// Try to parse AI recommendation for thing2 and thing3
	var rec struct {
		Recommendations []struct {
			Name   string `json:"name"`
			Reason string `json:"reason"`
		} `json:"recommendations"`
		Tip string `json:"tip"`
	}

	if err := json.Unmarshal([]byte(aiJSON), &rec); err == nil {
		if len(rec.Recommendations) > 0 {
			names := make([]string, 0, len(rec.Recommendations))
			for _, r := range rec.Recommendations {
				if len(r.Name) > 0 {
					names = append(names, r.Name)
				}
			}
			data["thing2"] = map[string]string{"value": "推荐: " + strings.Join(names, ", ")}
		}
		if rec.Tip != "" {
			data["thing3"] = map[string]string{"value": rec.Tip}
		}
	}

	// Fallback if parse failed
	if _, ok := data["thing2"]; !ok {
		data["thing2"] = map[string]string{"value": "打开饭友记查看今日推荐"}
	}
	if _, ok := data["thing3"]; !ok {
		data["thing3"] = map[string]string{"value": "按时吃饭，营养均衡"}
	}
	data["time4"] = map[string]string{"value": time.Now().In(cstLoc).Format("15:04")}

	return data
}
