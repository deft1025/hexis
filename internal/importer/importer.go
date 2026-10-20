package importer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type Options struct {
	Patch             string
	Kind              string
	Source            string
	SourceURL         string
	SourceLicense     string
	SourceAttribution string
	File              string
	Content           []byte
	DryRun            bool
}

func Import(ctx context.Context, db *sql.DB, options Options) (int, error) {
	if options.Patch == "" || options.Source == "" {
		return 0, fmt.Errorf("patch and source are required")
	}
	if options.Kind != "champions" && options.Kind != "augments" && options.Kind != "combos" && options.Kind != "items" && options.Kind != "guides" && options.Kind != "guide-previews" && options.Kind != "skills" {
		return 0, fmt.Errorf("unsupported kind %q", options.Kind)
	}
	var raw any
	if err := json.Unmarshal(options.Content, &raw); err != nil {
		return 0, err
	}
	records, err := records(raw)
	if err != nil {
		return 0, err
	}
	if len(records) == 0 {
		return 0, fmt.Errorf("input contains no records")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var patchID int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM patches WHERE version = ?`, options.Patch).Scan(&patchID)
	if err == sql.ErrNoRows {
		result, insertErr := tx.ExecContext(ctx, `INSERT INTO patches (version, status) VALUES (?, 'collecting')`, options.Patch)
		if insertErr != nil {
			return 0, insertErr
		}
		patchID, err = result.LastInsertId()
	}
	if err != nil {
		return 0, err
	}
	var status string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM patches WHERE id = ? FOR UPDATE`, patchID).Scan(&status); err != nil {
		return 0, err
	}
	if status == "published" && (options.Kind == "guides" || options.Kind == "guide-previews") {
		// Editorial content can evolve independently of static snapshots.
	} else if status == "published" && (options.Kind == "items" || options.Kind == "skills") {
		var count int
		table := "item_patch_profiles"
		if options.Kind == "skills" {
			table = "champion_patch_skills"
		}
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+table+` WHERE patch_id = ?`, patchID).Scan(&count); err != nil {
			return 0, err
		}
		if count > 0 {
			return 0, fmt.Errorf("published patch %s already has %s and cannot be overwritten", options.Patch, table)
		}
	} else if status != "collecting" && status != "draft" {
		return 0, fmt.Errorf("patch %s is %s and cannot be imported", options.Patch, status)
	}

	sourceID, err := ensureSource(ctx, tx, options)
	if err != nil {
		return 0, err
	}
	fileHash := sha256.Sum256(options.Content)
	fileInfo, _ := os.Stat(options.File)
	var fileSize int64
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO import_jobs
		(source_id, patch_id, kind, file_path, file_hash, file_size, status, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?, 'success', NOW(), NOW())`,
		sourceID, patchID, options.Kind, options.File, hex.EncodeToString(fileHash[:]), fileSize)
	if err != nil {
		return 0, err
	}
	count := 0
	if options.Kind == "skills" {
		var expected int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM champion_patch_profiles WHERE patch_id=?`, patchID).Scan(&expected); err != nil {
			return 0, err
		}
		seen := make(map[string]bool)
		for _, record := range records {
			id := stringValue(record, "id")
			if id == "" || seen[id] {
				return 0, fmt.Errorf("duplicate or empty skill champion ID")
			}
			seen[id] = true
		}
		if expected != len(records) {
			return 0, fmt.Errorf("skills import must cover all %d heroes in this patch", expected)
		}
	}
	for _, record := range records {
		if err := importRecord(ctx, tx, patchID, sourceID, options.Kind, record); err != nil {
			return 0, err
		}
		count++
	}
	if options.DryRun {
		return count, tx.Rollback()
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

func records(raw any) ([]map[string]any, error) {
	if object, ok := raw.(map[string]any); ok {
		for _, key := range []string{"data", "records", "items"} {
			if nested, exists := object[key]; exists {
				raw = nested
				break
			}
		}
	}
	array, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("input must be a JSON array or an object containing data/records/items")
	}
	result := make([]map[string]any, 0, len(array))
	for _, item := range array {
		object, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("each record must be a JSON object")
		}
		result = append(result, object)
	}
	return result, nil
}

func importRecord(ctx context.Context, tx *sql.Tx, patchID, sourceID int64, kind string, record map[string]any) error {
	if kind == "guide-previews" {
		return importGuideContent(ctx, tx, patchID, sourceID, record, true)
	}
	if kind == "skills" {
		return importSkills(ctx, tx, patchID, sourceID, record)
	}
	if kind == "guides" {
		return importGuide(ctx, tx, patchID, sourceID, record)
	}
	if kind == "items" {
		return importItem(ctx, tx, patchID, sourceID, record)
	}
	if err := validateStats(record); err != nil {
		return err
	}
	if kind == "combos" {
		if intValue(record, "sample_count", "samples") == 0 {
			return fmt.Errorf("combo requires a positive sample_count")
		}
		return importCombo(ctx, tx, patchID, sourceID, record)
	}
	rid := stringValue(record, "riot_id", "id", "key")
	name := stringValue(record, "name_zh", "name", "title")
	if rid == "" || name == "" {
		return fmt.Errorf("record requires id and name")
	}
	if kind == "champions" {
		_, err := tx.ExecContext(ctx, `INSERT INTO champions (riot_id, name_zh, name_en, icon_url)
			VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE name_zh=VALUES(name_zh), name_en=VALUES(name_en), icon_url=VALUES(icon_url)`,
			rid, name, stringValue(record, "name_en"), stringValue(record, "icon_url", "icon"))
		if err != nil {
			return err
		}
		var championID int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM champions WHERE riot_id = ?`, rid).Scan(&championID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO champion_patch_profiles (patch_id, champion_id, name_zh, name_en, icon_url, source_id)
			VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE name_zh=VALUES(name_zh), name_en=VALUES(name_en), icon_url=VALUES(icon_url), source_id=VALUES(source_id)`,
			patchID, championID, name, stringValue(record, "name_en"), stringValue(record, "icon_url", "icon"), sourceID)
		if err != nil {
			return err
		}
		if hasStat(record) {
			return importChampionStat(ctx, tx, patchID, sourceID, championID, record)
		}
		return nil
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO augments (riot_id, name_zh, name_en, description, rarity, icon_url)
		VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE name_zh=VALUES(name_zh), name_en=VALUES(name_en), description=VALUES(description), rarity=VALUES(rarity), icon_url=VALUES(icon_url)`,
		rid, name, stringValue(record, "name_en"), stringValue(record, "description", "desc"), stringValue(record, "rarity", "tier"), stringValue(record, "icon_url", "icon"))
	if err != nil {
		return err
	}
	var augmentID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM augments WHERE riot_id = ?`, rid).Scan(&augmentID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO augment_patch_profiles (patch_id, augment_id, name_zh, name_en, description, rarity, icon_url, source_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE name_zh=VALUES(name_zh), name_en=VALUES(name_en), description=VALUES(description), rarity=VALUES(rarity), icon_url=VALUES(icon_url), source_id=VALUES(source_id)`,
		patchID, augmentID, name, stringValue(record, "name_en"), stringValue(record, "description", "desc"), stringValue(record, "rarity", "tier"), stringValue(record, "icon_url", "icon"), sourceID)
	if err != nil {
		return err
	}
	if hasStat(record) {
		return importAugmentStat(ctx, tx, patchID, sourceID, augmentID, record)
	}
	return nil
}

func importCombo(ctx context.Context, tx *sql.Tx, patchID, sourceID int64, record map[string]any) error {
	values, ok := record["champion_ids"].([]any)
	if !ok || len(values) != 2 {
		return fmt.Errorf("combo requires exactly two champion_ids")
	}
	ids := make([]string, 2)
	for index, value := range values {
		ids[index], ok = value.(string)
		if !ok || ids[index] == "" {
			return fmt.Errorf("combo champion_ids must be strings")
		}
	}
	if ids[0] == ids[1] {
		return fmt.Errorf("combo champions must differ")
	}
	var championIDs [2]int64
	for index, riotID := range ids {
		if err := tx.QueryRowContext(ctx, `SELECT id FROM champions WHERE riot_id = ?`, riotID).Scan(&championIDs[index]); err != nil {
			return fmt.Errorf("combo champion %s: %w", riotID, err)
		}
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT 1 FROM champion_patch_profiles WHERE patch_id = ? AND champion_id = ?`, patchID, championIDs[index]).Scan(&exists); err != nil {
			return fmt.Errorf("combo champion %s is missing from patch: %w", riotID, err)
		}
	}
	sort.Slice(championIDs[:], func(left, right int) bool { return championIDs[left] < championIDs[right] })
	_, err := tx.ExecContext(ctx, `INSERT INTO champion_combo_stats
		(patch_id, champion_a_id, champion_b_id, sample_count, pick_rate, win_rate, source_id)
		VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE sample_count=VALUES(sample_count), pick_rate=VALUES(pick_rate), win_rate=VALUES(win_rate), source_id=VALUES(source_id)`,
		patchID, championIDs[0], championIDs[1], intValue(record, "sample_count", "samples"), nullableFloat(record, "pick_rate", "pickRate"), nullableFloat(record, "win_rate", "winRate"), sourceID)
	return err
}

func validateStats(record map[string]any) error {
	for _, key := range []string{"sample_count", "samples"} {
		if value, exists := record[key]; exists {
			number, ok := value.(float64)
			if !ok || number < 0 || number != float64(int64(number)) || number > 2147483647 {
				return fmt.Errorf("%s must be a non-negative integer", key)
			}
		}
	}
	for _, key := range []string{"pick_rate", "pickRate", "win_rate", "winRate"} {
		if value, exists := record[key]; exists && value != nil {
			number, ok := value.(float64)
			if !ok || number < 0 || number > 1 {
				return fmt.Errorf("%s must be between 0 and 1", key)
			}
		}
	}
	if hasStat(record) && intValue(record, "sample_count", "samples") == 0 {
		return fmt.Errorf("statistics require a positive sample_count")
	}
	return nil
}

func hasStat(record map[string]any) bool {
	for _, key := range []string{"sample_count", "samples", "pick_rate", "pickRate", "win_rate", "winRate"} {
		if _, ok := record[key]; ok {
			return true
		}
	}
	return false
}

func importChampionStat(ctx context.Context, tx *sql.Tx, patchID, sourceID, championID int64, record map[string]any) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO champion_patch_stats (patch_id, champion_id, sample_count, pick_rate, win_rate, tier, source_id)
		VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE sample_count=VALUES(sample_count), pick_rate=VALUES(pick_rate), win_rate=VALUES(win_rate), tier=VALUES(tier), source_id=VALUES(source_id)`,
		patchID, championID, intValue(record, "sample_count", "samples"), nullableFloat(record, "pick_rate", "pickRate"), nullableFloat(record, "win_rate", "winRate"), stringValue(record, "tier"), sourceID)
	return err
}

func importAugmentStat(ctx context.Context, tx *sql.Tx, patchID, sourceID, augmentID int64, record map[string]any) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO augment_patch_stats (patch_id, augment_id, sample_count, pick_rate, win_rate, tier, source_id)
		VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE sample_count=VALUES(sample_count), pick_rate=VALUES(pick_rate), win_rate=VALUES(win_rate), tier=VALUES(tier), source_id=VALUES(source_id)`,
		patchID, augmentID, intValue(record, "sample_count", "samples"), nullableFloat(record, "pick_rate", "pickRate"), nullableFloat(record, "win_rate", "winRate"), stringValue(record, "tier"), sourceID)
	return err
}

func ensureSource(ctx context.Context, tx *sql.Tx, options Options) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM data_sources WHERE name = ? LIMIT 1`, options.Source).Scan(&id)
	if err == nil {
		if options.SourceURL != "" || options.SourceLicense != "" || options.SourceAttribution != "" {
			_, err = tx.ExecContext(ctx, `UPDATE data_sources SET url = COALESCE(NULLIF(?, ''), url), license = COALESCE(NULLIF(?, ''), license), attribution = COALESCE(NULLIF(?, ''), attribution) WHERE id = ?`,
				options.SourceURL, options.SourceLicense, options.SourceAttribution, id)
			if err != nil {
				return 0, err
			}
		}
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO data_sources (name, url, license, attribution) VALUES (?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''))`,
		options.Source, options.SourceURL, options.SourceLicense, options.SourceAttribution)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func stringValue(record map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := record[key].(string); ok {
			return value
		}
	}
	return ""
}

func intValue(record map[string]any, keys ...string) int {
	for _, key := range keys {
		if value, ok := record[key].(float64); ok {
			return int(value)
		}
	}
	return 0
}

func nullableFloat(record map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := record[key].(float64); ok {
			return value
		}
	}
	return nil
}
