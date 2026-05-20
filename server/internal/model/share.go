package model

import "time"

// ShareRecord tracks user sharing events and invitation conversions.
type ShareRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	ShareType     string    `gorm:"size:20;not null" json:"share_type"` // poster, card, link
	InvitedUserID uint      `gorm:"index" json:"invited_user_id"`       // 0 if not yet converted
	CreatedAt     time.Time `json:"created_at"`
}

// InviteRelation records who invited whom.
type InviteRelation struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	InviterUserID uint      `gorm:"index;not null" json:"inviter_user_id"`
	InviteeUserID uint      `gorm:"index;not null" json:"invitee_user_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// Share stores the poster image configuration.
type Share struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	ImageURL string `gorm:"type:text;not null" json:"image_url"`
	Type     string `gorm:"size:20;default:poster" json:"type"`
}
