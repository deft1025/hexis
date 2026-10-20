package staticdata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAugmentCalculationsAndRuntimeTargets(t *testing.T) {
	const input = `{
"root":{"mSpell":{"DataValues":[
{"name":"HealthScalar","values":[0.03999999910593033]},
{"name":"CritChance","values":[0.25]},
{"name":"APRatio","values":[0.00044999999227002263]},
{"name":"MeleeRangeIncrease","values":[75]},
{"name":"RangedRangeIncrease","values":[50]},
{"name":"Reward","values":[80]}],
"mSpellCalculations":{
"TotalShield":{"__type":"GameCalculation","mFormulaParts":[{"__type":"ByCharLevelInterpolationCalculationPart","mStartValue":150,"mEndValue":450},{"__type":"StatByNamedDataValueCalculationPart","mStat":12,"mDataValue":"HealthScalar"}]},
"CritGranted":{"__type":"GameCalculation","mDisplayAsPercent":true,"mFormulaParts":[{"__type":"NamedDataValueCalculationPart","mDataValue":"CritChance"},{"__type":"StatByNamedDataValueCalculationPart","mDataValue":"APRatio"}]},
"{90a024ae}":{"__type":"GameCalculation","mDisplayAsPercent":true,"mFormulaParts":[{"__type":"NumberCalculationPart","mNumber":0.5}]},
"RangeIncrease":{"__type":"{e9a3c91d}","mFormulaParts":[{"__type":"NumberCalculationPart","mNumber":75}],"mRangedMultiplier":{"__type":"NumberCalculationPart","mNumber":0.6669999957084656}},
"RewardMax":{"__type":"GameCalculation","mFormulaParts":[{"__type":"NumberCalculationPart","mNumber":2}],"mMultiplier":{"__type":"NamedDataValueCalculationPart","mDataValue":"Reward"}},
"Breaks":{"__type":"GameCalculation","mFormulaParts":[{"__type":"ByCharLevelBreakpointsCalculationPart","mLevel1Value":1,"mBreakpoints":[{"mLevel":7,"mAdditionalBonusAtThisLevel":1},{"mLevel":11,"mAdditionalBonusAtThisLevel":1}]}]},
"Unknown":{"__type":"GameCalculation","mFormulaParts":[{"__type":"StatByCoefficientCalculationPart","mStat":99,"mCoefficient":0.2}]}
}}},
"quest":{"__type":"ModesQuestData","Milestones":[{"{7fec0982}":18}]},
"q":{"mSpell":{"DataValues":[{"name":"QAbilityHaste","values":[100]}]}}
}`
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		t.Fatal(err)
	}
	bin, err := decodeBin(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, template, expected string
		augment                  binEntry
		unresolved               int
	}{
		{"health formula", "@TotalShield@护盾", "（150~450（随英雄等级成长） + 0.04×最大生命值）护盾", binEntry{RootSpell: "root"}, 0},
		{"AP crit conversion", "@CritGranted@暴击几率", "（25 + 0.045×法术强度）%暴击几率", binEntry{RootSpell: "root"}, 0},
		{"hashed percent", "@SpinDamageAmp_Ult@", "50%", binEntry{RootSpell: "root"}, 0},
		{"exact ranged values", "@RangeIncrease@", "（近战75；远程50）", binEntry{RootSpell: "root"}, 0},
		{"multiplier", "@RewardMax@", "160", binEntry{RootSpell: "root"}, 0},
		{"breakpoints", "@Breaks@", "（1级1、7级2、11级3）", binEntry{RootSpell: "root"}, 0},
		{"runtime skill", "你的【{{SpellName}}】技能", "你选中的技能", binEntry{RootSpell: "root", IsAbilityAugment: true}, 0},
		{"known slot", "你的【{{SpellName}}】技能获得@QAbilityHaste@技能急速", "你的【Q】技能获得100技能急速", binEntry{RootSpell: "q", IsAbilityAugment: true}, 0},
		{"unknown stat", "@Unknown@", "一定数额的 具体数值及适用技能以游戏内为准。", binEntry{RootSpell: "root"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text, unresolved := renderAugmentDescription(tc.template, tc.augment, bin, nil)
			if text != tc.expected || len(unresolved) != tc.unresolved {
				t.Fatalf("got %q, %v", text, unresolved)
			}
		})
	}
	quest := binEntry{RootSpell: "root"}
	quest.QuestData.Quest = "quest"
	text, unresolved := renderAugmentDescription("参与击杀@QuestRequirement@次", quest, bin, nil)
	if text != "参与击杀18次" || len(unresolved) != 0 {
		t.Fatalf("quest: %q %v", text, unresolved)
	}
	quest.IsAbilityAugment = true
	_, unresolved = renderAugmentDescription("@QuestRequirement@", quest, bin, nil)
	if len(unresolved) != 1 {
		t.Fatal("runtime skill quest was incorrectly replaced with a fixed threshold")
	}
	if binKey("SpinDamageAmp_Ult") != "{90a024ae}" {
		t.Fatal("incorrect bin hash")
	}
	if number(0.30000001192092896*100) != "30" {
		t.Fatal("float32 noise leaked into text")
	}
}

func TestRebuildAugmentsVerifiesOriginalSources(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "raw"), 0700); err != nil {
		t.Fatal(err)
	}
	inputs := map[string]string{
		"augment-lists.json": `[{"modeName":"KIWI","augmentList":["a"]}]`,
		"kiwi.bin.json":      `{"a":{"__type":"AugmentData","AugmentNameId":"A","NameTra":"name","DescriptionTra":"desc","AugmentSmallIconPath":"assets/a.tex","RootSpell":"s"},"s":{"mSpell":{"mSpellCalculations":{"Amount":{"__type":"GameCalculation","mFormulaParts":[{"__type":"NumberCalculationPart","mNumber":25}]}}}}}`,
		"strings-zh.json":    `{"entries":{"name":"测试","desc":"@Amount@"}}`,
	}
	m := Manifest{Version: "16.20.1", Patch: "16.20", Mode: "KIWI", Locale: "zh_CN", Augments: 1}
	for name, content := range inputs {
		if err := os.WriteFile(filepath.Join(dir, "raw", name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256([]byte(content))
		m.Sources = append(m.Sources, SourceFile{File: "raw/" + name, URL: "https://raw.communitydragon.org/16.20/" + name, SHA256: hex.EncodeToString(h[:])})
	}
	content, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "augments.json"), []byte(`[{"id":"A","name":"测试"}]`), 0600); err != nil {
		t.Fatal(err)
	}
	_, _, profiles, unresolved, err := RebuildAugments(dir)
	if err != nil || len(profiles) != 1 || profiles[0].Description != "25" || len(unresolved) != 0 {
		t.Fatalf("rebuild: %v %v", profiles, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "raw/kiwi.bin.json"), []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, _, _, err := RebuildAugments(dir); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("accepted changed snapshot: %v", err)
	}
}
