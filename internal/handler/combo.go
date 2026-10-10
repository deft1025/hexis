package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hexis/internal/service"
)

func (h *Handler) Combos(c *gin.Context) {
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
	rows, err := h.combos.ListCombos(c.Request.Context(), int64(patch.ID), limit, offset)
	if err != nil {
		queryFailed(c)
		return
	}
	items := make([]ComboResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ComboResponse{
			Champions: []ComboChampionResponse{
				{ID: row.FirstRiotID, NameZh: row.FirstNameZh},
				{ID: row.SecondRiotID, NameZh: row.SecondNameZh},
			},
			SampleCount: row.SampleCount,
			PickRate:    nullFloat64Ptr(row.PickRate),
			WinRate:     nullFloat64Ptr(row.WinRate),
			UpdatedAt:   row.UpdatedAt,
			Source:      sourceResponse(row.SourceName, row.SourceURL, row.SourceLicense, row.SourceAttribution),
		})
	}
	c.JSON(http.StatusOK, gin.H{"patch": patch.Version, "items": items})
}
