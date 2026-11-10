package pgsql

import "time"

// PlaceTheme 由 Directus 维护餐厅主题与公开 COS 图标之间的稳定映射。
type PlaceTheme struct {
	Code      string    `gorm:"size:32;primaryKey;comment:主题编码" json:"code"`
	Name      string    `gorm:"size:64;not null;comment:主题中文名称" json:"name"`
	IconURL   string    `gorm:"size:1024;not null;comment:COS 图标公开地址" json:"icon_url"`
	Enabled   bool      `gorm:"not null;default:true;index;comment:是否启用" json:"enabled"`
	SortOrder int       `gorm:"not null;default:0;comment:展示顺序" json:"sort_order"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (PlaceTheme) TableName() string { return "place_themes" }
