package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"hexis/internal/config"
	"hexis/internal/staticdata"
)

type augmentDescriptionChange struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	AugmentID int64          `json:"augment_id"`
	Before    sql.NullString `json:"before"`
	After     string         `json:"after"`
}

func runRepairAugmentDescriptions(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("repair-augment-descriptions", flag.ContinueOnError)
	dir := flags.String("dir", "", "existing pinned snapshot directory")
	out := flags.String("out", "", "new report/backup file; never overwrites")
	apply := flags.Bool("apply", false, "update only this patch's descriptions in a transaction")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *dir == "" || *out == "" {
		return fmt.Errorf("requires --dir and --out")
	}
	manifest, previous, profiles, unresolved, err := staticdata.RebuildAugments(*dir)
	if err != nil {
		return err
	}
	old := make(map[string]staticdata.Profile, len(previous))
	for _, row := range previous {
		if _, exists := old[row.ID]; exists {
			return fmt.Errorf("duplicate snapshot ID %s", row.ID)
		}
		old[row.ID] = row
	}
	db, err := sql.Open("mysql", cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: !*apply, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var patchID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM patches WHERE version=?`, manifest.Patch).Scan(&patchID); err != nil {
		return err
	}
	query := `SELECT a.riot_id,p.augment_id,p.name_zh,p.description FROM augment_patch_profiles p JOIN augments a ON a.id=p.augment_id WHERE p.patch_id=? ORDER BY a.riot_id`
	if *apply {
		query += ` FOR UPDATE`
	}
	rows, err := tx.QueryContext(ctx, query, patchID)
	if err != nil {
		return err
	}
	current := make(map[string]augmentDescriptionChange)
	for rows.Next() {
		var row augmentDescriptionChange
		if err := rows.Scan(&row.ID, &row.AugmentID, &row.Name, &row.Before); err != nil {
			rows.Close()
			return err
		}
		current[row.ID] = row
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if len(current) != len(profiles) {
		return fmt.Errorf("database and snapshot have different augment counts")
	}
	changes := make([]augmentDescriptionChange, 0)
	for _, profile := range profiles {
		original, existed := old[profile.ID]
		row, found := current[profile.ID]
		if !found || !existed || row.Name != profile.Name || original.Name != profile.Name || original.IconURL != profile.IconURL || original.Rarity != profile.Rarity {
			return fmt.Errorf("catalog mismatch for %s", profile.ID)
		}
		if row.Before.Valid && row.Before.String == profile.Description {
			continue
		}
		if !row.Before.Valid || row.Before.String != original.Description {
			return fmt.Errorf("description %s was edited since the snapshot; review it manually", profile.ID)
		}
		row.After = profile.Description
		changes = append(changes, row)
	}
	report := struct {
		Patch          string                           `json:"patch"`
		CreatedAt      string                           `json:"created_at"`
		ApplyRequested bool                             `json:"apply_requested"`
		Sources        []staticdata.SourceFile          `json:"sources"`
		Changes        []augmentDescriptionChange       `json:"changes"`
		Unresolved     []staticdata.VariableDescription `json:"unresolved"`
	}{manifest.Patch, time.Now().UTC().Format(time.RFC3339), *apply, manifest.Sources, changes, unresolved}
	content, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(append(content, '\n'))
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if *apply {
		for _, row := range changes {
			result, err := tx.ExecContext(ctx, `UPDATE augment_patch_profiles SET description=? WHERE patch_id=? AND augment_id=?`, row.After, patchID, row.AugmentID)
			if err != nil {
				return err
			}
			count, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if count != 1 {
				return fmt.Errorf("unexpected affected rows for %s", row.ID)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Printf("patch %s: %d descriptions changed, %d unresolved; applied=%t; before/after backup: %s\n", manifest.Patch, len(changes), len(unresolved), *apply, *out)
	return nil
}
