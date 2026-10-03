package pgsql

import (
	"time"

	"gorm.io/gorm"
)

type VisitRecord struct {
	ID               uint           `gorm:"primaryKey;comment:足迹编号" json:"id"`
	UserID           uint           `gorm:"not null;index;comment:用户编号" json:"-"`
	PlaceID          uint           `gorm:"not null;index;comment:地点编号" json:"place_id"`
	Visibility       string         `gorm:"size:16;not null;default:'private';comment:可见范围" json:"visibility"`
	PublishStatus    string         `gorm:"size:32;not null;default:'draft';index;comment:兼容用发布状态" json:"publish_status"`
	RiskLevel        string         `gorm:"size:16;not null;default:'low';comment:兼容用风险等级" json:"risk_level"`
	CurrentVersionID *uint          `gorm:"index;comment:当前版本编号" json:"current_version_id,omitempty"`
	SubmittedAt      *time.Time     `gorm:"comment:兼容用提交时间" json:"submitted_at,omitempty"`
	PublishedAt      *time.Time     `gorm:"index;comment:兼容用发布时间" json:"published_at,omitempty"`
	LastConfirmedAt  *time.Time     `gorm:"comment:最后确认有效时间" json:"last_confirmed_at,omitempty"`
	HelpfulCount     int            `gorm:"not null;default:0;comment:兼容用有帮助数量" json:"helpful_count"`
	OutdatedCount    int            `gorm:"not null;default:0;comment:兼容用过时反馈数量" json:"outdated_count"`
	ReportCount      int            `gorm:"not null;default:0;comment:兼容用举报数量" json:"report_count"`
	CreateRequestKey *string        `gorm:"size:128;comment:创建请求幂等键" json:"-"`
	CreatedAt        time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index;comment:软删除时间" json:"-"`
}

func (VisitRecord) TableName() string { return "visit_records" }

type VisitRecordVersion struct {
	ID            uint         `gorm:"primaryKey;comment:足迹版本编号" json:"id"`
	RecordID      uint         `gorm:"not null;index;comment:足迹编号" json:"record_id"`
	VersionNo     int          `gorm:"not null;comment:版本序号" json:"version_no"`
	EditorUserID  uint         `gorm:"not null;index;comment:编辑用户编号" json:"-"`
	VisitDate     time.Time    `gorm:"type:date;not null;comment:到店日期" json:"visit_date"`
	ConsumerType  string       `gorm:"size:32;comment:消费类型" json:"consumer_type"`
	Conclusion    string       `gorm:"size:32;not null;comment:总体感受" json:"conclusion"`
	PriceMin      *float64     `gorm:"type:numeric(10,2);comment:最低消费金额" json:"price_min,omitempty"`
	PriceMax      *float64     `gorm:"type:numeric(10,2);comment:最高消费金额" json:"price_max,omitempty"`
	AverageCost   *float64     `gorm:"type:numeric(10,2);comment:人均消费金额" json:"average_cost,omitempty"`
	WaitMinutes   *int         `gorm:"comment:等待分钟数" json:"wait_minutes,omitempty"`
	MealPeriod    string       `gorm:"size:24;comment:用餐时段" json:"meal_period"`
	Dishes        JSONDocument `gorm:"type:jsonb;comment:菜品列表" json:"dishes"`
	Content       string       `gorm:"type:text;not null;comment:体验正文" json:"content"`
	Visibility    string       `gorm:"size:16;not null;comment:可见范围" json:"visibility"`
	ChangeSummary string       `gorm:"type:text;comment:修改摘要" json:"change_summary"`
	CreatedAt     time.Time    `gorm:"comment:创建时间" json:"created_at"`
}

func (VisitRecordVersion) TableName() string { return "visit_record_versions" }

type VisitRecordTagLink struct {
	RecordVersionID uint `gorm:"primaryKey;comment:足迹版本编号" json:"record_version_id"`
	TagID           uint `gorm:"primaryKey;comment:标签编号" json:"tag_id"`
}

func (VisitRecordTagLink) TableName() string { return "visit_record_tag_links" }

type RecordMedia struct {
	ID                 uint      `gorm:"primaryKey;comment:媒体编号" json:"id"`
	RecordID           uint      `gorm:"not null;index;comment:足迹编号" json:"record_id"`
	RecordVersionID    uint      `gorm:"not null;index;comment:足迹版本编号" json:"record_version_id"`
	ObjectKey          string    `gorm:"size:512;not null;comment:原始对象存储键" json:"-"`
	PublicURL          string    `gorm:"size:1024;comment:公开访问地址" json:"public_url"`
	MediaType          string    `gorm:"size:24;not null;comment:媒体类型" json:"media_type"`
	Width              int       `gorm:"comment:图片宽度" json:"width"`
	Height             int       `gorm:"comment:图片高度" json:"height"`
	SortOrder          int       `gorm:"not null;default:0;comment:展示顺序" json:"sort_order"`
	SafetyStatus       string    `gorm:"size:24;not null;default:'pending';comment:内容安全状态" json:"safety_status"`
	DesensitizeStatus  string    `gorm:"size:24;not null;default:'pending';comment:脱敏状态" json:"desensitize_status"`
	ProcessedObjectKey string    `gorm:"size:512;comment:处理后对象存储键" json:"-"`
	FaceDetected       bool      `gorm:"not null;default:false;comment:是否检测到人脸" json:"face_detected"`
	CreatedAt          time.Time `gorm:"comment:创建时间" json:"created_at"`
}

func (RecordMedia) TableName() string { return "record_media" }

type RecordEvidence struct {
	ID                uint       `gorm:"primaryKey;comment:消费凭证编号" json:"id"`
	RecordID          uint       `gorm:"not null;index;comment:足迹编号" json:"record_id"`
	UserID            uint       `gorm:"not null;index;comment:用户编号" json:"user_id"`
	EvidenceType      string     `gorm:"size:32;not null;comment:凭证类型" json:"evidence_type"`
	OriginalObjectKey string     `gorm:"size:512;not null;comment:原始对象存储键" json:"-"`
	MaskedObjectKey   string     `gorm:"size:512;comment:脱敏对象存储键" json:"-"`
	VerifyStatus      string     `gorm:"size:24;not null;default:'pending';comment:核验状态" json:"verify_status"`
	VerifiedBy        *uint      `gorm:"comment:核验人员编号" json:"verified_by,omitempty"`
	VerifiedAt        *time.Time `gorm:"comment:核验时间" json:"verified_at,omitempty"`
	RetentionUntil    *time.Time `gorm:"comment:保留截止时间" json:"retention_until,omitempty"`
	CreatedAt         time.Time  `gorm:"comment:创建时间" json:"created_at"`
}

func (RecordEvidence) TableName() string { return "record_evidences" }
