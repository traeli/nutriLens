package model

import "time"

// CityRoute is an operator-managed dining route shown on the mini program home
// page. Rows feed both the route swiper cards and the featured hand-drawn route
// card. Internal publishing fields are not exposed by the public API.
type CityRoute struct {
	ID         uint         `gorm:"primaryKey" json:"-"`
	CityCode   string       `gorm:"size:16;not null;index" json:"-"`
	Title      string       `gorm:"size:120;not null" json:"-"`
	Tag        string       `gorm:"size:32;not null;default:''" json:"-"`
	Meta       string       `gorm:"size:120;not null;default:''" json:"-"`
	Keyword    string       `gorm:"size:64;not null;default:''" json:"-"`
	Image      string       `gorm:"size:1024;not null;default:''" json:"-"`
	Footnote   string       `gorm:"size:240;not null;default:''" json:"-"`
	Stops      JSONDocument `gorm:"type:jsonb" json:"-"`
	IsFeatured bool         `gorm:"not null;default:false" json:"-"`
	Enabled    bool         `gorm:"not null;default:true;index" json:"-"`
	SortOrder  int          `gorm:"not null;default:0;index" json:"-"`
	CreatedAt  time.Time    `json:"-"`
	UpdatedAt  time.Time    `json:"-"`
}

func (CityRoute) TableName() string { return "city_routes" }

// CityRouteStop keeps an攻略 route as an ordered list of canonical places.
// Coordinates and addresses stay on Place so corrections are reflected in
// every route that references the place.
type CityRouteStop struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	RouteID   uint      `gorm:"not null;index" json:"-"`
	PlaceID   uint      `gorm:"not null;index" json:"place_id"`
	SortOrder int       `gorm:"not null;default:0" json:"index"`
	Note      string    `gorm:"size:240;not null;default:''" json:"note"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (CityRouteStop) TableName() string { return "city_route_stops" }
