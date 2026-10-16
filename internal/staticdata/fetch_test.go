package staticdata

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPatchVersion(t *testing.T) {
	for _, version := range []string{"latest", "16.20", "16.20.1/../../latest", ""} {
		if _, err := PatchVersion(version); err == nil {
			t.Fatalf("accepted unpinned version %q", version)
		}
	}
	if patch, err := PatchVersion("16.20.1"); err != nil || patch != "16.20" {
		t.Fatalf("PatchVersion = %q, %v", patch, err)
	}
}

func TestChampionCatalogVersionAndIDs(t *testing.T) {
	var zh, en championDocument
	if err := json.Unmarshal([]byte(`{"version":"16.20.1","data":{"Ahri":{"id":"Ahri","name":"九尾妖狐","image":{"full":"Ahri.png"}}}}`), &zh); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"version":"16.20.1","data":{"Ahri":{"id":"Ahri","name":"Ahri"}}}`), &en); err != nil {
		t.Fatal(err)
	}
	profiles, err := convertChampions("16.20.1", zh, en)
	if err != nil || len(profiles) != 1 || profiles[0].NameEn != "Ahri" {
		t.Fatalf("convertChampions = %v, %v", profiles, err)
	}
	if _, err := convertChampions("16.19.1", zh, en); err == nil {
		t.Fatal("accepted catalog from another version")
	}
	delete(en.Data, "Ahri")
	if _, err := convertChampions("16.20.1", zh, en); err == nil {
		t.Fatal("accepted mismatched catalogs")
	}
}

func TestAugmentsUseOnlyMayhemList(t *testing.T) {
	bin := map[string]binEntry{
		"shared": {Type: "AugmentData", AugmentNameID: "ARAM_Deft", NameTra: "Deft_Name", Description: "Deft_Desc", Icon: "ASSETS/Icons/Deft.tex"},
		"arena":  {Type: "AugmentData", AugmentNameID: "ArenaOnly"},
		"jade":   {Type: "AugmentData", AugmentNameID: "JadeOnly"},
	}
	translations := map[string]string{"deft_name": "灵巧", "deft_desc": "提升攻击速度。"}
	lists := []augmentList{{ModeName: "CHERRY", AugmentList: []string{"arena"}}, {ModeName: "KIWI", AugmentList: []string{"shared"}}, {ModeName: "KIWI_JADE", AugmentList: []string{"jade"}}}
	profiles, _, err := convertAugments("16.20", lists, bin, translations)
	if err != nil || len(profiles) != 1 || profiles[0].ID != "ARAM_Deft" || profiles[0].Rarity != "Silver" {
		t.Fatalf("convertAugments = %v, %v", profiles, err)
	}
	if profiles[0].IconURL != "https://raw.communitydragon.org/16.20/game/assets/icons/deft.png" {
		t.Fatal(profiles[0].IconURL)
	}
	content, _ := json.Marshal(profiles)
	for _, key := range []string{"sample_count", "win_rate", "pick_rate", "tier"} {
		if strings.Contains(string(content), `"`+key+`"`) {
			t.Fatalf("static download contains statistics: %s", content)
		}
	}
	for _, test := range []struct {
		name         string
		lists        []augmentList
		translations map[string]string
	}{
		{"missing translation", lists, map[string]string{"deft_name": "灵巧"}},
		{"missing mode", lists[:1], translations},
		{"missing augment", []augmentList{{ModeName: "KIWI", AugmentList: []string{"absent"}}}, translations},
		{"duplicate augment", []augmentList{{ModeName: "KIWI", AugmentList: []string{"shared", "shared"}}}, translations},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := convertAugments("16.20", test.lists, bin, test.translations); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestDecodeBinIgnoresLinkMetadata(t *testing.T) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(`{"__linked":["other.bin"],"Augment":{"__type":"AugmentData","AugmentNameId":"ARAM_Deft"}}`), &raw); err != nil {
		t.Fatal(err)
	}
	entries, err := decodeBin(raw)
	if err != nil || len(entries) != 1 || entries["Augment"].AugmentNameID != "ARAM_Deft" {
		t.Fatalf("decodeBin = %v, %v", entries, err)
	}
}

func TestRenderDescriptionPreservesKnownValues(t *testing.T) {
	text, tokens := renderDescription("获得%i:scaleAS%<speed>@AttackSpeed*100@%</speed>攻击速度。<br>{{ Item_Keyword_OnHit }}", []spellValue{{Name: "AttackSpeed", Values: []float64{0.6000000238, 0.6000000238}}}, map[string]string{"item_keyword_onhit": "<OnHit>攻击特效</OnHit>"})
	if text != "获得60%攻击速度。 攻击特效" || len(tokens) != 0 {
		t.Fatalf("renderDescription = %q, %v", text, tokens)
	}
}

func TestRenderDescriptionDoesNotInventRuntimeValues(t *testing.T) {
	text, tokens := renderDescription("获得@TotalShield@护盾、@Health@生命值，持续@Cooldown@秒。", []spellValue{{Name: "Health", Values: []float64{100, 200}}}, nil)
	if len(tokens) != 3 || strings.ContainsAny(text, "@<>{}") || strings.Contains(text, "100") || strings.Contains(text, "200") || !strings.Contains(text, "以游戏内为准") {
		t.Fatalf("renderDescription = %q, %v", text, tokens)
	}
}

func TestRenderDescriptionSelectsBasicAbilitySlot(t *testing.T) {
	text, tokens := renderDescription("【{{SpellName}}】获得@QAbilityHaste@技能急速。", []spellValue{{Name: "QAbilityHaste", Values: []float64{200, 200}}}, nil)
	if text != "【Q】获得200技能急速。" || len(tokens) != 0 {
		t.Fatalf("renderDescription = %q, %v", text, tokens)
	}
}
