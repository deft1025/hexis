package staticdata

import (
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"strings"
)

type spellCalculation struct {
	Type             string            `json:"__type"`
	Parts            []calculationPart `json:"mFormulaParts"`
	Percent          bool              `json:"mDisplayAsPercent"`
	Multiplier       *calculationPart  `json:"mMultiplier"`
	RangedMultiplier *calculationPart  `json:"mRangedMultiplier"`
}

type calculationPart struct {
	Type        string  `json:"__type"`
	Number      float64 `json:"mNumber"`
	DataValue   string  `json:"mDataValue"`
	Stat        int     `json:"mStat"`
	StatFormula int     `json:"mStatFormula"`
	Coefficient float64 `json:"mCoefficient"`
	Start       float64 `json:"mStartValue"`
	End         float64 `json:"mEndValue"`
	Level1      float64 `json:"mLevel1Value"`
	Breakpoints []struct {
		Level    int     `json:"mLevel"`
		Bonus    float64 `json:"mAdditionalBonusAtThisLevel"`
		PerLevel float64 `json:"mBonusPerLevelAtAndAfter"`
	} `json:"mBreakpoints"`
}

type formulaTerm struct {
	Number float64
	Stat   string
	End    *float64
	Note   string
}

type formula []formulaTerm

func binKey(name string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(name)))
	return fmt.Sprintf("{%08x}", h.Sum32())
}

func lookupCalculation(calcs map[string]spellCalculation, name string) (spellCalculation, bool) {
	for key, calc := range calcs {
		if strings.EqualFold(key, name) || key == binKey(name) {
			return calc, true
		}
	}
	return spellCalculation{}, false
}

func constantValue(values []spellValue, name string) (float64, bool) {
	for _, value := range values {
		if !strings.EqualFold(value.Name, name) || len(value.Values) == 0 {
			continue
		}
		n := value.Values[0]
		for _, v := range value.Values {
			if v != n {
				return 0, false
			}
		}
		return n, true
	}
	return 0, false
}

func statName(stat, kind int) (string, bool) {
	name, ok := map[int]string{0: "法术强度", 1: "护甲", 2: "攻击力", 6: "魔法抗性", 12: "最大生命值"}[stat]
	if !ok || (kind != 0 && kind != 2) {
		return "", false
	}
	if kind == 2 {
		if stat == 12 {
			name = "额外生命值"
		} else {
			name = "额外" + name
		}
	}
	return name, true
}

func resolvePart(part calculationPart, values []spellValue) (formula, bool) {
	switch part.Type {
	case "NumberCalculationPart":
		return formula{{Number: part.Number}}, true
	case "NamedDataValueCalculationPart":
		n, ok := constantValue(values, part.DataValue)
		return formula{{Number: n}}, ok
	case "StatByNamedDataValueCalculationPart", "StatByCoefficientCalculationPart":
		name, ok := statName(part.Stat, part.StatFormula)
		if !ok {
			return nil, false
		}
		n := part.Coefficient
		if part.Type == "StatByNamedDataValueCalculationPart" {
			n, ok = constantValue(values, part.DataValue)
		}
		return formula{{Number: n, Stat: name}}, ok
	case "ByCharLevelInterpolationCalculationPart":
		end := part.End
		return formula{{Number: part.Start, End: &end, Note: "随英雄等级成长"}}, true
	case "ByCharLevelBreakpointsCalculationPart":
		value := part.Level1
		notes := []string{"1级" + number(value)}
		for _, bp := range part.Breakpoints {
			if bp.PerLevel != 0 {
				return nil, false
			}
			value += bp.Bonus
			notes = append(notes, strconv.Itoa(bp.Level)+"级"+number(value))
		}
		return formula{{Note: strings.Join(notes, "、")}}, true
	}
	return nil, false
}

func (f formula) scale(n float64) formula {
	result := make(formula, len(f))
	for i, term := range f {
		result[i] = term
		result[i].Number *= n
		if term.End != nil {
			end := *term.End * n
			result[i].End = &end
		}
	}
	return result
}

func (f formula) scalar() (float64, bool) {
	n := 0.0
	for _, t := range f {
		if t.Stat != "" || t.End != nil || t.Note != "" {
			return 0, false
		}
		n += t.Number
	}
	return n, len(f) > 0
}

func number(n float64) string {
	return strconv.FormatFloat(math.Round(n*1e4)/1e4, 'f', -1, 64)
}

func (f formula) text(scale float64) string {
	parts := make([]string, 0, len(f))
	for _, t := range f.scale(scale) {
		if t.Note != "" && t.End == nil {
			parts = append(parts, "（"+t.Note+"）")
			continue
		}
		text := number(t.Number)
		if t.End != nil {
			text += "~" + number(*t.End) + "（" + t.Note + "）"
		}
		if t.Stat != "" {
			text += "×" + t.Stat
		}
		parts = append(parts, text)
	}
	text := strings.Join(parts, " + ")
	if len(parts) > 1 {
		text = "（" + text + "）"
	}
	return text
}

func resolveCalculation(calc spellCalculation, values []spellValue) (formula, bool) {
	var result formula
	if calc.Type != "GameCalculation" && calc.Type != "{e9a3c91d}" {
		return nil, false
	}
	for _, part := range calc.Parts {
		terms, ok := resolvePart(part, values)
		if !ok {
			return nil, false
		}
		result = append(result, terms...)
	}
	if len(result) == 0 {
		return nil, false
	}
	if calc.Multiplier != nil {
		part, ok := resolvePart(*calc.Multiplier, values)
		n, scalar := part.scalar()
		if !ok || !scalar {
			return nil, false
		}
		for _, term := range result {
			if term.Note != "" && term.End == nil && n != 1 {
				return nil, false
			}
		}
		result = result.scale(n)
	}
	return result, true
}

func calculationText(name string, calc spellCalculation, values []spellValue, multiplier float64, suffix string) (string, bool) {
	f, ok := resolveCalculation(calc, values)
	if !ok {
		return "", false
	}
	scale := multiplier
	if calc.Percent {
		if multiplier == 1 {
			scale = 100
		}
		suffix = "%"
	}
	if calc.RangedMultiplier != nil {
		part, ok := resolvePart(*calc.RangedMultiplier, values)
		ratio, scalar := part.scalar()
		base, baseScalar := f.scalar()
		if !ok || !scalar || !baseScalar {
			return "", false
		}
		ranged := base * ratio
		// Prefer explicit melee/ranged values over rounded tooltip multipliers.
		if strings.EqualFold(name, "RangeIncrease") {
			if n, found := constantValue(values, "MeleeRangeIncrease"); found {
				base = n
			}
			if n, found := constantValue(values, "RangedRangeIncrease"); found {
				ranged = n
			}
		}
		return "（近战" + number(math.Round(base*scale)) + suffix + "；远程" + number(math.Round(ranged*scale)) + suffix + "）", true
	}
	return f.text(scale) + suffix, true
}

func renderAugmentDescription(template string, augment binEntry, bin map[string]binEntry, translations map[string]string) (string, []string) {
	spell := bin[augment.RootSpell].Spell
	knownSlot := false
	for _, value := range spell.DataValues {
		if value.Name == "QAbilityHaste" || value.Name == "WAbilityHaste" || value.Name == "EAbilityHaste" {
			knownSlot = true
		}
	}
	if augment.IsAbilityAugment && !knownSlot {
		// The target is selected during a match, so there is no universal Q/W/E name.
		template = strings.ReplaceAll(template, "【{{SpellName}}】", "所选")
		template = strings.ReplaceAll(template, "{{SpellName}}", "所选")
		template = strings.ReplaceAll(template, "你的所选", "你选中的")
	}
	quest := bin[augment.QuestData.Quest]
	firstQuest := !augment.IsAbilityAugment && quest.Type == "ModesQuestData" && len(quest.Milestones) > 0
	usedQuest := false
	template = valuePattern.ReplaceAllStringFunc(template, func(token string) string {
		match := valuePattern.FindStringSubmatch(token)
		parts := strings.Split(match[1], "*")
		multiplier := 1.0
		if len(parts) > 2 {
			return token
		}
		if len(parts) == 2 {
			var err error
			multiplier, err = strconv.ParseFloat(parts[1], 64)
			if err != nil || math.IsInf(multiplier, 0) || math.IsNaN(multiplier) {
				return token
			}
		}
		name := parts[0]
		if strings.EqualFold(name, "QuestRequirement") && firstQuest {
			usedQuest = true
			return number(quest.Milestones[0].Requirement*multiplier) + match[2]
		}
		if calc, found := lookupCalculation(spell.Calculations, name); found {
			if text, ok := calculationText(name, calc, spell.DataValues, multiplier, match[2]); ok {
				return text
			}
		}
		return token
	})
	if usedQuest && len(quest.Milestones) > 1 {
		template += "<br>后续任务目标会随任务等级提高。"
	}
	return renderDescription(template, spell.DataValues, translations)
}
