package handler

import (
	"gorm.io/gorm"
	"hexis/internal/repository"
	"hexis/internal/service"
)

type Handler struct {
	db                  *gorm.DB
	querySvc            *service.QueryService
	champions           *repository.ChampionRepository
	augments            *repository.AugmentRepository
	combos              *repository.ComboRepository
	items               *repository.ItemRepository
	guides              *repository.GuideRepository
	skills              *repository.SkillRepository
	guidePreviewEnabled bool
}

func New(db *gorm.DB) *Handler {
	return NewWithGuidePreview(db, false)
}

func NewWithGuidePreview(db *gorm.DB, enabled bool) *Handler {
	h := &Handler{db: db, guidePreviewEnabled: enabled}
	if db != nil {
		patches := repository.NewPatchRepository(db)
		h.querySvc = service.NewQueryService(patches)
		h.champions = repository.NewChampionRepository(db)
		h.augments = repository.NewAugmentRepository(db)
		h.combos = repository.NewComboRepository(db)
		h.items = repository.NewItemRepository(db)
		h.guides = repository.NewGuideRepository(db)
		h.skills = repository.NewSkillRepository(db)
	}
	return h
}
