package model

import "time"

type DataSource struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string    `gorm:"column:name"`
	URL         *string   `gorm:"column:url"`
	License     *string   `gorm:"column:license"`
	Attribution *string   `gorm:"column:attribution"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (DataSource) TableName() string { return "data_sources" }
