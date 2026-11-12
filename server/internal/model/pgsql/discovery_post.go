package pgsql

import "time"

// DiscoveryPost 是由运营人员在 Directus 中维护的公开发现内容。
// 小程序用户只有阅读、收藏店铺和标记“有用”的权限。
type DiscoveryPost struct {
	ID             uint       `gorm:"primaryKey;comment:发现帖子编号" json:"id"`
	PlaceID        uint       `gorm:"not null;index;comment:关联地点编号" json:"place_id"`
	LegacyReviewID *uint      `gorm:"uniqueIndex;comment:迁移前评论编号" json:"-"`
	Title          string     `gorm:"size:200;not null;comment:帖子标题" json:"title"`
	Content        string     `gorm:"type:text;not null;comment:帖子正文" json:"content"`
	AuthorName     string     `gorm:"size:80;not null;default:'饭友记编辑';comment:运营作者名称" json:"author_name"`
	SourceName     string     `gorm:"size:200;not null;default:'';comment:信息来源名称" json:"source_name,omitempty"`
	SourceURL      string     `gorm:"size:1024;not null;default:'';comment:信息来源链接" json:"source_url,omitempty"`
	Status         string     `gorm:"size:24;not null;default:'draft';index;comment:发布状态" json:"status"`
	LikeCount      int        `gorm:"not null;default:0;comment:有用数量" json:"like_count"`
	PublishedAt    *time.Time `gorm:"index;comment:发布时间" json:"published_at,omitempty"`
	CreatedAt      time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

func (DiscoveryPost) TableName() string { return "discovery_posts" }

type DiscoveryPostImage struct {
	ID        uint      `gorm:"primaryKey;comment:帖子图片编号" json:"id"`
	PostID    uint      `gorm:"not null;uniqueIndex:uidx_discovery_post_image_slot;index;comment:帖子编号" json:"post_id"`
	ImageURL  string    `gorm:"size:1024;not null;comment:COS图片地址" json:"image_url"`
	ObjectKey string    `gorm:"size:512;not null;default:'';comment:COS对象键" json:"-"`
	SortOrder int       `gorm:"not null;uniqueIndex:uidx_discovery_post_image_slot;check:discovery_post_image_slot,sort_order >= 0 AND sort_order < 6;comment:图片顺序零到五" json:"sort_order"`
	Width     int       `gorm:"not null;default:0;comment:图片宽度" json:"width"`
	Height    int       `gorm:"not null;default:0;comment:图片高度" json:"height"`
	AltText   string    `gorm:"size:300;not null;default:'';comment:图片说明" json:"alt_text,omitempty"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}

func (DiscoveryPostImage) TableName() string { return "discovery_post_images" }

type DiscoveryPostLike struct {
	PostID    uint      `gorm:"primaryKey;comment:帖子编号" json:"post_id"`
	UserID    uint      `gorm:"primaryKey;comment:用户编号" json:"user_id"`
	CreatedAt time.Time `gorm:"comment:标记时间" json:"created_at"`
}

func (DiscoveryPostLike) TableName() string { return "discovery_post_likes" }
