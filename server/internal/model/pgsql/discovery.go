package pgsql

import "time"

// DiscoveryScenario 是首页可选的路线生成场景。空 CityCode 表示所有城市共用，
// 同 code 的城市级配置会覆盖全局配置。
type DiscoveryScenario struct {
	ID             uint      `gorm:"primaryKey;comment:场景编号" json:"-"`
	CityCode       string    `gorm:"size:16;not null;default:'';uniqueIndex:idx_discovery_scenario_scope;comment:城市行政编码，空值表示全局" json:"-"`
	Code           string    `gorm:"size:48;not null;uniqueIndex:idx_discovery_scenario_scope;comment:稳定场景编码" json:"code"`
	Label          string    `gorm:"size:32;not null;comment:场景展示名称" json:"label"`
	ConditionType  string    `gorm:"size:32;not null;default:'scene';comment:查询条件类型" json:"condition_type"`
	ConditionValue string    `gorm:"size:64;not null;comment:查询条件值" json:"condition_value"`
	Enabled        bool      `gorm:"not null;default:true;index;comment:是否启用" json:"-"`
	SortOrder      int       `gorm:"not null;default:0;index;comment:展示顺序" json:"sort_order"`
	CreatedAt      time.Time `gorm:"comment:创建时间" json:"-"`
	UpdatedAt      time.Time `gorm:"comment:更新时间" json:"-"`
}

func (DiscoveryScenario) TableName() string { return "discovery_scenarios" }
