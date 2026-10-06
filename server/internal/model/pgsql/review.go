package pgsql

import (
	"time"

	"gorm.io/gorm"
)

// RestaurantReview 是从一次私人足迹创建的公开评论。兼容迁移期间评论 ID 与来源足迹 ID 保持一致，
// 以便在拆分业务边界时继续兼容已有公开链接和互动数据。
type RestaurantReview struct {
	ID               uint                  `gorm:"primaryKey;autoIncrement:false;comment:评论编号" json:"id"`
	UserID           uint                  `gorm:"not null;index;comment:用户编号" json:"-"`
	PlaceID          uint                  `gorm:"not null;index;comment:地点编号" json:"place_id"`
	VisitRecordID    uint                  `gorm:"not null;uniqueIndex;comment:来源足迹编号" json:"visit_record_id"`
	CurrentVersionID *uint                 `gorm:"index;comment:当前评论版本编号" json:"current_version_id,omitempty"`
	PublishStatus    string                `gorm:"column:status;size:24;not null;index;comment:评论发布状态" json:"publish_status"`
	RiskLevel        string                `gorm:"size:16;not null;default:'low';comment:内容风险等级" json:"risk_level"`
	SubmittedAt      time.Time             `gorm:"comment:提交审核时间" json:"submitted_at"`
	PublishedAt      *time.Time            `gorm:"index;comment:公开发布时间" json:"published_at,omitempty"`
	HelpfulCount     int                   `gorm:"not null;default:0;comment:有帮助数量" json:"helpful_count"`
	OutdatedCount    int                   `gorm:"not null;default:0;comment:过时反馈数量" json:"outdated_count"`
	ReportCount      int                   `gorm:"not null;default:0;comment:举报数量" json:"report_count"`
	Media            JSONList[RecordMedia] `gorm:"type:jsonb;not null;default:'[]';comment:评论媒体快照" json:"-"`
	CreatedAt        time.Time             `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time             `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt        gorm.DeletedAt        `gorm:"index;comment:软删除时间" json:"-"`
}

func (RestaurantReview) TableName() string { return "restaurant_reviews" }

type RestaurantReviewVersion struct {
	ID            uint         `gorm:"primaryKey;comment:评论版本编号" json:"id"`
	ReviewID      uint         `gorm:"not null;index;comment:评论编号" json:"review_id"`
	VersionNo     int          `gorm:"not null;comment:版本序号" json:"version_no"`
	EditorUserID  uint         `gorm:"not null;index;comment:编辑用户编号" json:"-"`
	VisitDate     time.Time    `gorm:"type:date;not null;comment:到店日期" json:"visit_date"`
	Conclusion    string       `gorm:"size:32;not null;comment:总体感受" json:"conclusion"`
	AverageCost   *float64     `gorm:"type:numeric(10,2);comment:人均消费金额" json:"average_cost,omitempty"`
	WaitMinutes   *int         `gorm:"comment:等待分钟数" json:"wait_minutes,omitempty"`
	MealPeriod    string       `gorm:"size:24;comment:用餐时段" json:"meal_period"`
	Dishes        JSONDocument `gorm:"type:jsonb;comment:菜品列表" json:"dishes"`
	Content       string       `gorm:"type:text;not null;comment:评论正文" json:"content"`
	ChangeSummary string       `gorm:"type:text;comment:修改摘要" json:"change_summary"`
	CreatedAt     time.Time    `gorm:"comment:创建时间" json:"created_at"`
}

func (RestaurantReviewVersion) TableName() string { return "restaurant_review_versions" }

type RestaurantReviewTagLink struct {
	ReviewVersionID uint `gorm:"primaryKey;comment:评论版本编号" json:"review_version_id"`
	TagID           uint `gorm:"primaryKey;comment:标签编号" json:"tag_id"`
}

func (RestaurantReviewTagLink) TableName() string { return "restaurant_review_tag_links" }
