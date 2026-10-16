package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type GuideRecommendation struct {
	ID       string `json:"id"`
	Priority string `json:"priority"`
	Reason   string `json:"reason"`
}

type GuideStage struct {
	Label   string   `json:"label"`
	ItemIDs []string `json:"item_ids"`
	Reason  string   `json:"reason"`
}

type GuideSection struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type GuideReference struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Note string `json:"note,omitempty"`
}

type Guide struct {
	ID                    string                `json:"id"`
	Title                 string                `json:"title"`
	Category              string                `json:"category"`
	PlaystyleName         string                `json:"playstyle_name,omitempty"`
	Summary               string                `json:"summary"`
	ChampionIDs           []string              `json:"champion_ids"`
	Tags                  []string              `json:"tags"`
	CoreAugmentIDs        []string              `json:"core_augment_ids"`
	AlternativeAugmentIDs []string              `json:"alternative_augment_ids"`
	Recommendations       []GuideRecommendation `json:"recommendations"`
	Stages                []GuideStage          `json:"stages"`
	Sections              []GuideSection        `json:"sections"`
	BasicSections         []GuideSection        `json:"basic_sections,omitempty"`
	References            []GuideReference      `json:"references,omitempty"`
	Reviewed              bool                  `json:"reviewed"`
}

var guideKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,100}$`)

func validateGuide(g Guide) error {
	return validateGuideContent(g, false)
}

func validateGuideContent(g Guide, preview bool) error {
	if !guideKeyPattern.MatchString(g.ID) || strings.TrimSpace(g.Title) == "" || len([]rune(g.Title)) > 200 || strings.TrimSpace(g.Summary) == "" {
		return fmt.Errorf("guide requires id, title (up to 200 characters) and summary")
	}
	if g.Category != "fun" && g.Category != "basic" && g.Category != "general" {
		return fmt.Errorf("guide category must be fun, basic or general")
	}
	if g.PlaystyleName != "" && (g.Category != "fun" || strings.TrimSpace(g.PlaystyleName) == "" || len([]rune(g.PlaystyleName)) > 80) {
		return fmt.Errorf("playstyle_name is only allowed for fun guides, up to 80 characters")
	}
	if len(g.BasicSections) > 0 && g.Category != "fun" {
		return fmt.Errorf("basic_sections belongs to a fun guide; use sections for hero basics")
	}
	if !preview && !g.Reviewed {
		return fmt.Errorf("guide must be reviewed before import")
	}
	if preview && g.Reviewed {
		return fmt.Errorf("preview guide must remain unreviewed")
	}
	if g.Category != "general" && len(g.ChampionIDs) == 0 {
		return fmt.Errorf("hero guide requires champion_ids")
	}
	if len(g.Sections) == 0 || (!preview && len(g.Stages) == 0) {
		return fmt.Errorf("guide requires gameplay sections and equipment stages")
	}
	for _, s := range append(append([]GuideSection{}, g.Sections...), g.BasicSections...) {
		if strings.TrimSpace(s.Title) == "" || strings.TrimSpace(s.Body) == "" {
			return fmt.Errorf("guide section requires title and body")
		}
	}
	for _, s := range g.Stages {
		if s.Label == "" || len(s.ItemIDs) == 0 || strings.TrimSpace(s.Reason) == "" {
			return fmt.Errorf("guide stage requires label, items and reason")
		}
	}
	seen := make(map[string]bool)
	for _, ref := range g.References {
		u, err := url.Parse(ref.URL)
		if strings.TrimSpace(ref.Name) == "" || err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
			return fmt.Errorf("guide reference requires name and http(s) URL")
		}
	}
	for _, r := range g.Recommendations {
		if r.ID == "" || (r.Priority != "core" && r.Priority != "alternative" && r.Priority != "situational") || strings.TrimSpace(r.Reason) == "" || seen[r.ID] {
			return fmt.Errorf("guide recommendation requires unique id, priority and reason")
		}
		seen[r.ID] = true
	}
	return nil
}

func importGuide(ctx context.Context, tx *sql.Tx, patchID, sourceID int64, record map[string]any) error {
	return importGuideContent(ctx, tx, patchID, sourceID, record, false)
}

func importGuideContent(ctx context.Context, tx *sql.Tx, patchID, sourceID int64, record map[string]any, preview bool) error {
	if hasStat(record) {
		return fmt.Errorf("editorial guides cannot contain match statistics")
	}
	content, err := json.Marshal(record)
	if err != nil {
		return err
	}
	var g Guide
	if err := json.Unmarshal(content, &g); err != nil {
		return err
	}
	if err := validateGuideContent(g, preview); err != nil {
		return err
	}
	check := func(kind, id string) error {
		queries := map[string]string{
			"champion": `SELECT COUNT(*) FROM champion_patch_profiles p JOIN champions c ON c.id=p.champion_id WHERE p.patch_id=? AND c.riot_id=?`,
			"augment":  `SELECT COUNT(*) FROM augment_patch_profiles p JOIN augments a ON a.id=p.augment_id WHERE p.patch_id=? AND a.riot_id=?`,
			"item":     `SELECT COUNT(*) FROM item_patch_profiles p JOIN items i ON i.id=p.item_id WHERE p.patch_id=? AND p.mode='KIWI' AND i.riot_id=?`,
		}
		var count int
		if err := tx.QueryRowContext(ctx, queries[kind], patchID, id).Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("guide %s references %s %q missing in this version", g.ID, kind, id)
		}
		return nil
	}
	for _, id := range g.ChampionIDs {
		if err := check("champion", id); err != nil {
			return err
		}
	}
	augmentIDs := append(append([]string{}, g.CoreAugmentIDs...), g.AlternativeAugmentIDs...)
	for _, r := range g.Recommendations {
		augmentIDs = append(augmentIDs, r.ID)
	}
	for _, id := range augmentIDs {
		if err := check("augment", id); err != nil {
			return err
		}
	}
	for _, stage := range g.Stages {
		for _, id := range stage.ItemIDs {
			if err := check("item", id); err != nil {
				return err
			}
		}
	}
	content, err = json.Marshal(g)
	if err != nil {
		return err
	}
	table := "gameplay_guides"
	if preview {
		table = "gameplay_guide_previews"
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO `+table+` (patch_id,guide_key,title,category,content,source_id) VALUES (?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE title=VALUES(title),category=VALUES(category),content=VALUES(content),source_id=VALUES(source_id)`,
		patchID, g.ID, g.Title, g.Category, string(content), sourceID)
	return err
}
