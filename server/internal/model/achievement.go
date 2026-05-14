package model

import "time"

// Achievement defines an achievement that users can unlock.
type Achievement struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Code           string    `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Name           string    `gorm:"size:100;not null" json:"name"`
	Icon           string    `gorm:"size:50" json:"icon"`
	Description    string    `gorm:"type:text" json:"description"`
	Rarity         string    `gorm:"size:20;not null" json:"rarity"` // common, rare, epic, legendary
	ConditionType  string    `gorm:"size:50;not null" json:"condition_type"`
	ConditionValue int       `json:"condition_value"`
	RewardType     string    `gorm:"size:32" json:"reward_type"` // quota, poster_template, vip
	RewardValue    int       `json:"reward_value"`               // quota amount or template id
	Enabled        bool      `gorm:"default:true" json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
}

// UserAchievement records which achievements a user has unlocked.
type UserAchievement struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"uniqueIndex:idx_user_achievement;not null" json:"user_id"`
	AchievementID uint      `gorm:"uniqueIndex:idx_user_achievement;not null" json:"achievement_id"`
	UnlockedAt    time.Time `json:"unlocked_at"`
}
