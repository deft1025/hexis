package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"gorm.io/gorm"
)

type ItemProfileRow struct {
	RiotID            string
	NameZh            string
	Description       string
	IconURL           string
	Mode              string
	Categories        json.RawMessage
	Price             int
	SellPrice         int
	Purchasable       bool
	Recipe            json.RawMessage
	Stats             json.RawMessage
	RequiredChampion  sql.NullString
	SourceName        string
	SourceURL         sql.NullString
	SourceLicense     sql.NullString
	SourceAttribution sql.NullString
}

type ItemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) *ItemRepository { return &ItemRepository{db: db} }

func (r *ItemRepository) ListProfiles(ctx context.Context, patchID int64, limit, offset int) ([]ItemProfileRow, error) {
	var rows []ItemProfileRow
	err := r.db.WithContext(ctx).Raw(`SELECT i.riot_id, p.name_zh, p.description, p.icon_url, p.mode,
		p.categories, p.price, p.sell_price, p.purchasable, p.recipe, p.stats, p.required_champion,
		s.name AS source_name, s.url AS source_url, s.license AS source_license, s.attribution AS source_attribution
		FROM item_patch_profiles p JOIN items i ON i.id = p.item_id
		JOIN data_sources s ON s.id = p.source_id
		WHERE p.patch_id = ? AND p.mode = 'KIWI' ORDER BY p.name_zh, i.riot_id LIMIT ? OFFSET ?`,
		patchID, limit, offset).Scan(&rows).Error
	return rows, err
}
