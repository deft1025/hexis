package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"hexis/internal/repository"
)

// --- Response types ---

type SourceResponse struct {
	Name        string  `json:"name"`
	URL         *string `json:"url"`
	License     *string `json:"license"`
	Attribution *string `json:"attribution"`
}

type StatsResponse struct {
	SampleCount int64          `json:"sample_count"`
	PickRate    *float64       `json:"pick_rate"`
	WinRate     *float64       `json:"win_rate"`
	Tier        *string        `json:"tier"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Source      SourceResponse `json:"source"`
}

type ChampionResponse struct {
	ID      string         `json:"id"`
	NameZh  string         `json:"name_zh"`
	NameEn  *string        `json:"name_en"`
	IconURL *string        `json:"icon_url"`
	Source  SourceResponse `json:"source"`
	Stats   *StatsResponse `json:"stats"`
}

type AugmentResponse struct {
	ID          string         `json:"id"`
	NameZh      string         `json:"name_zh"`
	NameEn      *string        `json:"name_en"`
	IconURL     *string        `json:"icon_url"`
	Description *string        `json:"description"`
	Rarity      *string        `json:"rarity"`
	Source      SourceResponse `json:"source"`
	Stats       *StatsResponse `json:"stats"`
}

type ComboChampionResponse struct {
	ID     string `json:"id"`
	NameZh string `json:"name_zh"`
}

type ComboResponse struct {
	Champions   []ComboChampionResponse `json:"champions"`
	SampleCount int64                   `json:"sample_count"`
	PickRate    *float64                `json:"pick_rate"`
	WinRate     *float64                `json:"win_rate"`
	UpdatedAt   time.Time               `json:"updated_at"`
	Source      SourceResponse          `json:"source"`
}

type PatchResponse struct {
	Version    string     `json:"version"`
	Status     string     `json:"status"`
	ReleasedAt *time.Time `json:"released_at,omitempty"`
}

// --- Conversion helpers ---

func nullStringPtr(v sql.NullString) *string {
	if v.Valid {
		return &v.String
	}
	return nil
}

func nullFloat64Ptr(v sql.NullFloat64) *float64 {
	if v.Valid {
		return &v.Float64
	}
	return nil
}

func sourceResponse(name string, url, license, attribution sql.NullString) SourceResponse {
	return SourceResponse{
		Name:        name,
		URL:         nullStringPtr(url),
		License:     nullStringPtr(license),
		Attribution: nullStringPtr(attribution),
	}
}

func statSourceResponse(name sql.NullString, url, license, attribution sql.NullString) SourceResponse {
	return SourceResponse{
		Name:        name.String,
		URL:         nullStringPtr(url),
		License:     nullStringPtr(license),
		Attribution: nullStringPtr(attribution),
	}
}

func buildStats(row repository.ChampionProfileRow) *StatsResponse {
	if !row.SampleCount.Valid {
		return nil
	}
	return &StatsResponse{
		SampleCount: row.SampleCount.Int64,
		PickRate:    nullFloat64Ptr(row.PickRate),
		WinRate:     nullFloat64Ptr(row.WinRate),
		Tier:        nullStringPtr(row.Tier),
		UpdatedAt:   row.UpdatedAt.Time,
		Source:      statSourceResponse(row.StatSourceName, row.StatSourceURL, row.StatSourceLicense, row.StatSourceAttribution),
	}
}

func buildAugmentStats(row repository.AugmentProfileRow) *StatsResponse {
	if !row.SampleCount.Valid {
		return nil
	}
	return &StatsResponse{
		SampleCount: row.SampleCount.Int64,
		PickRate:    nullFloat64Ptr(row.PickRate),
		WinRate:     nullFloat64Ptr(row.WinRate),
		Tier:        nullStringPtr(row.Tier),
		UpdatedAt:   row.UpdatedAt.Time,
		Source:      statSourceResponse(row.StatSourceName, row.StatSourceURL, row.StatSourceLicense, row.StatSourceAttribution),
	}
}

// --- Error helpers ---

func dbNotConfigured(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database is not configured"})
}

func queryFailed(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
}
