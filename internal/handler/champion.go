package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hexis/internal/model"
	"hexis/internal/service"
)

func (h *Handler) Champions(c *gin.Context) {
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
	rows, err := h.champions.ListProfiles(c.Request.Context(), int64(patch.ID), limit, offset)
	if err != nil {
		queryFailed(c)
		return
	}
	items := make([]ChampionResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ChampionResponse{
			ID:      row.RiotID,
			NameZh:  row.NameZh,
			NameEn:  nullStringPtr(row.NameEn),
			IconURL: nullStringPtr(row.IconURL),
			Source:  sourceResponse(row.SourceName, row.SourceURL, row.SourceLicense, row.SourceAttribution),
			Stats:   buildStats(row),
		})
	}
	c.JSON(http.StatusOK, gin.H{"patch": patch.Version, "items": items})
}

func (h *Handler) Champion(c *gin.Context) {
	if h.db == nil {
		dbNotConfigured(c)
		return
	}
	patch, ok := h.resolvePatch(c)
	if !ok {
		return
	}
	rows, err := h.champions.FindProfile(c.Request.Context(), int64(patch.ID), c.Param("id"))
	if err != nil {
		queryFailed(c)
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "champion not found"})
		return
	}
	row := rows[0]
	c.JSON(http.StatusOK, gin.H{
		"patch": patch.Version,
		"item": ChampionResponse{
			ID:      row.RiotID,
			NameZh:  row.NameZh,
			NameEn:  nullStringPtr(row.NameEn),
			IconURL: nullStringPtr(row.IconURL),
			Source:  sourceResponse(row.SourceName, row.SourceURL, row.SourceLicense, row.SourceAttribution),
			Stats:   buildStats(row),
		},
	})
}

// resolvePatch resolves the published patch from the "patch" query parameter
// (empty = latest) and writes an error response on failure.
func (h *Handler) resolvePatch(c *gin.Context) (*model.Patch, bool) {
	patch, err := h.querySvc.ResolvePatch(c.Request.Context(), c.Query("patch"))
	if err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "published patch not found"})
			return nil, false
		}
		queryFailed(c)
		return nil, false
	}
	return patch, true
}
