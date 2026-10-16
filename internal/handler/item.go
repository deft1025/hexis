package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"hexis/internal/service"
)

type ItemResponse struct {
	ID               string          `json:"id"`
	NameZh           string          `json:"name_zh"`
	Description      string          `json:"description"`
	IconURL          string          `json:"icon_url"`
	Mode             string          `json:"mode"`
	Categories       json.RawMessage `json:"categories"`
	Price            int             `json:"price"`
	SellPrice        int             `json:"sell_price"`
	Purchasable      bool            `json:"purchasable"`
	Recipe           json.RawMessage `json:"recipe"`
	Attributes       json.RawMessage `json:"attributes"`
	RequiredChampion *string         `json:"required_champion"`
	Source           SourceResponse  `json:"source"`
	Stats            *StatsResponse  `json:"stats"`
}

func (h *Handler) Items(c *gin.Context) {
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
	rows, err := h.items.ListProfiles(c.Request.Context(), int64(patch.ID), limit, offset)
	if err != nil {
		queryFailed(c)
		return
	}
	items := make([]ItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ItemResponse{ID: row.RiotID, NameZh: row.NameZh, Description: row.Description,
			IconURL: row.IconURL, Mode: row.Mode, Categories: row.Categories, Price: row.Price, SellPrice: row.SellPrice,
			Purchasable: row.Purchasable, Recipe: row.Recipe, Attributes: row.Stats,
			RequiredChampion: nullStringPtr(row.RequiredChampion),
			Source:           sourceResponse(row.SourceName, row.SourceURL, row.SourceLicense, row.SourceAttribution)})
	}
	c.JSON(http.StatusOK, gin.H{"patch": patch.Version, "items": items})
}
