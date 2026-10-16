package repository

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"
	"hexis/internal/model"
)

type SkillRepository struct{ db *gorm.DB }

func NewSkillRepository(db *gorm.DB) *SkillRepository { return &SkillRepository{db: db} }

func (r *SkillRepository) Find(ctx context.Context, patchID int64, id string) (model.SkillsResult, error) {
	var row struct {
		Version   string
		Title     string
		Roles     json.RawMessage
		Skills    json.RawMessage
		SourceURL string
	}
	result := r.db.WithContext(ctx).Raw(`SELECT s.data_version AS version,s.title,s.roles,s.skills,s.source_url
		FROM champion_patch_skills s JOIN champions c ON c.id=s.champion_id JOIN champion_patch_profiles p ON p.patch_id=s.patch_id AND p.champion_id=s.champion_id
		WHERE s.patch_id=? AND c.riot_id=?`, patchID, id).Scan(&row)
	if result.Error != nil {
		return model.SkillsResult{}, result.Error
	}
	if result.RowsAffected == 0 {
		return model.SkillsResult{}, gorm.ErrRecordNotFound
	}
	p := model.SkillsResult{Version: row.Version, Title: row.Title, SourceURL: row.SourceURL}
	if err := json.Unmarshal(row.Roles, &p.Roles); err != nil {
		return p, err
	}
	if err := json.Unmarshal(row.Skills, &p.Skills); err != nil {
		return p, err
	}
	return p, nil
}
