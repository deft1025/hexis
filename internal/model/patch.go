package model

import "time"

type Patch struct {
	ID         uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	Version    string     `gorm:"column:version"`
	Status     string     `gorm:"column:status"`
	ReleasedAt *time.Time `gorm:"column:released_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (Patch) TableName() string { return "patches" }
