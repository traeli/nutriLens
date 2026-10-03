package pgsql

import (
	"time"

	"gorm.io/gorm"
)

// NutritionRecord 保存用户的私人饮食历史，刻意不包含地点、发布、审核或社交互动字段。
type NutritionRecord struct {
	ID                uint           `gorm:"primaryKey;comment:饮食记录编号" json:"id"`
	UserID            uint           `gorm:"not null;index;comment:用户编号" json:"-"`
	MealPeriod        string         `gorm:"size:24;comment:用餐时段" json:"meal_period"`
	EatenAt           time.Time      `gorm:"comment:进食时间" json:"eaten_at"`
	SourceType        string         `gorm:"size:24;not null;comment:记录来源类型" json:"source_type"`
	Description       string         `gorm:"type:text;comment:饮食描述" json:"description"`
	Foods             JSONDocument   `gorm:"type:jsonb;comment:食物明细" json:"foods"`
	Calories          *float64       `gorm:"comment:估算热量（千卡）" json:"calories,omitempty"`
	ProteinGrams      *float64       `gorm:"comment:估算蛋白质（克）" json:"protein_grams,omitempty"`
	FatGrams          *float64       `gorm:"comment:估算脂肪（克）" json:"fat_grams,omitempty"`
	CarbohydrateGrams *float64       `gorm:"comment:估算碳水化合物（克）" json:"carbohydrate_grams,omitempty"`
	Advice            string         `gorm:"type:text;not null;default:'';comment:本餐饮食建议" json:"advice"`
	ImageObjectKey    string         `gorm:"size:512;comment:餐食图片对象存储键" json:"image_url,omitempty"`
	CreatedAt         time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index;comment:软删除时间" json:"-"`
}

func (NutritionRecord) TableName() string { return "nutrition_records" }
