package model

import "time"

// City is an operator-managed city that is available to miniapp users.
// Internal publishing fields are intentionally not exposed by the public API.
type City struct {
	Code        string    `gorm:"size:16;primaryKey" json:"code"`
	Name        string    `gorm:"size:64;not null" json:"name"`
	Description string    `gorm:"size:240;not null;default:''" json:"desc"`
	Image       string    `gorm:"size:1024;not null;default:''" json:"image"`
	Enabled     bool      `gorm:"not null;default:true;index" json:"-"`
	IsDefault   bool      `gorm:"not null;default:false" json:"is_default"`
	SortOrder   int       `gorm:"not null;default:0;index" json:"-"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func (City) TableName() string { return "cities" }
