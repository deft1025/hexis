package schema

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

//go:embed 003_items.sql 004_item_profiles.sql 005_gameplay_guides.sql 006_champion_skills.sql 007_gameplay_guide_previews.sql
var itemSchema embed.FS

func MigrateItems(ctx context.Context, db *sql.DB) error {
	for _, name := range []string{"003_items.sql", "004_item_profiles.sql", "005_gameplay_guides.sql", "006_champion_skills.sql", "007_gameplay_guide_previews.sql"} {
		statement, err := itemSchema.ReadFile(name)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, string(statement)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}
