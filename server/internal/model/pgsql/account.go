package pgsql

import (
	"time"

	"gorm.io/gorm"
)

const UserTableName = "nutrilens_users"

type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	OpenID        string         `gorm:"column:open_id;size:128;unique;not null" json:"-"`
	Email         *string        `gorm:"column:email;size:254;uniqueIndex:idx_nutrilens_users_email" json:"-"`
	Nickname      string         `gorm:"size:64" json:"nickname"`
	AvatarURL     string         `gorm:"size:512" json:"avatar_url"`
	AccountStatus string         `gorm:"size:24;not null;default:'active';index" json:"account_status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return UserTableName }

type PrivacyAgreement struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	AgreementType string    `gorm:"size:64;not null" json:"agreement_type"`
	Version       string    `gorm:"size:32;not null" json:"version"`
	IPHash        string    `gorm:"size:128" json:"-"`
	ClientVersion string    `gorm:"size:32" json:"client_version"`
	AgreedAt      time.Time `json:"agreed_at"`
}

func (PrivacyAgreement) TableName() string { return "privacy_agreements" }
