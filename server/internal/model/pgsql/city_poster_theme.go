package pgsql

import "time"

// CityPosterTheme 配置非地理类饮食记忆海报，不得包含道路、边界、坐标或导航几何数据。
type CityPosterTheme struct {
	CityCode        string    `gorm:"size:16;primaryKey;comment:城市行政编码" json:"city_code"`
	Version         int       `gorm:"not null;default:1;comment:主题版本" json:"version"`
	Title           string    `gorm:"size:80;not null;comment:海报标题" json:"title"`
	Subtitle        string    `gorm:"size:180;not null;default:'';comment:海报副标题" json:"subtitle"`
	BackgroundImage string    `gorm:"size:1024;not null;default:'';comment:背景图片地址" json:"background_image"`
	AccentColor     string    `gorm:"size:16;not null;default:'#C7FF35';comment:强调色" json:"accent_color"`
	SecondaryColor  string    `gorm:"size:16;not null;default:'#F2E7C9';comment:辅助色" json:"secondary_color"`
	Motifs          []byte    `gorm:"type:jsonb;not null;default:'[]';comment:装饰性主题词" json:"-"`
	CreatedAt       time.Time `gorm:"comment:创建时间" json:"-"`
	UpdatedAt       time.Time `gorm:"comment:更新时间" json:"-"`
}

func (CityPosterTheme) TableName() string { return "city_poster_themes" }
