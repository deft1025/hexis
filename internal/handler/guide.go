package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"hexis/internal/importer"
	"hexis/internal/repository"
	"hexis/internal/service"
)

type GuideResponse struct {
	importer.Guide
	Source SourceResponse `json:"source"`
}

func (h *Handler) Guides(c *gin.Context) {
	preview := c.Query("preview") == "1"
	if preview && !h.guidePreviewEnabled {
		c.JSON(http.StatusNotFound, gin.H{"error": "guide preview is disabled"})
		return
	}
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
	var rows []repository.GuideRow
	if preview {
		rows, err = h.guides.ListPreviews(c.Request.Context(), int64(patch.ID), limit, offset)
	} else {
		rows, err = h.guides.List(c.Request.Context(), int64(patch.ID), limit, offset)
	}
	if err != nil {
		queryFailed(c)
		return
	}
	items := make([]GuideResponse, 0, len(rows))
	for _, row := range rows {
		var g importer.Guide
		if err := json.Unmarshal(row.Content, &g); err != nil {
			queryFailed(c)
			return
		}
		items = append(items, GuideResponse{Guide: g, Source: sourceResponse(row.SourceName, row.SourceURL, row.SourceLicense, row.SourceAttribution)})
	}
	c.JSON(http.StatusOK, gin.H{"patch": patch.Version, "items": items, "preview": preview})
}
