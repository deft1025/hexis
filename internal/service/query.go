package service

import (
	"context"
	"errors"
	"strconv"

	"gorm.io/gorm"
	"hexis/internal/model"
	"hexis/internal/repository"
)

type QueryService struct {
	patches *repository.PatchRepository
}

func NewQueryService(patches *repository.PatchRepository) *QueryService {
	return &QueryService{patches: patches}
}

// ResolvePatch returns the latest published patch, or a specific published
// version when version is non-empty.
func (s *QueryService) ResolvePatch(ctx context.Context, version string) (*model.Patch, error) {
	if s == nil || s.patches == nil {
		return nil, errDatabaseNotConfigured
	}
	return s.patches.LatestPublished(ctx, version)
}

var errDatabaseNotConfigured = errors.New("database is not configured")

// ParsePage validates and defaults limit/offset query parameters.
func ParsePage(limitStr, offsetStr string) (int, int, error) {
	limit, offset := 100, 0
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 || parsed > 100 {
			return 0, 0, errors.New("limit must be between 1 and 100")
		}
		limit = parsed
	}
	if offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil || parsed < 0 {
			return 0, 0, errors.New("offset must be non-negative")
		}
		offset = parsed
	}
	return limit, offset, nil
}

// IsRecordNotFound reports whether err is a gorm record-not-found error.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
