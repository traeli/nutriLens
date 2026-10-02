package pgsql

import "time"

// CityPosterTheme configures a non-geographic dining-memory poster. It must
// not contain road, boundary, coordinate, or navigation geometry.
type CityPosterTheme struct {
	CityCode        string    `gorm:"size:16;primaryKey" json:"city_code"`
	Version         int       `gorm:"not null;default:1" json:"version"`
	Title           string    `gorm:"size:80;not null" json:"title"`
	Subtitle        string    `gorm:"size:180;not null;default:''" json:"subtitle"`
	BackgroundImage string    `gorm:"size:1024;not null;default:''" json:"background_image"`
	AccentColor     string    `gorm:"size:16;not null;default:'#C7FF35'" json:"accent_color"`
	SecondaryColor  string    `gorm:"size:16;not null;default:'#F2E7C9'" json:"secondary_color"`
	Motifs          []byte    `gorm:"type:jsonb;not null;default:'[]'" json:"-"`
	CreatedAt       time.Time `json:"-"`
	UpdatedAt       time.Time `json:"-"`
}

func (CityPosterTheme) TableName() string { return "city_poster_themes" }
