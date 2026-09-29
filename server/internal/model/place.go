package model

import "time"

type Place struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:160;not null" json:"name"`
	Category     string    `gorm:"size:32;not null" json:"category"`
	CityCode     string    `gorm:"size:16;not null;index" json:"city_code"`
	District     string    `gorm:"size:64" json:"district"`
	BusinessArea string    `gorm:"size:64" json:"business_area"`
	Address      string    `gorm:"size:300;not null" json:"address"`
	Description  string    `gorm:"type:text;not null;default:''" json:"description"`
	Longitude    float64   `gorm:"type:numeric(10,7);not null" json:"longitude"`
	Latitude     float64   `gorm:"type:numeric(10,7);not null" json:"latitude"`
	POIProvider  *string   `gorm:"size:32" json:"poi_provider,omitempty"`
	POIID        *string   `gorm:"size:128" json:"poi_id,omitempty"`
	Status       string    `gorm:"size:24;not null;default:'pending';index" json:"status"`
	MergedIntoID *uint     `gorm:"index" json:"merged_into_id,omitempty"`
	CreatedBy    *uint     `gorm:"index" json:"created_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Place) TableName() string { return "places" }

type Tag struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Code      string `gorm:"size:64;not null;uniqueIndex" json:"code"`
	Name      string `gorm:"size:64;not null" json:"name"`
	GroupName string `gorm:"size:64;not null;index" json:"group_name"`
	Enabled   bool   `gorm:"not null;default:true;index" json:"enabled"`
	SortOrder int    `gorm:"not null;default:0" json:"sort_order"`
}

func (Tag) TableName() string { return "tags" }
