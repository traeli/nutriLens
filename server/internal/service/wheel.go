package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"nutrilens/internal/model"

	"gorm.io/gorm"
)

type WheelService struct {
	DB       *gorm.DB
	DeepSeek *DeepSeekService
}

func NewWheelService(db *gorm.DB, ds *DeepSeekService) *WheelService {
	return &WheelService{DB: db, DeepSeek: ds}
}

// ListDishes returns dishes visible to a user (system + user's own), optionally filtered by category.
func (s *WheelService) ListDishes(userID uint, category string) ([]model.Dish, error) {
	var dishes []model.Dish
	query := s.DB.Where("is_system = ? OR user_id = ?", true, userID)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Order("is_system DESC, weight DESC").Find(&dishes).Error
	return dishes, err
}

// Spin performs a weighted random selection from visible dishes.
func (s *WheelService) Spin(userID uint, category string) (*model.Dish, int, error) {
	dishes, err := s.ListDishes(userID, category)
	if err != nil {
		return nil, 0, err
	}
	if len(dishes) == 0 {
		return nil, 0, nil
	}

	totalWeight := 0
	for _, d := range dishes {
		totalWeight += d.Weight
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	target := r.Intn(totalWeight)
	cumWeight := 0
	for i := range dishes {
		cumWeight += dishes[i].Weight
		if target < cumWeight {
			return &dishes[i], len(dishes), nil
		}
	}
	return &dishes[len(dishes)-1], len(dishes), nil
}

// CreateDish adds a user-defined dish.
func (s *WheelService) CreateDish(dish *model.Dish) error {
	return s.DB.Create(dish).Error
}

// CreateDishWithAI analyzes a dish name via AI, fills in nutritional data, and saves.
// If a dish with the same name already exists for this user, reuse its cached data instead of calling AI.
func (s *WheelService) CreateDishWithAI(ctx context.Context, userID uint, name string, weight int) (*model.Dish, error) {
	if weight <= 0 {
		weight = 50
	}

	// Check if user already has a dish with this name
	var existing model.Dish
	if err := s.DB.Where("user_id = ? AND name = ? AND is_system = false", userID, name).First(&existing).Error; err == nil {
		// Already exists, update weight only
		existing.Weight = weight
		s.DB.Save(&existing)
		return &existing, nil
	}

	// Check if a system dish with this name exists (to reuse cached data)
	if err := s.DB.Where("name = ? AND is_system = true", name).First(&existing).Error; err == nil {
		dish := &model.Dish{
			Name:       name,
			Category:   existing.Category,
			Calories:   existing.Calories,
			Unit:       existing.Unit,
			UnitAmount: existing.UnitAmount,
			Weight:     weight,
			IsSystem:   false,
			UserID:     userID,
			Nutrients:  existing.Nutrients,
		}
		if err := s.DB.Create(dish).Error; err != nil {
			return nil, err
		}
		return dish, nil
	}

	// No cached data — call AI
	result, err := s.DeepSeek.AnalyzeDish(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// Log AI call
	aiJSON, _ := json.Marshal(result)
	s.DB.Create(&model.AICallLog{
		UserID:     userID,
		CallType:   "analyze_dish",
		Input:      name,
		IsFood:     result.IsFood,
		AIResponse: string(aiJSON),
	})

	if !result.IsFood {
		return nil, fmt.Errorf("not_food")
	}

	dish := &model.Dish{
		Name:       name,
		Category:   result.Category,
		Calories:   result.Calories,
		Unit:       result.Unit,
		UnitAmount: result.UnitAmount,
		Weight:     weight,
		IsSystem:   false,
		UserID:     userID,
		Nutrients:  result.Nutrients,
	}
	if err := s.DB.Create(dish).Error; err != nil {
		return nil, err
	}
	return dish, nil
}

// DeleteDish deletes a user-defined dish (cannot delete system dishes).
func (s *WheelService) DeleteDish(userID uint, id string) error {
	return s.DB.Where("id = ? AND user_id = ? AND is_system = ?", id, userID, false).Delete(&model.Dish{}).Error
}
