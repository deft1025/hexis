package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

func validateItem(record map[string]any) error {
	id := stringValue(record, "id", "riot_id")
	if value, err := strconv.ParseUint(id, 10, 32); err != nil || value == 0 || stringValue(record, "name", "name_zh") == "" || stringValue(record, "icon_url") == "" {
		return fmt.Errorf("item requires a numeric string id, name and icon_url")
	}
	if stringValue(record, "mode") != "KIWI" {
		return fmt.Errorf("item mode must be KIWI")
	}
	if hasStat(record) {
		return fmt.Errorf("item profiles cannot contain match statistics")
	}
	for _, key := range []string{"price", "sell_price"} {
		number, ok := record[key].(float64)
		if !ok || number < 0 || number > 2147483647 || number != float64(int64(number)) {
			return fmt.Errorf("item %s must be a non-negative integer", key)
		}
	}
	if _, ok := record["purchasable"].(bool); !ok {
		return fmt.Errorf("item purchasable must be a boolean")
	}
	for _, key := range []string{"categories", "recipe"} {
		values, ok := record[key].([]any)
		if !ok || (key == "categories" && len(values) == 0) {
			return fmt.Errorf("item %s must be an array", key)
		}
		for _, value := range values {
			text, ok := value.(string)
			if !ok || text == "" {
				return fmt.Errorf("item %s must contain nonempty strings", key)
			}
			if key == "recipe" {
				if id, err := strconv.ParseUint(text, 10, 32); err != nil || id == 0 {
					return fmt.Errorf("item recipe must contain numeric string ids")
				}
			}
		}
	}
	stats, ok := record["stats"].(map[string]any)
	if !ok {
		return fmt.Errorf("item stats must be an object containing static attributes")
	}
	for _, value := range stats {
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("item stats must contain numeric attributes")
		}
	}
	return nil
}

func importItem(ctx context.Context, tx *sql.Tx, patchID, sourceID int64, record map[string]any) error {
	if err := validateItem(record); err != nil {
		return err
	}
	rid := stringValue(record, "id", "riot_id")
	if _, err := tx.ExecContext(ctx, `INSERT INTO items (riot_id) VALUES (?) ON DUPLICATE KEY UPDATE riot_id=VALUES(riot_id)`, rid); err != nil {
		return err
	}
	var itemID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM items WHERE riot_id = ?`, rid).Scan(&itemID); err != nil {
		return err
	}
	categories, _ := json.Marshal(record["categories"])
	recipe, _ := json.Marshal(record["recipe"])
	stats, _ := json.Marshal(record["stats"])
	_, err := tx.ExecContext(ctx, `INSERT INTO item_patch_profiles
		(patch_id, item_id, mode, name_zh, description, icon_url, categories, price, sell_price, purchasable, recipe, stats, required_champion, source_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?)
		ON DUPLICATE KEY UPDATE name_zh=VALUES(name_zh), description=VALUES(description), icon_url=VALUES(icon_url),
		categories=VALUES(categories), price=VALUES(price), sell_price=VALUES(sell_price), purchasable=VALUES(purchasable),
		recipe=VALUES(recipe), stats=VALUES(stats), required_champion=VALUES(required_champion), source_id=VALUES(source_id)`,
		patchID, itemID, "KIWI", stringValue(record, "name", "name_zh"), stringValue(record, "description"),
		stringValue(record, "icon_url"), string(categories), intValue(record, "price"), intValue(record, "sell_price"),
		record["purchasable"], string(recipe), string(stats), stringValue(record, "required_champion"), sourceID)
	return err
}
