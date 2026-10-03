package pgsql

import (
	"time"
)

type PlaceFavorite struct {
	UserID    uint      `gorm:"primaryKey;comment:用户编号" json:"user_id"`
	PlaceID   uint      `gorm:"primaryKey;comment:地点编号" json:"place_id"`
	CreatedAt time.Time `gorm:"comment:收藏时间" json:"created_at"`
}

func (PlaceFavorite) TableName() string { return "place_favorites" }

type HelpfulVote struct {
	UserID    uint      `gorm:"primaryKey;comment:用户编号" json:"user_id"`
	RecordID  uint      `gorm:"primaryKey;comment:评论编号" json:"record_id"`
	CreatedAt time.Time `gorm:"comment:投票时间" json:"created_at"`
}

func (HelpfulVote) TableName() string { return "helpful_votes" }

type OutdatedSignal struct {
	ID        uint       `gorm:"primaryKey;comment:过时反馈编号" json:"id"`
	UserID    uint       `gorm:"not null;index;comment:用户编号" json:"user_id"`
	RecordID  uint       `gorm:"not null;index;comment:评论编号" json:"record_id"`
	Reason    string     `gorm:"type:text;comment:反馈原因" json:"reason"`
	Status    string     `gorm:"size:24;not null;comment:处理状态" json:"status"`
	HandledAt *time.Time `gorm:"comment:处理时间" json:"handled_at,omitempty"`
	CreatedAt time.Time  `gorm:"comment:创建时间" json:"created_at"`
}

func (OutdatedSignal) TableName() string { return "outdated_signals" }

type ContentReport struct {
	ID             uint       `gorm:"primaryKey;comment:举报编号" json:"id"`
	ReporterUserID uint       `gorm:"not null;index;comment:举报用户编号" json:"reporter_user_id"`
	TargetType     string     `gorm:"size:32;not null;comment:举报对象类型" json:"target_type"`
	TargetID       uint       `gorm:"not null;comment:举报对象编号" json:"target_id"`
	ReasonCode     string     `gorm:"size:64;not null;comment:举报原因编码" json:"reason_code"`
	Description    string     `gorm:"type:text;comment:举报说明" json:"description"`
	Status         string     `gorm:"size:24;not null;comment:处理状态" json:"status"`
	Priority       int        `gorm:"comment:处理优先级" json:"priority"`
	Result         string     `gorm:"type:text;comment:处理结果" json:"result"`
	ResolvedAt     *time.Time `gorm:"comment:解决时间" json:"resolved_at,omitempty"`
	CreatedAt      time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

func (ContentReport) TableName() string { return "content_reports" }

type AppealCase struct {
	ID                uint       `gorm:"primaryKey;comment:申诉案件编号" json:"id"`
	CaseType          string     `gorm:"size:32;not null;comment:案件类型" json:"case_type"`
	RecordID          *uint      `gorm:"comment:关联评论编号" json:"record_id,omitempty"`
	ComplainantType   string     `gorm:"size:32;not null;comment:申诉人类型" json:"complainant_type"`
	ComplainantUserID *uint      `gorm:"comment:申诉用户编号" json:"complainant_user_id,omitempty"`
	ClaimText         string     `gorm:"type:text;not null;comment:申诉内容" json:"claim_text"`
	Status            string     `gorm:"size:24;not null;comment:案件状态" json:"status"`
	Decision          string     `gorm:"type:text;comment:处理决定" json:"decision"`
	ResponseDueAt     *time.Time `gorm:"comment:答复截止时间" json:"response_due_at,omitempty"`
	ClosedAt          *time.Time `gorm:"comment:关闭时间" json:"closed_at,omitempty"`
	CreatedAt         time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

func (AppealCase) TableName() string { return "appeal_cases" }

type RouteFavorite struct {
	UserID    uint      `gorm:"primaryKey;comment:用户编号" json:"user_id"`
	RouteID   uint      `gorm:"primaryKey;comment:路线编号" json:"route_id"`
	CreatedAt time.Time `gorm:"comment:收藏时间" json:"created_at"`
}

func (RouteFavorite) TableName() string { return "route_favorites" }

type RouteJourney struct {
	ID               uint         `gorm:"primaryKey;comment:路线行程编号" json:"id"`
	UserID           uint         `gorm:"not null;index;comment:用户编号" json:"user_id"`
	RouteID          uint         `gorm:"not null;index;comment:路线编号" json:"route_id"`
	Status           string       `gorm:"size:24;not null;comment:行程状态" json:"status"`
	CompletedStopIDs JSONDocument `gorm:"type:jsonb;comment:已完成站点编号列表" json:"completed_stop_ids"`
	StartedAt        time.Time    `gorm:"comment:开始时间" json:"started_at"`
	CompletedAt      *time.Time   `gorm:"comment:完成时间" json:"completed_at,omitempty"`
	UpdatedAt        time.Time    `gorm:"comment:更新时间" json:"updated_at"`
}

func (RouteJourney) TableName() string { return "route_journeys" }
