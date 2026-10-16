package staticdata

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const mode = "KIWI"

type Profile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NameEn      string `json:"name_en,omitempty"`
	IconURL     string `json:"icon_url"`
	Description string `json:"description,omitempty"`
	Rarity      string `json:"rarity,omitempty"`
}

type SourceFile struct {
	URL    string `json:"url"`
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

type VariableDescription struct {
	ID       string   `json:"id"`
	Template string   `json:"template"`
	Tokens   []string `json:"tokens"`
}

type Manifest struct {
	Version              string                `json:"version"`
	Patch                string                `json:"patch"`
	Mode                 string                `json:"mode"`
	Locale               string                `json:"locale"`
	FetchedAt            time.Time             `json:"fetched_at"`
	Champions            int                   `json:"champions"`
	Augments             int                   `json:"augments"`
	Items                int                   `json:"items"`
	SkillChampions       int                   `json:"skill_champions"`
	StatisticsScope      string                `json:"statistics_scope"`
	Sources              []SourceFile          `json:"sources"`
	VariableDescriptions []VariableDescription `json:"variable_descriptions"`
}

var versionPattern = regexp.MustCompile(`^([0-9]+\.[0-9]+)\.[0-9]+$`)

func PatchVersion(version string) (string, error) {
	match := versionPattern.FindStringSubmatch(version)
	if match == nil {
		return "", fmt.Errorf("--version must be a pinned Data Dragon version such as 16.20.1")
	}
	return match[1], nil
}

func Fetch(ctx context.Context, version, out string) (*Manifest, error) {
	patch, err := PatchVersion(version)
	if err != nil {
		return nil, err
	}
	manifest := &Manifest{
		Version: version, Patch: patch, Mode: mode, Locale: "zh_CN", FetchedAt: time.Now().UTC(),
		StatisticsScope: "CN ARAM Mayhem only; this download contains no match statistics",
	}
	client := &http.Client{Timeout: 90 * time.Second}
	ddragon := "https://ddragon.leagueoflegends.com/cdn/" + version
	cdragon := "https://raw.communitydragon.org/" + patch
	if err := os.MkdirAll(filepath.Join(out, "raw"), 0755); err != nil {
		return nil, err
	}
	download := func(name, url string, target any) error {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		response, err := client.Do(request)
		if err != nil {
			return fmt.Errorf("download %s: %w", url, err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("download %s: HTTP %d", url, response.StatusCode)
		}
		const maxSize = 64 << 20
		content, err := io.ReadAll(io.LimitReader(response.Body, maxSize+1))
		if err != nil {
			return err
		}
		if len(content) > maxSize {
			return fmt.Errorf("download %s exceeds 64 MiB", url)
		}
		if err := json.Unmarshal(content, target); err != nil {
			return fmt.Errorf("decode %s: %w", url, err)
		}
		file := filepath.Join("raw", name)
		if err := os.WriteFile(filepath.Join(out, file), content, 0644); err != nil {
			return err
		}
		hash := sha256.Sum256(content)
		manifest.Sources = append(manifest.Sources, SourceFile{URL: url, File: file, SHA256: hex.EncodeToString(hash[:])})
		return nil
	}
	var zh, en championDocument
	if err := download("champions-zh.json", ddragon+"/data/zh_CN/champion.json", &zh); err != nil {
		return nil, err
	}
	if err := download("champions-en.json", ddragon+"/data/en_US/champion.json", &en); err != nil {
		return nil, err
	}
	champions, err := convertChampions(version, zh, en)
	if err != nil {
		return nil, err
	}
	var lists []augmentList
	var rawBin map[string]json.RawMessage
	var translations struct {
		Entries map[string]string `json:"entries"`
	}
	if err := download("augment-lists.json", cdragon+"/plugins/rcp-be-lol-game-data/global/zh_cn/v1/augment-lists.json", &lists); err != nil {
		return nil, err
	}
	if err := download("kiwi.bin.json", cdragon+"/game/maps/modespecificdata/kiwi.bin.json", &rawBin); err != nil {
		return nil, err
	}
	bin, err := decodeBin(rawBin)
	if err != nil {
		return nil, err
	}
	// Localized string tables are stored under en_us even for the zh_cn bundle.
	if err := download("strings-zh.json", cdragon+"/game/zh_cn/data/menu/en_us/lol.stringtable.json", &translations); err != nil {
		return nil, err
	}
	augments, variable, err := convertAugments(patch, lists, bin, translations.Entries)
	if err != nil {
		return nil, err
	}
	manifest.Champions, manifest.Augments = len(champions), len(augments)
	manifest.VariableDescriptions = variable
	var itemCatalog itemDocument
	var mapData map[string]mapEntry
	if err := download("items-zh.json", ddragon+"/data/zh_CN/item.json", &itemCatalog); err != nil {
		return nil, err
	}
	var rawMap map[string]json.RawMessage
	if err := download("map12.bin.json", cdragon+"/game/data/maps/shipping/map12/map12.bin.json", &rawMap); err != nil {
		return nil, err
	}
	mapData = make(map[string]mapEntry, len(rawMap))
	for path, content := range rawMap {
		if strings.HasPrefix(path, "__") {
			continue
		}
		var entry mapEntry
		if err := json.Unmarshal(content, &entry); err != nil {
			return nil, fmt.Errorf("decode map entry %s: %w", path, err)
		}
		mapData[path] = entry
	}
	items, err := convertItems(version, itemCatalog, mapData)
	if err != nil {
		return nil, err
	}
	manifest.Items = len(items)
	ids := make([]string, 0, len(champions))
	for _, p := range champions {
		ids = append(ids, p.ID)
	}
	skills, skillSources, err := FetchSkills(ctx, client, version, out, ids)
	if err != nil {
		return nil, err
	}
	manifest.SkillChampions = len(skills)
	manifest.Sources = append(manifest.Sources, skillSources...)
	for _, file := range []struct {
		name string
		data any
	}{{"champions.json", champions}, {"augments.json", augments}, {"items.json", items}, {"manifest.json", manifest}} {
		content, err := json.MarshalIndent(file.data, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(out, file.name), append(content, '\n'), 0644); err != nil {
			return nil, err
		}
	}
	return manifest, nil
}

type championDocument struct {
	Version string `json:"version"`
	Data    map[string]struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Image struct {
			Full string `json:"full"`
		} `json:"image"`
	} `json:"data"`
}

func convertChampions(version string, zh, en championDocument) ([]Profile, error) {
	if zh.Version != version || en.Version != version || len(zh.Data) == 0 || len(zh.Data) != len(en.Data) {
		return nil, fmt.Errorf("champion catalogs must be nonempty and match version %s and each other", version)
	}
	profiles := make([]Profile, 0, len(zh.Data))
	for id, champion := range zh.Data {
		english, exists := en.Data[id]
		if !exists || id != champion.ID || english.ID != id || champion.Name == "" || english.Name == "" || champion.Image.Full != id+".png" {
			return nil, fmt.Errorf("invalid champion %q", id)
		}
		profiles = append(profiles, Profile{ID: id, Name: champion.Name, NameEn: english.Name,
			IconURL: "https://ddragon.leagueoflegends.com/cdn/" + version + "/img/champion/" + champion.Image.Full})
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	return profiles, nil
}

type augmentList struct {
	ModeName    string   `json:"modeName"`
	AugmentList []string `json:"augmentList"`
}

type spellValue struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
}

type binEntry struct {
	Type          string `json:"__type"`
	AugmentNameID string `json:"AugmentNameId"`
	NameTra       string `json:"NameTra"`
	Description   string `json:"DescriptionTra"`
	RootSpell     string `json:"RootSpell"`
	Icon          string `json:"AugmentSmallIconPath"`
	Rarity        int    `json:"rarity"`
	Spell         struct {
		DataValues []spellValue `json:"DataValues"`
	} `json:"mSpell"`
}

func decodeBin(raw map[string]json.RawMessage) (map[string]binEntry, error) {
	entries := make(map[string]binEntry, len(raw))
	for path, content := range raw {
		if strings.HasPrefix(path, "__") {
			continue
		}
		var entry binEntry
		if err := json.Unmarshal(content, &entry); err != nil {
			return nil, fmt.Errorf("decode bin entry %s: %w", path, err)
		}
		entries[path] = entry
	}
	return entries, nil
}

func convertAugments(patch string, lists []augmentList, bin map[string]binEntry, translations map[string]string) ([]Profile, []VariableDescription, error) {
	var paths []string
	for _, list := range lists {
		if list.ModeName == mode {
			if paths != nil {
				return nil, nil, fmt.Errorf("duplicate %s augment lists", mode)
			}
			paths = list.AugmentList
		}
	}
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("missing %s augment list", mode)
	}
	profiles := make([]Profile, 0, len(paths))
	variable := make([]VariableDescription, 0)
	seen := make(map[string]bool)
	for _, path := range paths {
		augment, exists := bin[path]
		if !exists || augment.Type != "AugmentData" || augment.AugmentNameID == "" || seen[augment.AugmentNameID] {
			return nil, nil, fmt.Errorf("missing, invalid or duplicate augment %q", path)
		}
		seen[augment.AugmentNameID] = true
		name := translations[strings.ToLower(augment.NameTra)]
		template := translations[strings.ToLower(augment.Description)]
		if strings.TrimSpace(name) == "" || strings.TrimSpace(template) == "" {
			return nil, nil, fmt.Errorf("missing Chinese translation for %s", augment.AugmentNameID)
		}
		if augment.Rarity < 0 || augment.Rarity > 2 {
			return nil, nil, fmt.Errorf("unknown rarity for %s", augment.AugmentNameID)
		}
		icon := strings.ToLower(augment.Icon)
		if !strings.HasPrefix(icon, "assets/") || strings.Contains(icon, "..") || !strings.HasSuffix(icon, ".tex") {
			return nil, nil, fmt.Errorf("invalid icon for %s", augment.AugmentNameID)
		}
		description, tokens := renderDescription(template, bin[augment.RootSpell].Spell.DataValues, translations)
		if len(tokens) > 0 {
			variable = append(variable, VariableDescription{ID: augment.AugmentNameID, Template: template, Tokens: tokens})
		}
		profiles = append(profiles, Profile{ID: augment.AugmentNameID, Name: name, Description: description,
			Rarity:  []string{"Silver", "Gold", "Prismatic"}[augment.Rarity],
			IconURL: "https://raw.communitydragon.org/" + patch + "/game/" + strings.TrimSuffix(icon, ".tex") + ".png"})
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	return profiles, variable, nil
}

var (
	templatePattern = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)
	valuePattern    = regexp.MustCompile(`@([^@]+)@(%)?`)
	iconPattern     = regexp.MustCompile(`%i:[^%]+%`)
)

func PlainDescription(template string) string {
	text, _ := renderDescription(template, nil, nil)
	return text
}

func renderDescription(template string, values []spellValue, translations map[string]string) (string, []string) {
	unresolved := make(map[string]bool)
	text := template
	for depth := 0; depth < 5 && templatePattern.MatchString(text); depth++ {
		text = templatePattern.ReplaceAllStringFunc(text, func(token string) string {
			key := strings.TrimSpace(templatePattern.FindStringSubmatch(token)[1])
			if strings.EqualFold(key, "SpellName") {
				for _, value := range values {
					switch value.Name {
					case "QAbilityHaste":
						return "Q"
					case "WAbilityHaste":
						return "W"
					case "EAbilityHaste":
						return "E"
					}
				}
				unresolved[key] = true
				return "对应"
			}
			if translated := translations[strings.ToLower(key)]; translated != "" {
				return translated
			}
			unresolved[key] = true
			return "对应效果"
		})
	}
	// Only substitute constants. Level-dependent values and runtime calculations
	// remain qualitative rather than being evaluated with invented player stats.
	text = valuePattern.ReplaceAllStringFunc(text, func(token string) string {
		match := valuePattern.FindStringSubmatch(token)
		parts := strings.Split(match[1], "*")
		multiplier := 1.0
		valid := len(parts) == 1
		if len(parts) == 2 {
			var err error
			multiplier, err = strconv.ParseFloat(parts[1], 64)
			valid = err == nil && !math.IsInf(multiplier, 0) && !math.IsNaN(multiplier)
		}
		if valid {
			for _, value := range values {
				if !strings.EqualFold(value.Name, parts[0]) || len(value.Values) == 0 {
					continue
				}
				constant := true
				for _, number := range value.Values {
					constant = constant && number == value.Values[0]
				}
				if constant {
					number := math.Round(value.Values[0]*multiplier*10000) / 10000
					return strconv.FormatFloat(number, 'f', -1, 64) + match[2]
				}
			}
		}
		unresolved[match[1]] = true
		if match[2] != "" {
			return "一定比例的"
		}
		return "一定数额的"
	})
	text = iconPattern.ReplaceAllString(text, "")
	text = templatePattern.ReplaceAllString(text, "对应效果")
	tokenizer := html.NewTokenizer(strings.NewReader(text))
	var plain strings.Builder
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			tokens := make([]string, 0, len(unresolved))
			for token := range unresolved {
				tokens = append(tokens, token)
			}
			sort.Strings(tokens)
			result := strings.Join(strings.Fields(plain.String()), " ")
			for _, unit := range []string{"秒", "次", "只", "个"} {
				result = strings.ReplaceAll(result, "一定数额的"+unit, "若干"+unit)
			}
			if len(tokens) > 0 {
				result += " 具体数值及适用技能以游戏内为准。"
			}
			return result, tokens
		case html.TextToken:
			plain.Write(tokenizer.Text())
		case html.StartTagToken, html.SelfClosingTagToken:
			tag, _ := tokenizer.TagName()
			if string(tag) == "br" || string(tag) == "p" {
				plain.WriteByte(' ')
			}
		}
	}
}
