package pgsql

import "time"

// CityRoute 是运营维护并展示在小程序首页的餐饮路线，同时为路线轮播卡片和精选路线卡片提供数据；
// 内部发布字段不会通过公开接口返回。
type CityRoute struct {
	ID         uint         `gorm:"primaryKey;comment:路线编号" json:"-"`
	CityCode   string       `gorm:"size:16;not null;index;comment:城市行政编码" json:"-"`
	Title      string       `gorm:"size:120;not null;comment:路线标题" json:"-"`
	Tag        string       `gorm:"size:32;not null;default:'';comment:路线标签" json:"-"`
	Meta       string       `gorm:"size:120;not null;default:'';comment:路线摘要" json:"-"`
	Keyword    string       `gorm:"size:64;not null;default:'';comment:搜索关键词" json:"-"`
	Image      string       `gorm:"size:1024;not null;default:'';comment:路线图片" json:"-"`
	Footnote   string       `gorm:"size:240;not null;default:'';comment:路线补充说明" json:"-"`
	Stops      JSONDocument `gorm:"type:jsonb;comment:兼容用站点名称列表" json:"-"`
	IsFeatured bool         `gorm:"not null;default:false;comment:是否精选" json:"-"`
	Enabled    bool         `gorm:"not null;default:true;index;comment:是否启用" json:"-"`
	SortOrder  int          `gorm:"not null;default:0;index;comment:展示顺序" json:"-"`
	CreatedAt  time.Time    `gorm:"comment:创建时间" json:"-"`
	UpdatedAt  time.Time    `gorm:"comment:更新时间" json:"-"`
}

func (CityRoute) TableName() string { return "city_routes" }

// CityRouteStop 将攻略路线保存为有序的标准地点列表。经纬度和地址统一保存在 Place 中，
// 修正地点信息后，所有引用该地点的路线都会同步生效。
type CityRouteStop struct {
	ID        uint      `gorm:"primaryKey;comment:路线站点编号" json:"-"`
	RouteID   uint      `gorm:"not null;index;comment:路线编号" json:"-"`
	PlaceID   uint      `gorm:"not null;index;comment:地点编号" json:"place_id"`
	SortOrder int       `gorm:"not null;default:0;comment:站点顺序" json:"index"`
	Note      string    `gorm:"size:240;not null;default:'';comment:站点说明" json:"note"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"-"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"-"`
}

func (CityRouteStop) TableName() string { return "city_route_stops" }
