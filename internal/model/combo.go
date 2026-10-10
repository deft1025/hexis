package model

import "time"

type ChampionComboStat struct {
	PatchID      uint64    `gorm:"column:patch_id;primaryKey"`
	ChampionAID  uint64    `gorm:"column:champion_a_id;primaryKey"`
	ChampionBID  uint64    `gorm:"column:champion_b_id;primaryKey"`
	SampleCount  uint32    `gorm:"column:sample_count"`
	PickRate     *float64  `gorm:"column:pick_rate"`
	WinRate      *float64  `gorm:"column:win_rate"`
	SourceID     uint64    `gorm:"column:source_id"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (ChampionComboStat) TableName() string { return "champion_combo_stats" }
