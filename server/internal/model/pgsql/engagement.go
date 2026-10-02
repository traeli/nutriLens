package pgsql

import (
	"time"

	"gorm.io/gorm"
)

type PlaceFavorite struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	PlaceID   uint      `gorm:"primaryKey" json:"place_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (PlaceFavorite) TableName() string { return "place_favorites" }

type HelpfulVote struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	RecordID  uint      `gorm:"primaryKey" json:"record_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (HelpfulVote) TableName() string { return "helpful_votes" }

type OutdatedSignal struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"user_id"`
	RecordID  uint       `gorm:"not null;index" json:"record_id"`
	Reason    string     `gorm:"type:text" json:"reason"`
	Status    string     `gorm:"size:24;not null" json:"status"`
	HandledAt *time.Time `json:"handled_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (OutdatedSignal) TableName() string { return "outdated_signals" }

type ContentReport struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ReporterUserID uint       `gorm:"not null;index" json:"reporter_user_id"`
	TargetType     string     `gorm:"size:32;not null" json:"target_type"`
	TargetID       uint       `gorm:"not null" json:"target_id"`
	ReasonCode     string     `gorm:"size:64;not null" json:"reason_code"`
	Description    string     `gorm:"type:text" json:"description"`
	Status         string     `gorm:"size:24;not null" json:"status"`
	Priority       int        `json:"priority"`
	Result         string     `gorm:"type:text" json:"result"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (ContentReport) TableName() string { return "content_reports" }

type AppealCase struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	CaseType          string     `gorm:"size:32;not null" json:"case_type"`
	RecordID          *uint      `json:"record_id,omitempty"`
	ComplainantType   string     `gorm:"size:32;not null" json:"complainant_type"`
	ComplainantUserID *uint      `json:"complainant_user_id,omitempty"`
	ClaimText         string     `gorm:"type:text;not null" json:"claim_text"`
	Status            string     `gorm:"size:24;not null" json:"status"`
	Decision          string     `gorm:"type:text" json:"decision"`
	ResponseDueAt     *time.Time `json:"response_due_at,omitempty"`
	ClosedAt          *time.Time `json:"closed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (AppealCase) TableName() string { return "appeal_cases" }

type RouteFavorite struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	RouteID   uint      `gorm:"primaryKey" json:"route_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (RouteFavorite) TableName() string { return "route_favorites" }

type RouteJourney struct {
	ID               uint         `gorm:"primaryKey" json:"id"`
	UserID           uint         `gorm:"not null;index" json:"user_id"`
	RouteID          uint         `gorm:"not null;index" json:"route_id"`
	Status           string       `gorm:"size:24;not null" json:"status"`
	CompletedStopIDs JSONDocument `gorm:"type:jsonb" json:"completed_stop_ids"`
	StartedAt        time.Time    `json:"started_at"`
	CompletedAt      *time.Time   `json:"completed_at,omitempty"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

func (RouteJourney) TableName() string { return "route_journeys" }

type NutritionRecord struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	UserID            uint           `gorm:"not null;index" json:"-"`
	MealPeriod        string         `gorm:"size:24" json:"meal_period"`
	EatenAt           time.Time      `json:"eaten_at"`
	SourceType        string         `gorm:"size:24;not null" json:"source_type"`
	Description       string         `gorm:"type:text" json:"description"`
	Foods             JSONDocument   `gorm:"type:jsonb" json:"foods"`
	Calories          *float64       `json:"calories,omitempty"`
	ProteinGrams      *float64       `json:"protein_grams,omitempty"`
	FatGrams          *float64       `json:"fat_grams,omitempty"`
	CarbohydrateGrams *float64       `json:"carbohydrate_grams,omitempty"`
	ImageObjectKey    string         `gorm:"size:512" json:"image_url,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (NutritionRecord) TableName() string { return "nutrition_records" }
