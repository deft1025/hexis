package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"hexis/internal/config"
	"hexis/internal/importer"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	// 加载 .env,供 --dsn 默认值使用
	cfg := config.Load()

	switch os.Args[1] {
	case "import":
		if err := runImport(cfg, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "publish":
		if err := runPublish(cfg, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func runImport(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("import", flag.ExitOnError)
	patch := flags.String("patch", "", "patch version, for example 15.21")
	kind := flags.String("kind", "champions", "champions, augments or combos")
	file := flags.String("file", "", "JSON input file")
	source := flags.String("source", "manual", "data source name")
	sourceURL := flags.String("source-url", "", "data source URL")
	sourceLicense := flags.String("source-license", "", "data source license")
	sourceAttribution := flags.String("source-attribution", "", "data source attribution")
	dsn := flags.String("dsn", cfg.Database.DSN(), "MySQL DSN")
	_ = flags.Parse(args)
	if *patch == "" || *file == "" || *dsn == "" {
		return fmt.Errorf("--patch, --file and --dsn are required")
	}
	if *kind != "champions" && *kind != "augments" && *kind != "combos" {
		return fmt.Errorf("unsupported --kind %q", *kind)
	}
	content, err := os.ReadFile(*file)
	if err != nil {
		return err
	}
	if !json.Valid(content) {
		return fmt.Errorf("input is not valid JSON")
	}
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return err
	}
	ctx := context.Background()
	count, err := importer.Import(ctx, db, importer.Options{
		Patch: *patch, Kind: *kind, Source: *source, SourceURL: *sourceURL,
		SourceLicense: *sourceLicense, SourceAttribution: *sourceAttribution, File: *file, Content: content,
	})
	if err != nil {
		return err
	}
	fmt.Printf("imported %d %s records for patch %s\n", count, *kind, *patch)
	return nil
}

func runPublish(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("publish", flag.ExitOnError)
	patch := flags.String("patch", "", "patch version")
	dsn := flags.String("dsn", cfg.Database.DSN(), "MySQL DSN")
	_ = flags.Parse(args)
	if *patch == "" || *dsn == "" {
		return fmt.Errorf("--patch and --dsn are required")
	}
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var patchID int64
	var status string
	if err := tx.QueryRowContext(ctx, `SELECT id, status FROM patches WHERE version = ? FOR UPDATE`, *patch).Scan(&patchID, &status); err != nil {
		return fmt.Errorf("patch %s: %w", *patch, err)
	}
	if status != "draft" && status != "collecting" {
		return fmt.Errorf("patch %s is %s and cannot be published", *patch, status)
	}
	for _, table := range []string{"champion_patch_profiles", "augment_patch_profiles"} {
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+table+` WHERE patch_id = ?`, patchID).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("patch %s has no %s", *patch, table)
		}
	}
	for _, table := range []string{"champion_patch_stats", "augment_patch_stats", "champion_combo_stats"} {
		var invalid int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+table+` WHERE patch_id = ? AND (sample_count = 0 OR pick_rate < 0 OR pick_rate > 1 OR win_rate < 0 OR win_rate > 1)`, patchID).Scan(&invalid); err != nil {
			return err
		}
		if invalid > 0 {
			return fmt.Errorf("patch %s has %d invalid %s rows", *patch, invalid, table)
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE patches SET status = 'published', released_at = NOW() WHERE id = ?`, patchID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Printf("published patch %s\n", *patch)
	return nil
}

func usage() {
	message := `usage:
  go run ./cmd/importer import --patch 15.21 --kind champions --file ./champions.json
  go run ./cmd/importer import --patch 15.21 --kind augments --file ./augments.json
  go run ./cmd/importer import --patch 15.21 --kind combos --file ./combos.json
  go run ./cmd/importer publish --patch 15.21`
	fmt.Println(strings.TrimSpace(message))
}
