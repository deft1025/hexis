package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hexis/internal/service"
)

func (h *Handler) Augments(c *gin.Context) {
	if h.db == nil {
		dbNotConfigured(c)
		return
	}
	patch, ok := h.resolvePatch(c)
	if !ok {
		return
	}
	limit, offset, err := service.ParsePage(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rows, err := h.augments.ListProfiles(c.Request.Context(), int64(patch.ID), limit, offset)
	if err != nil {
		queryFailed(c)
		return
	}
	items := make([]AugmentResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, AugmentResponse{
			ID:          row.RiotID,
			NameZh:      row.NameZh,
			NameEn:      nullStringPtr(row.NameEn),
			IconURL:     nullStringPtr(row.IconURL),
			Description: nullStringPtr(row.Description),
			Rarity:      nullStringPtr(row.Rarity),
			Source:      sourceResponse(row.SourceName, row.SourceURL, row.SourceLicense, row.SourceAttribution),
			Stats:       buildAugmentStats(row),
		})
	}
	c.JSON(http.StatusOK, gin.H{"patch": patch.Version, "items": items})
}
