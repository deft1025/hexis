package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"hexis/internal/config"
	"hexis/internal/importer"
	"hexis/internal/staticdata"
	schema "hexis/sql"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	// 加载 .env,供 --dsn 默认值使用
	cfg := config.Load()

	switch os.Args[1] {
	case "migrate":
		if err := runMigrate(cfg, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "sync":
		if err := runSync(cfg, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "fetch":
		if err := runFetch(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "backfill-skills":
		if err := runBackfillSkills(cfg, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "status":
		if err := runStatus(cfg, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
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

func runFetch(args []string) error {
	flags := flag.NewFlagSet("fetch", flag.ContinueOnError)
	version := flags.String("version", "", "pinned Data Dragon version, for example 16.20.1")
	out := flags.String("out", "", "output directory (default data/<patch>)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	patch, err := staticdata.PatchVersion(*version)
	if err != nil {
		return err
	}
	if *out == "" {
		*out = filepath.Join("data", patch)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	manifest, err := staticdata.Fetch(ctx, *version, *out)
	if err != nil {
		return err
	}
	fmt.Printf("downloaded patch %s: %d champions, %d KIWI augments, %d KIWI items, %d champion skill sets to %s\n", patch, manifest.Champions, manifest.Augments, manifest.Items, manifest.SkillChampions, *out)
	fmt.Printf("%d descriptions have variable or unresolved values; see manifest.json\n", len(manifest.VariableDescriptions))
	return nil
}

func runMigrate(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	dsn := flags.String("dsn", cfg.Database.DSN(), "MySQL DSN")
	if err := flags.Parse(args); err != nil {
		return err
	}
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := schema.MigrateItems(ctx, db); err != nil {
		return err
	}
	fmt.Println("item, gameplay guide and skill tables ready")
	return nil
}

func runBackfillSkills(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("backfill-skills", flag.ContinueOnError)
	version := flags.String("version", "", "pinned Data Dragon version matching existing hero profiles")
	if err := flags.Parse(args); err != nil {
		return err
	}
	patch, err := staticdata.PatchVersion(*version)
	if err != nil {
		return err
	}
	if err := runMigrate(cfg, nil); err != nil {
		return err
	}
	db, err := sql.Open("mysql", cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM champion_patch_skills s JOIN patches p ON p.id=s.patch_id WHERE p.version=?`, patch).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("patch %s already has skills; backfill only fills an empty skill snapshot", patch)
	}
	rows, err := db.QueryContext(ctx, `SELECT c.riot_id,p.icon_url FROM champion_patch_profiles p JOIN champions c ON c.id=p.champion_id JOIN patches v ON v.id=p.patch_id WHERE v.version=? ORDER BY c.riot_id`, patch)
	if err != nil {
		return err
	}
	ids := []string{}
	for rows.Next() {
		var id, icon string
		if err := rows.Scan(&id, &icon); err != nil {
			rows.Close()
			return err
		}
		if icon != "https://ddragon.leagueoflegends.com/cdn/"+*version+"/img/champion/"+id+".png" {
			rows.Close()
			return fmt.Errorf("hero %s does not use requested version %s", id, *version)
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	out := filepath.Join("data", patch)
	fmt.Printf("downloading skill sets for %d heroes in patch %s\n", len(ids), patch)
	profiles, sources, err := staticdata.FetchSkills(ctx, &http.Client{Timeout: 40 * time.Second}, *version, out, ids)
	if err != nil {
		return err
	}
	manifest, err := json.MarshalIndent(struct {
		Version   string                  `json:"version"`
		Champions int                     `json:"champions"`
		Sources   []staticdata.SourceFile `json:"sources"`
	}{*version, len(profiles), sources}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "skills-manifest.json"), append(manifest, '\n'), 0644); err != nil {
		return err
	}
	if err := runImport(cfg, []string{"--patch", patch, "--kind", "skills", "--file", filepath.Join(out, "skills.json"), "--source", "Riot Data Dragon champion skills " + *version, "--source-url", "https://ddragon.leagueoflegends.com/cdn/" + *version + "/data/zh_CN/champion.json", "--source-attribution", "Riot Games; zh_CN; pinned champion abilities"}); err != nil {
		return err
	}
	return nil
}

func runSync(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("sync", flag.ContinueOnError)
	version := flags.String("version", "", "pinned Data Dragon version")
	publish := flags.Bool("publish", false, "publish after all static catalogs import successfully")
	if err := flags.Parse(args); err != nil {
		return err
	}
	patch, err := staticdata.PatchVersion(*version)
	if err != nil {
		return err
	}
	db, err := sql.Open("mysql", cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var status string
	err = db.QueryRowContext(ctx, `SELECT status FROM patches WHERE version = ?`, patch).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil && status != "draft" && status != "collecting" {
		return fmt.Errorf("patch %s is %s; sync requires an unpublished version", patch, status)
	}
	if err := runMigrate(cfg, nil); err != nil {
		return err
	}
	if err := runFetch([]string{"--version", *version}); err != nil {
		return err
	}
	for _, catalog := range []struct{ kind, source, url string }{
		{"champions", "Riot Data Dragon " + *version, "https://ddragon.leagueoflegends.com/cdn/" + *version + "/data/zh_CN/champion.json"},
		{"augments", "CommunityDragon KIWI " + patch, "https://raw.communitydragon.org/" + patch + "/game/maps/modespecificdata/kiwi.bin.json"},
		{"items", "Riot Data Dragon + CommunityDragon KIWI items " + *version, "https://raw.communitydragon.org/" + patch + "/game/data/maps/shipping/map12/map12.bin.json"},
		{"skills", "Riot Data Dragon champion skills " + *version, "https://ddragon.leagueoflegends.com/cdn/" + *version + "/data/zh_CN/champion.json"},
	} {
		if err := runImport(cfg, []string{"--patch", patch, "--kind", catalog.kind,
			"--file", filepath.Join("data", patch, catalog.kind+".json"), "--source", catalog.source,
			"--source-url", catalog.url, "--source-attribution", "Riot Games assets; zh_CN; KIWI static data"}); err != nil {
			return fmt.Errorf("sync %s: %w", catalog.kind, err)
		}
	}
	if *publish {
		return runPublish(cfg, []string{"--patch", patch})
	}
	fmt.Printf("patch %s ready for review; publish with: go run ./cmd/importer publish --patch %s\n", patch, patch)
	return nil
}

func runStatus(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	dsn := flags.String("dsn", cfg.Database.DSN(), "MySQL DSN")
	if err := flags.Parse(args); err != nil {
		return err
	}
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, `SELECT p.version, p.status,
		(SELECT COUNT(*) FROM champion_patch_profiles WHERE patch_id = p.id),
		(SELECT COUNT(*) FROM augment_patch_profiles WHERE patch_id = p.id),
		(SELECT COUNT(*) FROM item_patch_profiles WHERE patch_id = p.id),
		(SELECT COUNT(*) FROM champion_patch_skills WHERE patch_id = p.id),
		(SELECT COUNT(*) FROM champion_patch_stats WHERE patch_id = p.id),
		(SELECT COUNT(*) FROM augment_patch_stats WHERE patch_id = p.id),
		(SELECT COUNT(*) FROM champion_combo_stats WHERE patch_id = p.id)
		FROM patches p ORDER BY p.id DESC`)
	if err != nil {
		return err
	}
	defer rows.Close()
	fmt.Println("patch\tstatus\tchampions\taugments\titems\tskill_sets\tchampion_stats\taugment_stats\tcombos")
	for rows.Next() {
		var version, status string
		var champions, augments, items, skillSets, championStats, augmentStats, combos int
		if err := rows.Scan(&version, &status, &champions, &augments, &items, &skillSets, &championStats, &augmentStats, &combos); err != nil {
			return err
		}
		fmt.Printf("%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n", version, status, champions, augments, items, skillSets, championStats, augmentStats, combos)
	}
	return rows.Err()
}

func runImport(cfg *config.Config, args []string) error {
	flags := flag.NewFlagSet("import", flag.ExitOnError)
	patch := flags.String("patch", "", "patch version, for example 15.21")
	kind := flags.String("kind", "champions", "champions, augments, items, skills, guides, guide-previews or combos")
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
	if *kind != "champions" && *kind != "augments" && *kind != "combos" && *kind != "items" && *kind != "guides" && *kind != "guide-previews" && *kind != "skills" {
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
	for _, table := range []string{"champion_patch_profiles", "augment_patch_profiles", "item_patch_profiles"} {
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+table+` WHERE patch_id = ?`, patchID).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("patch %s has no %s", *patch, table)
		}
	}
	var missingSkills int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM champion_patch_profiles p LEFT JOIN champion_patch_skills s ON s.patch_id=p.patch_id AND s.champion_id=p.champion_id WHERE p.patch_id=? AND s.champion_id IS NULL`, patchID).Scan(&missingSkills); err != nil {
		return err
	}
	if missingSkills > 0 {
		return fmt.Errorf("patch %s has %d heroes without imported skills", *patch, missingSkills)
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
  go run ./cmd/importer migrate
  go run ./cmd/importer sync --version 16.21.1
  go run ./cmd/importer fetch --version 16.20.1
  go run ./cmd/importer status
  go run ./cmd/importer backfill-skills --version 16.20.1
  go run ./cmd/importer import --patch 15.21 --kind champions --file ./champions.json
  go run ./cmd/importer import --patch 15.21 --kind augments --file ./augments.json
  go run ./cmd/importer import --patch 16.20 --kind items --file ./data/16.20/items.json
  go run ./cmd/importer import --patch 16.20 --kind guides --file ./data/16.20/guides.json
  go run ./cmd/importer import --patch 15.21 --kind combos --file ./combos.json
  go run ./cmd/importer publish --patch 15.21`
	fmt.Println(strings.TrimSpace(message))
}
