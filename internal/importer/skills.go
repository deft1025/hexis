package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"hexis/internal/staticdata"
)

func validateSkills(patch string, record map[string]any) (staticdata.SkillProfile, error) {
	var p staticdata.SkillProfile
	content, err := json.Marshal(record)
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(content, &p); err != nil {
		return p, err
	}
	version, err := staticdata.PatchVersion(p.Version)
	if err != nil || version != patch || p.ID == "" || p.Title == "" || len(p.Roles) == 0 || len(p.Skills) != 5 || hasStat(record) {
		return p, fmt.Errorf("skills require matching patch, champion, title, roles and five abilities without match statistics")
	}
	expected := "https://ddragon.leagueoflegends.com/cdn/" + p.Version
	if p.SourceURL != expected+"/data/zh_CN/champion/"+p.ID+".json" {
		return p, fmt.Errorf("skill source does not match champion version")
	}
	for index, skill := range p.Skills {
		folder := "spell"
		if index == 0 {
			folder = "passive"
		}
		if skill.Key != []string{"被动", "Q", "W", "E", "R"}[index] || skill.Name == "" || skill.Description == "" || !strings.HasPrefix(skill.IconURL, expected+"/img/"+folder+"/") || !strings.HasSuffix(skill.IconURL, ".png") {
			return p, fmt.Errorf("invalid skill %d", index)
		}
	}
	return p, nil
}

func importSkills(ctx context.Context, tx *sql.Tx, patchID, sourceID int64, record map[string]any) error {
	var patch string
	if err := tx.QueryRowContext(ctx, `SELECT version FROM patches WHERE id=?`, patchID).Scan(&patch); err != nil {
		return err
	}
	p, err := validateSkills(patch, record)
	if err != nil {
		return err
	}
	var championID int64
	var iconURL string
	if err := tx.QueryRowContext(ctx, `SELECT p.champion_id,p.icon_url FROM champion_patch_profiles p JOIN champions c ON c.id=p.champion_id WHERE p.patch_id=? AND c.riot_id=?`, patchID, p.ID).Scan(&championID, &iconURL); err != nil {
		return fmt.Errorf("skills champion %s missing in patch: %w", p.ID, err)
	}
	if iconURL != "https://ddragon.leagueoflegends.com/cdn/"+p.Version+"/img/champion/"+p.ID+".png" {
		return fmt.Errorf("skills source version differs from hero snapshot: %s", p.ID)
	}
	roles, _ := json.Marshal(p.Roles)
	skills, _ := json.Marshal(p.Skills)
	_, err = tx.ExecContext(ctx, `INSERT INTO champion_patch_skills (patch_id,champion_id,data_version,title,roles,skills,source_url,source_id)
		VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE data_version=VALUES(data_version),title=VALUES(title),roles=VALUES(roles),skills=VALUES(skills),source_url=VALUES(source_url),source_id=VALUES(source_id)`, patchID, championID, p.Version, p.Title, string(roles), string(skills), p.SourceURL, sourceID)
	return err
}
