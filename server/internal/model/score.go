package model

import "time"

// UserScore tracks a user's accumulated score and streak.
type UserScore struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	TotalScore    int        `gorm:"default:0" json:"total_score"`
	WeekScore     int        `gorm:"default:0" json:"week_score"`
	StreakDays    int        `gorm:"default:0" json:"streak_days"`
	LastCheckinAt *time.Time `json:"last_checkin_at"`
	BonusQuota    int        `gorm:"default:0" json:"bonus_quota"` // 额外识别次数奖励
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ScoreLog records individual score changes.
type ScoreLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	Action      string    `gorm:"size:50;not null" json:"action"`
	Points      int       `json:"points"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// FriendRelation records bidirectional friend relationships via invitation.
type FriendRelation struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	FriendUserID uint      `gorm:"index;not null" json:"friend_user_id"`
	CreatedAt    time.Time `json:"created_at"`
}
