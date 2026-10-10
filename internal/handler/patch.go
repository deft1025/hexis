package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hexis/internal/service"
)

func (h *Handler) LatestPatch(c *gin.Context) {
	if h.db == nil {
		dbNotConfigured(c)
		return
	}
	patch, err := h.querySvc.ResolvePatch(c.Request.Context(), "")
	if err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no published patch"})
			return
		}
		queryFailed(c)
		return
	}
	c.JSON(http.StatusOK, PatchResponse{
		Version:    patch.Version,
		Status:     patch.Status,
		ReleasedAt: patch.ReleasedAt,
	})
}
