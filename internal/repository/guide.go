package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"gorm.io/gorm"
)

type GuideRow struct {
	Content           json.RawMessage
	SourceName        string
	SourceURL         sql.NullString
	SourceLicense     sql.NullString
	SourceAttribution sql.NullString
}

type GuideRepository struct{ db *gorm.DB }

func NewGuideRepository(db *gorm.DB) *GuideRepository { return &GuideRepository{db: db} }

func (r *GuideRepository) List(ctx context.Context, patchID int64, limit, offset int) ([]GuideRow, error) {
	return r.list(ctx, patchID, limit, offset, false)
}

func (r *GuideRepository) ListPreviews(ctx context.Context, patchID int64, limit, offset int) ([]GuideRow, error) {
	return r.list(ctx, patchID, limit, offset, true)
}

func (r *GuideRepository) list(ctx context.Context, patchID int64, limit, offset int, preview bool) ([]GuideRow, error) {
	table := "gameplay_guides"
	if preview {
		table = "gameplay_guide_previews"
	}
	var rows []GuideRow
	err := r.db.WithContext(ctx).Raw(`SELECT g.content,s.name AS source_name,s.url AS source_url,
		s.license AS source_license,s.attribution AS source_attribution FROM `+table+` g
		JOIN data_sources s ON s.id=g.source_id WHERE g.patch_id=? ORDER BY g.title,g.guide_key LIMIT ? OFFSET ?`, patchID, limit, offset).Scan(&rows).Error
	return rows, err
}
