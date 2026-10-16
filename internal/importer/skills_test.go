package importer

import (
	"encoding/json"
	"testing"

	"hexis/internal/model"
	"hexis/internal/staticdata"
)

func TestValidateSkillsVersionCoverageAndStaticData(t *testing.T) {
	p := staticdata.SkillProfile{ID: "Ahri", SkillsResult: model.SkillsResult{Version: "16.20.1", Title: "阿狸", Roles: []string{"法师"}, SourceURL: "https://ddragon.leagueoflegends.com/cdn/16.20.1/data/zh_CN/champion/Ahri.json"}}
	for index, key := range []string{"被动", "Q", "W", "E", "R"} {
		folder := "spell"
		if index == 0 {
			folder = "passive"
		}
		p.Skills = append(p.Skills, model.Skill{Key: key, Name: "技能", Description: "技能效果", IconURL: "https://ddragon.leagueoflegends.com/cdn/16.20.1/img/" + folder + "/Ahri.png"})
	}
	record := func() map[string]any {
		content, _ := json.Marshal(p)
		var r map[string]any
		json.Unmarshal(content, &r)
		return r
	}
	if _, err := validateSkills("16.20", record()); err != nil {
		t.Fatal(err)
	}
	if _, err := validateSkills("16.19", record()); err == nil {
		t.Fatal("accepted wrong patch")
	}
	for _, mutate := range []func(map[string]any){
		func(r map[string]any) { r["win_rate"] = 0.5 }, func(r map[string]any) { r["version"] = "latest" }, func(r map[string]any) { r["skills"] = []any{} },
		func(r map[string]any) { r["source_url"] = "https://example.com/Ahri.json" }, func(r map[string]any) { r["roles"] = []any{} },
		func(r map[string]any) { r["skills"].([]any)[1].(map[string]any)["key"] = "R" },
		func(r map[string]any) {
			r["skills"].([]any)[1].(map[string]any)["icon_url"] = "https://ddragon.leagueoflegends.com/cdn/16.19.1/img/spell/Ahri.png"
		},
	} {
		r := record()
		mutate(r)
		if _, err := validateSkills("16.20", r); err == nil {
			t.Fatalf("accepted invalid profile: %+v", r)
		}
	}
}
