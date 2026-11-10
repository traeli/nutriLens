package pgsql

import "time"

type Place struct {
	ID                      uint         `gorm:"primaryKey;comment:地点编号" json:"id"`
	Name                    string       `gorm:"size:160;not null;index;comment:地点名称" json:"name"`
	Category                string       `gorm:"size:32;not null;comment:地点分类" json:"category"`
	ThemeCode               string       `gorm:"size:32;not null;default:'general';index;comment:地点主餐饮主题" json:"theme_code"`
	CityCode                string       `gorm:"size:16;not null;index;comment:城市行政编码" json:"city_code"`
	District                string       `gorm:"size:64;comment:行政区" json:"district"`
	BusinessArea            string       `gorm:"size:64;comment:商圈" json:"business_area"`
	Address                 string       `gorm:"size:300;not null;comment:详细地址" json:"address"`
	Description             string       `gorm:"type:text;not null;default:'';comment:地点客观简介" json:"description"`
	Longitude               float64      `gorm:"type:numeric(10,7);not null;comment:经度" json:"longitude"`
	Latitude                float64      `gorm:"type:numeric(10,7);not null;comment:纬度" json:"latitude"`
	POIProvider             *string      `gorm:"column:poi_provider;size:32;comment:地点数据提供方" json:"poi_provider,omitempty"`
	POIID                   *string      `gorm:"column:poi_id;size:128;comment:提供方地点编号" json:"poi_id,omitempty"`
	Status                  string       `gorm:"size:24;not null;default:'pending';index;comment:地点状态" json:"status"`
	AverageCost             *float64     `gorm:"type:numeric(10,2);comment:运营采集的常见人均消费" json:"average_cost,omitempty"`
	TypicalWaitMinutes      *int         `gorm:"comment:运营采集的典型排队分钟数" json:"typical_wait_minutes,omitempty"`
	MealPeriods             JSONDocument `gorm:"type:jsonb;comment:适合用餐时段" json:"meal_periods,omitempty"`
	RecommendationSource    string       `gorm:"size:32;not null;default:'';comment:推荐依据来源" json:"recommendation_source,omitempty"`
	RecommendationSourceURL string       `gorm:"size:1024;not null;default:'';comment:推荐依据来源链接" json:"recommendation_source_url,omitempty"`
	RecommendationNote      string       `gorm:"type:text;not null;default:'';comment:推荐依据摘要" json:"recommendation_note,omitempty"`
	SignalConfidence        string       `gorm:"size:16;not null;default:'';comment:推荐依据置信度" json:"signal_confidence,omitempty"`
	MergedIntoID            *uint        `gorm:"index;comment:合并后的目标地点编号" json:"merged_into_id,omitempty"`
	CreatedBy               *uint        `gorm:"index;comment:提交用户编号" json:"created_by,omitempty"`
	CreatedAt               time.Time    `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt               time.Time    `gorm:"comment:更新时间" json:"updated_at"`
}

func (Place) TableName() string { return "places" }

type Tag struct {
	ID        uint   `gorm:"primaryKey;comment:标签编号" json:"id"`
	Code      string `gorm:"size:64;not null;unique;comment:标签编码" json:"code"`
	Name      string `gorm:"size:64;not null;comment:标签名称" json:"name"`
	GroupName string `gorm:"size:64;not null;index;comment:标签分组" json:"group_name"`
	Enabled   bool   `gorm:"not null;default:true;index;comment:是否启用" json:"enabled"`
	SortOrder int    `gorm:"not null;default:0;comment:展示顺序" json:"sort_order"`
}

func (Tag) TableName() string { return "tags" }
