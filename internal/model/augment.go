package model

import "time"

type Augment struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	RiotID      string    `gorm:"column:riot_id"`
	NameZh      string    `gorm:"column:name_zh"`
	NameEn      *string   `gorm:"column:name_en"`
	Description *string   `gorm:"column:description"`
	Rarity      *string   `gorm:"column:rarity"`
	IconURL     *string   `gorm:"column:icon_url"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (Augment) TableName() string { return "augments" }

type AugmentPatchProfile struct {
	PatchID     uint64  `gorm:"column:patch_id;primaryKey"`
	AugmentID   uint64  `gorm:"column:augment_id;primaryKey"`
	NameZh      string  `gorm:"column:name_zh"`
	NameEn      *string `gorm:"column:name_en"`
	Description *string `gorm:"column:description"`
	Rarity      *string `gorm:"column:rarity"`
	IconURL     *string `gorm:"column:icon_url"`
	SourceID    uint64  `gorm:"column:source_id"`
}

func (AugmentPatchProfile) TableName() string { return "augment_patch_profiles" }

type AugmentPatchStat struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	PatchID     uint64    `gorm:"column:patch_id"`
	AugmentID   uint64    `gorm:"column:augment_id"`
	SampleCount uint32    `gorm:"column:sample_count"`
	PickRate    *float64  `gorm:"column:pick_rate"`
	WinRate     *float64  `gorm:"column:win_rate"`
	Tier        *string   `gorm:"column:tier"`
	SourceID    *uint64   `gorm:"column:source_id"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (AugmentPatchStat) TableName() string { return "augment_patch_stats" }
