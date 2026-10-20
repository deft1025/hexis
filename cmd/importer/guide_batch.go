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
	"hexis/internal/importer"
)

func runGuideBatch(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("import-guide-batch", flag.ContinueOnError)
	file := flags.String("file", "", "batch manifest of edited preview files and individual sources")
	apply := flags.Bool("apply", false, "commit previews after validating every file; default rolls back all validation transactions")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *file == "" {
		return fmt.Errorf("import-guide-batch requires --file")
	}
	content, err := os.ReadFile(*file)
	if err != nil {
		return err
	}
	var batch struct {
		Patch string `json:"patch"`
		Files []struct {
			File        string `json:"file"`
			Source      string `json:"source"`
			URL         string `json:"source_url"`
			License     string `json:"source_license"`
			Attribution string `json:"source_attribution"`
		} `json:"files"`
	}
	if err := json.Unmarshal(content, &batch); err != nil {
		return err
	}
	if batch.Patch == "" || len(batch.Files) == 0 {
		return fmt.Errorf("batch requires patch and files")
	}
	options := []importer.Options{}
	seen := map[string]bool{}
	for _, f := range batch.Files {
		if f.File != filepath.Base(f.File) || f.Source == "" || f.URL == "" {
			return fmt.Errorf("batch file requires a basename, unique source name and URL")
		}
		filename := filepath.Join(filepath.Dir(*file), f.File)
		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		var guides []importer.Guide
		if err := json.Unmarshal(data, &guides); err != nil {
			return err
		}
		for _, g := range guides {
			if seen[g.ID] {
				return fmt.Errorf("duplicate batch guide ID: %s", g.ID)
			}
			seen[g.ID] = true
		}
		options = append(options, importer.Options{Patch: batch.Patch, Kind: "guide-previews", Source: f.Source, SourceURL: f.URL, SourceLicense: f.License, SourceAttribution: f.Attribution, File: filename, Content: data, DryRun: true})
	}
	db, err := sql.Open("mysql", cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	var patchID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM patches WHERE version=?`, batch.Patch).Scan(&patchID); err != nil {
		return fmt.Errorf("target patch must already exist: %w", err)
	}
	total := 0
	for i, o := range options {
		n, err := importer.Import(ctx, db, o)
		if err != nil {
			return fmt.Errorf("validation %s: %w; no files committed", o.File, err)
		}
		total += n
		if (i+1)%20 == 0 {
			fmt.Printf("validated %d/%d files\n", i+1, len(options))
		}
	}
	fmt.Printf("validated %d preview records in %d files for patch %s; validation transactions rolled back\n", total, len(options), batch.Patch)
	if !*apply {
		return nil
	}
	committed := 0
	for i, o := range options {
		o.DryRun = false
		n, err := importer.Import(ctx, db, o)
		if err != nil {
			return fmt.Errorf("import %s: %w; %d files and %d records committed; rerun after fixing (upsert)", o.File, err, i, committed)
		}
		committed += n
		if (i+1)%20 == 0 {
			fmt.Printf("imported %d/%d files\n", i+1, len(options))
		}
	}
	fmt.Printf("imported %d preview records; existing IDs updated, other records retained; no reviewed guides or catalogs changed\n", committed)
	return nil
}
