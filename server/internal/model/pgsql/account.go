package pgsql

import (
	"time"

	"gorm.io/gorm"
)

const UserTableName = "nutrilens_users"

type User struct {
	ID            uint           `gorm:"primaryKey;comment:用户编号" json:"id"`
	OpenID        string         `gorm:"column:open_id;size:128;unique;not null;comment:微信用户唯一标识" json:"-"`
	Nickname      string         `gorm:"size:64;comment:用户昵称" json:"nickname"`
	AvatarURL     string         `gorm:"size:512;comment:头像地址" json:"avatar_url"`
	AccountStatus string         `gorm:"size:24;not null;default:'active';index;comment:账号状态" json:"account_status"`
	CreatedAt     time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index;comment:软删除时间" json:"-"`
}

func (User) TableName() string { return UserTableName }

type PrivacyAgreement struct {
	ID            uint      `gorm:"primaryKey;comment:协议确认记录编号" json:"id"`
	UserID        uint      `gorm:"column:user_id;not null;index;comment:用户编号" json:"user_id"`
	AgreementType string    `gorm:"size:64;not null;comment:协议类型" json:"agreement_type"`
	Version       string    `gorm:"size:32;not null;comment:协议版本" json:"version"`
	IPHash        string    `gorm:"size:128;comment:客户端地址摘要" json:"-"`
	ClientVersion string    `gorm:"size:32;comment:客户端版本" json:"client_version"`
	AgreedAt      time.Time `gorm:"comment:同意时间" json:"agreed_at"`
}

func (PrivacyAgreement) TableName() string { return "privacy_agreements" }
