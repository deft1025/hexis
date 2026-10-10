package repository

import (
	"context"

	"gorm.io/gorm"
	"hexis/internal/model"
)

type PatchRepository struct {
	db *gorm.DB
}

func NewPatchRepository(db *gorm.DB) *PatchRepository {
	return &PatchRepository{db: db}
}

// LatestPublished returns the latest published patch, or a specific published
// version when version is non-empty.
func (r *PatchRepository) LatestPublished(ctx context.Context, version string) (*model.Patch, error) {
	query := r.db.WithContext(ctx).Where("status = ?", "published")
	if version != "" {
		query = query.Where("version = ?", version)
	} else {
		query = query.Order("released_at DESC, id DESC")
	}
	var patch model.Patch
	err := query.First(&patch).Error
	if err != nil {
		return nil, err
	}
	return &patch, nil
}
