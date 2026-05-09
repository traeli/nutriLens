package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONMap for jsonb fields
type JSONMap map[string]float64

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	OpenID    string         `gorm:"uniqueIndex;size:128;not null" json:"openid"`
	Nickname  string         `gorm:"size:64" json:"nickname"`
	AvatarURL string         `gorm:"size:512" json:"avatar_url"`
	Height    float64        `json:"height"` // cm
	Weight    float64        `json:"weight"` // kg
	Age       int            `json:"age"`
	Gender    int            `json:"gender"`                        // 0=unknown, 1=male, 2=female
	Tag       string         `gorm:"size:32;default:''" json:"tag"` // 用户标签: 空串=普通用户, vip=无限制
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type FoodRecord struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	GroupID      string         `gorm:"size:64;index" json:"group_id"` // 同一次分析的多条记录共享
	MealType     int            `json:"meal_type"`                     // 1=breakfast, 2=lunch, 3=dinner, 4=snack
	FoodName     string         `gorm:"size:256" json:"food_name"`
	Unit         string         `gorm:"size:16;default:'kg'" json:"unit"` // 单位: kg, bowl, piece 等
	UnitAmount   float64        `gorm:"default:0" json:"unit_amount"`     // 单位数量
	Calories     float64        `json:"calories"`
	ImageURL     string         `gorm:"size:512" json:"image_url"`
	Description  string         `gorm:"type:text" json:"description"`
	AISuggestion string         `gorm:"type:text" json:"ai_suggestion"`
	Nutrients    JSONMap        `gorm:"type:jsonb" json:"nutrients"` // protein, carbs, fat, fiber, sugar, vitamin_c, etc.
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Dish struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"size:128;not null" json:"name"`
	Category   string         `gorm:"size:64" json:"category"`
	Calories   float64        `json:"calories"`
	Unit       string         `gorm:"size:16;default:'kg'" json:"unit"`
	UnitAmount float64        `gorm:"default:0" json:"unit_amount"`
	Weight     int            `gorm:"default:50" json:"weight"` // 1-100 喜爱程度
	IsSystem   bool           `gorm:"default:false" json:"is_system"`
	UserID     uint           `gorm:"index" json:"user_id"`
	ImageURL   string         `gorm:"size:512" json:"image_url"`
	Recipe     string         `gorm:"type:text" json:"recipe"`
	Nutrients  JSONMap        `gorm:"type:jsonb" json:"nutrients"` // 缓存AI分析的营养数据
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type PrivacyAgreement struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	AgreementType string    `gorm:"size:64;not null" json:"agreement_type"` // user_info, privacy
	Version       string    `gorm:"size:32" json:"version"`
	AgreedAt      time.Time `json:"agreed_at"`
}

type AICallLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	CallType   string    `gorm:"size:32;not null" json:"call_type"` // analyze_text, analyze_image, analyze_dish
	Input      string    `gorm:"type:text" json:"input"`
	IsFood     bool      `gorm:"default:true" json:"is_food"`
	AIResponse string    `gorm:"type:text" json:"ai_response"`
	CreatedAt  time.Time `json:"created_at"`
}
