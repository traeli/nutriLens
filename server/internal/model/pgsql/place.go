package pgsql

import "time"

type Place struct {
	ID           uint      `gorm:"primaryKey;comment:地点编号" json:"id"`
	Name         string    `gorm:"size:160;not null;comment:地点名称" json:"name"`
	Category     string    `gorm:"size:32;not null;comment:地点分类" json:"category"`
	CityCode     string    `gorm:"size:16;not null;index;comment:城市行政编码" json:"city_code"`
	District     string    `gorm:"size:64;comment:行政区" json:"district"`
	BusinessArea string    `gorm:"size:64;comment:商圈" json:"business_area"`
	Address      string    `gorm:"size:300;not null;comment:详细地址" json:"address"`
	Description  string    `gorm:"type:text;not null;default:'';comment:地点客观简介" json:"description"`
	Longitude    float64   `gorm:"type:numeric(10,7);not null;comment:经度" json:"longitude"`
	Latitude     float64   `gorm:"type:numeric(10,7);not null;comment:纬度" json:"latitude"`
	POIProvider  *string   `gorm:"column:poi_provider;size:32;comment:地点数据提供方" json:"poi_provider,omitempty"`
	POIID        *string   `gorm:"column:poi_id;size:128;comment:提供方地点编号" json:"poi_id,omitempty"`
	Status       string    `gorm:"size:24;not null;default:'pending';index;comment:地点状态" json:"status"`
	MergedIntoID *uint     `gorm:"index;comment:合并后的目标地点编号" json:"merged_into_id,omitempty"`
	CreatedBy    *uint     `gorm:"index;comment:提交用户编号" json:"created_by,omitempty"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"comment:更新时间" json:"updated_at"`
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
