package pgsql

import "time"

type UserTrustProfile struct {
	UserID              uint         `gorm:"primaryKey;comment:用户编号" json:"user_id"`
	TrustLevel          string       `gorm:"size:24;not null;comment:可信等级" json:"trust_level"`
	AccountAgeScore     int          `gorm:"not null;default:0;comment:账号年龄得分" json:"account_age_score"`
	VerifiedRecordCount int          `gorm:"not null;default:0;comment:已核验记录数量" json:"verified_record_count"`
	ViolationCount      int          `gorm:"not null;default:0;comment:违规次数" json:"violation_count"`
	DailyPublishLimit   int          `gorm:"not null;default:1;comment:每日发布上限" json:"daily_publish_limit"`
	RiskFlags           JSONDocument `gorm:"type:jsonb;comment:风险标记" json:"risk_flags"`
	UpdatedAt           time.Time    `gorm:"comment:更新时间" json:"updated_at"`
}

func (UserTrustProfile) TableName() string { return "user_trust_profiles" }

type ContributionAccount struct {
	UserID      uint      `gorm:"primaryKey;comment:用户编号" json:"user_id"`
	TotalPoints int       `gorm:"not null;default:0;comment:累计贡献积分" json:"total_points"`
	MonthPoints int       `gorm:"not null;default:0;comment:本月贡献积分" json:"month_points"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (ContributionAccount) TableName() string { return "contribution_accounts" }

type Badge struct {
	ID             uint         `gorm:"primaryKey;comment:徽章编号" json:"id"`
	Code           string       `gorm:"size:64;not null;unique;comment:徽章编码" json:"code"`
	Name           string       `gorm:"size:100;not null;comment:徽章名称" json:"name"`
	Description    string       `gorm:"type:text;comment:徽章说明" json:"description"`
	IconURL        string       `gorm:"size:1024;comment:徽章图标地址" json:"icon_url"`
	ConditionType  string       `gorm:"size:64;not null;comment:解锁条件类型" json:"condition_type"`
	ConditionValue JSONDocument `gorm:"type:jsonb;comment:解锁条件参数" json:"condition_value"`
	Enabled        bool         `gorm:"not null;default:true;comment:是否启用" json:"enabled"`
	CreatedAt      time.Time    `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time    `gorm:"comment:更新时间" json:"updated_at"`
}

func (Badge) TableName() string { return "badges" }

type UserBadge struct {
	UserID     uint      `gorm:"primaryKey;comment:用户编号" json:"user_id"`
	BadgeID    uint      `gorm:"primaryKey;comment:徽章编号" json:"badge_id"`
	UnlockedAt time.Time `gorm:"comment:解锁时间" json:"unlocked_at"`
}

func (UserBadge) TableName() string { return "user_badges" }
