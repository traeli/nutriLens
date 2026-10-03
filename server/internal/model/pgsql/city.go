package pgsql

import "time"

// City 表示运营维护并向小程序用户开放的城市；内部发布字段不会通过公开接口返回。
type City struct {
	Code        string    `gorm:"size:16;primaryKey;comment:城市行政编码" json:"code"`
	Name        string    `gorm:"size:64;not null;comment:城市名称" json:"name"`
	Description string    `gorm:"size:240;not null;default:'';comment:城市简介" json:"desc"`
	Image       string    `gorm:"size:1024;not null;default:'';comment:城市展示图片" json:"image"`
	Enabled     bool      `gorm:"not null;default:true;index;comment:是否启用" json:"-"`
	IsDefault   bool      `gorm:"not null;default:false;comment:是否为默认城市" json:"is_default"`
	SortOrder   int       `gorm:"not null;default:0;index;comment:展示顺序" json:"-"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"-"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"-"`
}

func (City) TableName() string { return "cities" }
