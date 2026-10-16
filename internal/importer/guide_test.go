package importer

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGuideRequiresReviewedCompleteContent(t *testing.T) {
	valid := Guide{ID: "tank-control", Title: "控制与坦度", Category: "basic", Summary: "控制后承受伤害", Reviewed: true,
		ChampionIDs: []string{"TahmKench"}, Sections: []GuideSection{{Title: "联动", Body: "经过核对的效果说明"}},
		Stages:          []GuideStage{{Label: "核心", ItemIDs: []string{"3083"}, Reason: "生命值选择理由"}},
		Recommendations: []GuideRecommendation{{ID: "ARAM_CourageOfTheColossus", Priority: "core", Reason: "触发条件说明"}}}
	if err := validateGuide(valid); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Guide){
		func(g *Guide) { g.Reviewed = false }, func(g *Guide) { g.Category = "win" }, func(g *Guide) { g.ChampionIDs = nil },
		func(g *Guide) { g.Sections = nil }, func(g *Guide) { g.Stages = nil }, func(g *Guide) { g.ID = "../other" },
		func(g *Guide) {
			g.Recommendations = []GuideRecommendation{{ID: "x", Priority: "S", Reason: "unverified"}}
		},
		func(g *Guide) { g.PlaystyleName = "仅流派允许的名称" },
		func(g *Guide) {
			g.BasicSections = []GuideSection{{Title: "仅流派允许的思路", Body: "不应混入通用基础"}}
		},
	} {
		g := valid
		change(&g)
		if validateGuide(g) == nil {
			t.Fatalf("accepted invalid guide: %+v", g)
		}
	}
}

func TestPlaystyleContentAndPartialPreview(t *testing.T) {
	g := Guide{ID: "flow", Title: "流派", Category: "fun", Summary: "联动", ChampionIDs: []string{"TahmKench"},
		PlaystyleName: "恢复流", Sections: []GuideSection{{Title: "机制", Body: "公开机制提示"}},
		BasicSections: []GuideSection{{Title: "操作", Body: "本流派独立条件"}}}
	if err := validateGuideContent(g, true); err != nil {
		t.Fatal(err)
	}
	g.Reviewed = true
	if validateGuide(g) == nil {
		t.Fatal("published guide without equipment accepted")
	}
	g.Stages = []GuideStage{{Label: "核心", ItemIDs: []string{"3083"}, Reason: "已核对的选择理由"}}
	if err := validateGuide(g); err != nil {
		t.Fatal(err)
	}
	g.BasicSections = []GuideSection{{Title: "操作", Body: " "}}
	if validateGuide(g) == nil {
		t.Fatal("empty playstyle basic section accepted")
	}
}

func TestPreviewSamplesCannotBeImportedAsReviewedGuides(t *testing.T) {
	for _, name := range []string{"tahmkench", "garen", "katarina", "varus", "karma", "soraka"} {
		content, err := os.ReadFile("../../docs/guide-previews/" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var samples []Guide
		if err := json.Unmarshal(content, &samples); err != nil {
			t.Fatal(err)
		}
		want := 1
		if name == "tahmkench" {
			want = 4
		} else if name == "garen" {
			want = 2
		}
		if len(samples) != want {
			t.Fatalf("expected %d samples for %s", want, name)
		}
		for _, sample := range samples {
			if err := validateGuideContent(sample, true); err != nil {
				t.Fatalf("%s: %v", sample.ID, err)
			}
			if validateGuide(sample) == nil {
				t.Fatalf("unreviewed %s accepted as published content", sample.ID)
			}
			sample.Reviewed = true
			if validateGuideContent(sample, true) == nil {
				t.Fatal("reviewed guide accepted in preview table")
			}
		}
	}
}

func TestGuideReferencesRejectUnsafeLinks(t *testing.T) {
	g := Guide{ID: "source-link", Title: "来源", Category: "fun", Summary: "样例", ChampionIDs: []string{"Karma"},
		Sections: []GuideSection{{Title: "机制", Body: "公开提示"}}}
	for _, ref := range []GuideReference{
		{Name: "来源", URL: "javascript:alert(1)"},
		{Name: "来源", URL: "//example.com"},
		{Name: "来源", URL: "https://user:password@example.com"},
		{Name: "", URL: "https://example.com"},
	} {
		g.References = []GuideReference{ref}
		if validateGuideContent(g, true) == nil {
			t.Fatalf("accepted invalid reference: %+v", ref)
		}
	}
	g.References = []GuideReference{{Name: "补充来源", URL: "https://example.com/augment"}}
	if err := validateGuideContent(g, true); err != nil {
		t.Fatal(err)
	}
}
