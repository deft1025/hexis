package repository

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type ComboRow struct {
	FirstRiotID       string          `gorm:"column:first_riot_id"`
	FirstNameZh       string          `gorm:"column:first_name_zh"`
	SecondRiotID      string          `gorm:"column:second_riot_id"`
	SecondNameZh      string          `gorm:"column:second_name_zh"`
	SampleCount       int64           `gorm:"column:sample_count"`
	PickRate          sql.NullFloat64 `gorm:"column:pick_rate"`
	WinRate           sql.NullFloat64 `gorm:"column:win_rate"`
	SourceName        string          `gorm:"column:source_name"`
	SourceURL         sql.NullString  `gorm:"column:source_url"`
	SourceLicense     sql.NullString  `gorm:"column:source_license"`
	SourceAttribution sql.NullString  `gorm:"column:source_attribution"`
	UpdatedAt         time.Time       `gorm:"column:updated_at"`
}

const comboSQL = `SELECT a.riot_id AS first_riot_id, pa.name_zh AS first_name_zh, b.riot_id AS second_riot_id, pb.name_zh AS second_name_zh,
	c.sample_count AS sample_count, c.pick_rate AS pick_rate, c.win_rate AS win_rate,
	s.name AS source_name, s.url AS source_url, s.license AS source_license, s.attribution AS source_attribution,
	c.updated_at AS updated_at
	FROM champion_combo_stats c
	JOIN champions a ON a.id = c.champion_a_id
	JOIN champions b ON b.id = c.champion_b_id
	JOIN champion_patch_profiles pa ON pa.patch_id = c.patch_id AND pa.champion_id = a.id
	JOIN champion_patch_profiles pb ON pb.patch_id = c.patch_id AND pb.champion_id = b.id
	JOIN data_sources s ON s.id = c.source_id
	WHERE c.patch_id = ? ORDER BY c.sample_count DESC, a.riot_id, b.riot_id LIMIT ? OFFSET ?`

type ComboRepository struct {
	db *gorm.DB
}

func NewComboRepository(db *gorm.DB) *ComboRepository {
	return &ComboRepository{db: db}
}

func (r *ComboRepository) ListCombos(ctx context.Context, patchID int64, limit, offset int) ([]ComboRow, error) {
	var rows []ComboRow
	err := r.db.WithContext(ctx).Raw(comboSQL, patchID, limit, offset).Scan(&rows).Error
	return rows, err
}
