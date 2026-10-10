package repository

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type AugmentProfileRow struct {
	RiotID                string          `gorm:"column:riot_id"`
	NameZh                string          `gorm:"column:name_zh"`
	NameEn                sql.NullString  `gorm:"column:name_en"`
	IconURL               sql.NullString  `gorm:"column:icon_url"`
	SourceName            string          `gorm:"column:source_name"`
	SourceURL             sql.NullString  `gorm:"column:source_url"`
	SourceLicense         sql.NullString  `gorm:"column:source_license"`
	SourceAttribution     sql.NullString  `gorm:"column:source_attribution"`
	SampleCount           sql.NullInt64   `gorm:"column:sample_count"`
	PickRate              sql.NullFloat64 `gorm:"column:pick_rate"`
	WinRate               sql.NullFloat64 `gorm:"column:win_rate"`
	Tier                  sql.NullString  `gorm:"column:tier"`
	StatSourceName        sql.NullString  `gorm:"column:stat_source_name"`
	StatSourceURL         sql.NullString  `gorm:"column:stat_source_url"`
	StatSourceLicense     sql.NullString  `gorm:"column:stat_source_license"`
	StatSourceAttribution sql.NullString  `gorm:"column:stat_source_attribution"`
	UpdatedAt             sql.NullTime    `gorm:"column:updated_at"`
	Description           sql.NullString  `gorm:"column:description"`
	Rarity                sql.NullString  `gorm:"column:rarity"`
}

const augmentProfileSQL = `SELECT a.riot_id AS riot_id, p.name_zh AS name_zh, p.name_en AS name_en, p.icon_url AS icon_url,
	ps.name AS source_name, ps.url AS source_url, ps.license AS source_license, ps.attribution AS source_attribution,
	s.sample_count AS sample_count, s.pick_rate AS pick_rate, s.win_rate AS win_rate, s.tier AS tier,
	ss.name AS stat_source_name, ss.url AS stat_source_url, ss.license AS stat_source_license, ss.attribution AS stat_source_attribution,
	s.updated_at AS updated_at,
	p.description AS description, p.rarity AS rarity
	FROM augment_patch_profiles p
	JOIN augments a ON a.id = p.augment_id
	JOIN data_sources ps ON ps.id = p.source_id
	LEFT JOIN augment_patch_stats s ON s.patch_id = p.patch_id AND s.augment_id = p.augment_id
	LEFT JOIN data_sources ss ON ss.id = s.source_id
	WHERE p.patch_id = ? ORDER BY p.name_zh, a.riot_id LIMIT ? OFFSET ?`

type AugmentRepository struct {
	db *gorm.DB
}

func NewAugmentRepository(db *gorm.DB) *AugmentRepository {
	return &AugmentRepository{db: db}
}

func (r *AugmentRepository) ListProfiles(ctx context.Context, patchID int64, limit, offset int) ([]AugmentProfileRow, error) {
	var rows []AugmentProfileRow
	err := r.db.WithContext(ctx).Raw(augmentProfileSQL, patchID, limit, offset).Scan(&rows).Error
	return rows, err
}
