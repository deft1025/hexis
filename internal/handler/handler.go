package handler

import (
	"gorm.io/gorm"
	"hexis/internal/repository"
	"hexis/internal/service"
)

type Handler struct {
	db        *gorm.DB
	querySvc  *service.QueryService
	champions *repository.ChampionRepository
	augments  *repository.AugmentRepository
	combos    *repository.ComboRepository
}

func New(db *gorm.DB) *Handler {
	h := &Handler{db: db}
	if db != nil {
		patches := repository.NewPatchRepository(db)
		h.querySvc = service.NewQueryService(patches)
		h.champions = repository.NewChampionRepository(db)
		h.augments = repository.NewAugmentRepository(db)
		h.combos = repository.NewComboRepository(db)
	}
	return h
}
