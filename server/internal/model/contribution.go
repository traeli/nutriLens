package model

import "time"

type UserTrustProfile struct {
	UserID              uint         `gorm:"primaryKey" json:"user_id"`
	TrustLevel          string       `gorm:"size:24;not null" json:"trust_level"`
	AccountAgeScore     int          `gorm:"not null;default:0" json:"account_age_score"`
	VerifiedRecordCount int          `gorm:"not null;default:0" json:"verified_record_count"`
	ViolationCount      int          `gorm:"not null;default:0" json:"violation_count"`
	DailyPublishLimit   int          `gorm:"not null;default:1" json:"daily_publish_limit"`
	RiskFlags           JSONDocument `gorm:"type:jsonb" json:"risk_flags"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

func (UserTrustProfile) TableName() string { return "user_trust_profiles" }

type ContributionAccount struct {
	UserID      uint      `gorm:"primaryKey" json:"user_id"`
	TotalPoints int       `gorm:"not null;default:0" json:"total_points"`
	MonthPoints int       `gorm:"not null;default:0" json:"month_points"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ContributionAccount) TableName() string { return "contribution_accounts" }

type Badge struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	Code           string       `gorm:"size:64;not null;uniqueIndex" json:"code"`
	Name           string       `gorm:"size:100;not null" json:"name"`
	Description    string       `gorm:"type:text" json:"description"`
	IconURL        string       `gorm:"size:1024" json:"icon_url"`
	ConditionType  string       `gorm:"size:64;not null" json:"condition_type"`
	ConditionValue JSONDocument `gorm:"type:jsonb" json:"condition_value"`
	Enabled        bool         `gorm:"not null;default:true" json:"enabled"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

func (Badge) TableName() string { return "badges" }

type UserBadge struct {
	UserID     uint      `gorm:"primaryKey" json:"user_id"`
	BadgeID    uint      `gorm:"primaryKey" json:"badge_id"`
	UnlockedAt time.Time `json:"unlocked_at"`
}

func (UserBadge) TableName() string { return "user_badges" }
