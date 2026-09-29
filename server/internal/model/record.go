package model

import (
	"time"

	"gorm.io/gorm"
)

type VisitRecord struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	UserID           uint           `gorm:"not null;index" json:"-"`
	PlaceID          uint           `gorm:"not null;index" json:"place_id"`
	Visibility       string         `gorm:"size:16;not null;default:'private'" json:"visibility"`
	PublishStatus    string         `gorm:"size:32;not null;default:'draft';index" json:"publish_status"`
	RiskLevel        string         `gorm:"size:16;not null;default:'low'" json:"risk_level"`
	CurrentVersionID *uint          `gorm:"index" json:"current_version_id,omitempty"`
	SubmittedAt      *time.Time     `json:"submitted_at,omitempty"`
	PublishedAt      *time.Time     `gorm:"index" json:"published_at,omitempty"`
	LastConfirmedAt  *time.Time     `json:"last_confirmed_at,omitempty"`
	HelpfulCount     int            `gorm:"not null;default:0" json:"helpful_count"`
	OutdatedCount    int            `gorm:"not null;default:0" json:"outdated_count"`
	ReportCount      int            `gorm:"not null;default:0" json:"report_count"`
	CreateRequestKey *string        `gorm:"size:128" json:"-"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (VisitRecord) TableName() string { return "visit_records" }

type VisitRecordVersion struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	RecordID      uint         `gorm:"not null;index" json:"record_id"`
	VersionNo     int          `gorm:"not null" json:"version_no"`
	EditorUserID  uint         `gorm:"not null;index" json:"-"`
	VisitDate     time.Time    `gorm:"type:date;not null" json:"visit_date"`
	ConsumerType  string       `gorm:"size:32" json:"consumer_type"`
	Conclusion    string       `gorm:"size:32;not null" json:"conclusion"`
	PriceMin      *float64     `gorm:"type:numeric(10,2)" json:"price_min,omitempty"`
	PriceMax      *float64     `gorm:"type:numeric(10,2)" json:"price_max,omitempty"`
	AverageCost   *float64     `gorm:"type:numeric(10,2)" json:"average_cost,omitempty"`
	WaitMinutes   *int         `json:"wait_minutes,omitempty"`
	MealPeriod    string       `gorm:"size:24" json:"meal_period"`
	Dishes        JSONDocument `gorm:"type:jsonb" json:"dishes"`
	Content       string       `gorm:"type:text;not null" json:"content"`
	Visibility    string       `gorm:"size:16;not null" json:"visibility"`
	ChangeSummary string       `gorm:"type:text" json:"change_summary"`
	CreatedAt     time.Time    `json:"created_at"`
}

func (VisitRecordVersion) TableName() string { return "visit_record_versions" }

type VisitRecordTagLink struct {
	RecordVersionID uint `gorm:"primaryKey" json:"record_version_id"`
	TagID           uint `gorm:"primaryKey" json:"tag_id"`
}

func (VisitRecordTagLink) TableName() string { return "visit_record_tag_links" }

type RecordMedia struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	RecordID           uint      `gorm:"not null;index" json:"record_id"`
	RecordVersionID    uint      `gorm:"not null;index" json:"record_version_id"`
	ObjectKey          string    `gorm:"size:512;not null" json:"-"`
	PublicURL          string    `gorm:"size:1024" json:"public_url"`
	MediaType          string    `gorm:"size:24;not null" json:"media_type"`
	Width              int       `json:"width"`
	Height             int       `json:"height"`
	SortOrder          int       `gorm:"not null;default:0" json:"sort_order"`
	SafetyStatus       string    `gorm:"size:24;not null;default:'pending'" json:"safety_status"`
	DesensitizeStatus  string    `gorm:"size:24;not null;default:'pending'" json:"desensitize_status"`
	ProcessedObjectKey string    `gorm:"size:512" json:"-"`
	FaceDetected       bool      `gorm:"not null;default:false" json:"face_detected"`
	CreatedAt          time.Time `json:"created_at"`
}

func (RecordMedia) TableName() string { return "record_media" }

type RecordEvidence struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	RecordID          uint       `gorm:"not null;index" json:"record_id"`
	UserID            uint       `gorm:"not null;index" json:"user_id"`
	EvidenceType      string     `gorm:"size:32;not null" json:"evidence_type"`
	OriginalObjectKey string     `gorm:"size:512;not null" json:"-"`
	MaskedObjectKey   string     `gorm:"size:512" json:"-"`
	VerifyStatus      string     `gorm:"size:24;not null;default:'pending'" json:"verify_status"`
	VerifiedBy        *uint      `json:"verified_by,omitempty"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
	RetentionUntil    *time.Time `json:"retention_until,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (RecordEvidence) TableName() string { return "record_evidences" }
