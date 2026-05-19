package model

import "time"

// NotifySetting stores per-user push notification preferences.
type NotifySetting struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserID           uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	BreakfastEnabled bool      `gorm:"default:true" json:"breakfast_enabled"`
	BreakfastTime    string    `gorm:"size:5;default:'08:00'" json:"breakfast_time"`
	LunchEnabled     bool      `gorm:"default:true" json:"lunch_enabled"`
	LunchTime        string    `gorm:"size:5;default:'12:00'" json:"lunch_time"`
	DinnerEnabled    bool      `gorm:"default:true" json:"dinner_enabled"`
	DinnerTime       string    `gorm:"size:5;default:'18:00'" json:"dinner_time"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// NotifyLog records each push notification attempt (idempotency guard).
type NotifyLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	MealType  int       `json:"meal_type"` // 1=breakfast, 2=lunch, 3=dinner
	Content   string    `gorm:"type:text" json:"content"`
	Success   bool      `json:"success"`
	ErrorMsg  string    `gorm:"type:text" json:"error_msg"`
	CreatedAt time.Time `json:"created_at"`
}

// MealRecommendation caches AI-generated meal recommendations per (user, date, meal_type).
type MealRecommendation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_date_meal;not null" json:"user_id"`
	Date      string    `gorm:"size:10;uniqueIndex:idx_user_date_meal;not null" json:"date"` // 2006-01-02
	MealType  int       `gorm:"uniqueIndex:idx_user_date_meal;not null" json:"meal_type"`
	Recipient string    `gorm:"type:text" json:"recipient"` // JSON with dish recommendations
	CreatedAt time.Time `json:"created_at"`
}
