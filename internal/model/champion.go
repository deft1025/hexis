package model

import "time"

type Champion struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	RiotID    string    `gorm:"column:riot_id"`
	NameZh    string    `gorm:"column:name_zh"`
	NameEn    *string   `gorm:"column:name_en"`
	IconURL   *string   `gorm:"column:icon_url"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Champion) TableName() string { return "champions" }

type ChampionPatchProfile struct {
	PatchID    uint64  `gorm:"column:patch_id;primaryKey"`
	ChampionID uint64  `gorm:"column:champion_id;primaryKey"`
	NameZh     string  `gorm:"column:name_zh"`
	NameEn     *string `gorm:"column:name_en"`
	IconURL    *string `gorm:"column:icon_url"`
	SourceID   uint64  `gorm:"column:source_id"`
}

func (ChampionPatchProfile) TableName() string { return "champion_patch_profiles" }

type ChampionPatchStat struct {
	ID          uint64   `gorm:"column:id;primaryKey;autoIncrement"`
	PatchID     uint64   `gorm:"column:patch_id"`
	ChampionID  uint64   `gorm:"column:champion_id"`
	Mode        string   `gorm:"column:mode"`
	SampleCount uint32   `gorm:"column:sample_count"`
	PickRate    *float64 `gorm:"column:pick_rate"`
	WinRate     *float64 `gorm:"column:win_rate"`
	Tier        *string  `gorm:"column:tier"`
	SourceID    *uint64  `gorm:"column:source_id"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (ChampionPatchStat) TableName() string { return "champion_patch_stats" }
