package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) ChampionSkills(c *gin.Context) {
	if h.db == nil {
		dbNotConfigured(c)
		return
	}
	patch, ok := h.resolvePatch(c)
	if !ok {
		return
	}
	result, err := h.skills.Find(c.Request.Context(), int64(patch.ID), c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "champion skills not imported for this patch"})
		return
	}
	if err != nil {
		queryFailed(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"patch": patch.Version, "item": result})
}
