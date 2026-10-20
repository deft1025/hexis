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
)

type guideBackupRow struct {
	Table       string          `json:"table"`
	Key         string          `json:"key"`
	Content     json.RawMessage `json:"content"`
	SourceID    int64           `json:"source_id"`
	Source      string          `json:"source"`
	URL         sql.NullString  `json:"source_url"`
	License     sql.NullString  `json:"source_license"`
	Attribution sql.NullString  `json:"source_attribution"`
}

func runBackupGuides(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("backup-guides", flag.ContinueOnError)
	patch := flags.String("patch", "", "existing patch to snapshot")
	out := flags.String("out", "", "new JSON file; refuses to overwrite an existing backup")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *patch == "" || *out == "" {
		return fmt.Errorf("backup-guides requires --patch and --out")
	}
	db, err := sql.Open("mysql", cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var patchID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM patches WHERE version=?`, *patch).Scan(&patchID); err != nil {
		return err
	}
	backup := struct {
		Patch     string           `json:"patch"`
		CreatedAt string           `json:"created_at"`
		Rows      []guideBackupRow `json:"rows"`
	}{Patch: *patch, CreatedAt: time.Now().UTC().Format(time.RFC3339), Rows: []guideBackupRow{}}
	for _, table := range []string{"gameplay_guides", "gameplay_guide_previews"} {
		rows, err := tx.QueryContext(ctx, `SELECT g.guide_key,g.content,g.source_id,s.name,s.url,s.license,s.attribution FROM `+table+` g JOIN data_sources s ON s.id=g.source_id WHERE g.patch_id=? ORDER BY g.guide_key`, patchID)
		if err != nil {
			return err
		}
		for rows.Next() {
			r := guideBackupRow{Table: table}
			if err := rows.Scan(&r.Key, &r.Content, &r.SourceID, &r.Source, &r.URL, &r.License, &r.Attribution); err != nil {
				rows.Close()
				return err
			}
			backup.Rows = append(backup.Rows, r)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(backup, "", "  ")
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
	_, err = f.Write(append(data, '\n'))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	fmt.Printf("backed up %d gameplay records and their sources for patch %s to %s; database unchanged\n", len(backup.Rows), *patch, *out)
	return nil
}
