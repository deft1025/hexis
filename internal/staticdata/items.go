package staticdata

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type ItemProfile struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	IconURL          string             `json:"icon_url"`
	Mode             string             `json:"mode"`
	Categories       []string           `json:"categories"`
	Price            int                `json:"price"`
	SellPrice        int                `json:"sell_price"`
	Purchasable      bool               `json:"purchasable"`
	Recipe           []string           `json:"recipe"`
	Stats            map[string]float64 `json:"stats"`
	RequiredChampion string             `json:"required_champion,omitempty"`
}

type itemDocument struct {
	Version string `json:"version"`
	Data    map[string]struct {
		Name             string             `json:"name"`
		Description      string             `json:"description"`
		Tags             []string           `json:"tags"`
		From             []string           `json:"from"`
		Stats            map[string]float64 `json:"stats"`
		RequiredChampion string             `json:"requiredChampion"`
		HideFromAll      bool               `json:"hideFromAll"`
		InStore          *bool              `json:"inStore"`
		Image            struct {
			Full string `json:"full"`
		} `json:"image"`
		Gold struct {
			Total       int  `json:"total"`
			Sell        int  `json:"sell"`
			Purchasable bool `json:"purchasable"`
		} `json:"gold"`
	} `json:"data"`
}

type mapEntry struct {
	Type      string   `json:"__type"`
	ModeName  string   `json:"mModeName"`
	ItemLists []string `json:"itemLists"`
	Items     []string `json:"mItems"`
}

func convertItems(version string, catalog itemDocument, mapData map[string]mapEntry) ([]ItemProfile, error) {
	if catalog.Version != version || len(catalog.Data) == 0 {
		return nil, fmt.Errorf("item catalog must be nonempty and match version %s", version)
	}
	config, exists := mapData["Maps/Shipping/Map12/Modes/KIWI"]
	if !exists || config.Type != "GameModeMapData" || !strings.EqualFold(config.ModeName, mode) || len(config.ItemLists) == 0 {
		return nil, fmt.Errorf("missing KIWI item configuration")
	}
	profiles := make([]ItemProfile, 0)
	seen := make(map[string]bool)
	for _, path := range config.ItemLists {
		list, exists := mapData[path]
		if !exists || list.Type != "GameModeItemList" || len(list.Items) == 0 {
			return nil, fmt.Errorf("missing KIWI item list %s", path)
		}
		for _, itemPath := range list.Items {
			id := strings.TrimPrefix(itemPath, "Items/")
			if _, err := strconv.ParseUint(id, 10, 32); err != nil || id == itemPath {
				return nil, fmt.Errorf("invalid item reference %s", itemPath)
			}
			if seen[id] {
				continue
			}
			seen[id] = true
			item, exists := catalog.Data[id]
			if !exists {
				return nil, fmt.Errorf("KIWI item %s missing from catalog", id)
			}
			// Runtime-only objects and entries explicitly hidden from the shop are
			// not part of the public catalog. Transformed items remain inspectable.
			if item.HideFromAll || (item.InStore != nil && !*item.InStore) {
				continue
			}
			if strings.TrimSpace(item.Name) == "" || item.Image.Full != id+".png" || item.Gold.Total < 0 || item.Gold.Sell < 0 {
				return nil, fmt.Errorf("invalid item profile %s", id)
			}
			description, _ := renderDescription(item.Description, nil, nil)
			if item.From == nil {
				item.From = []string{}
			}
			if item.Stats == nil {
				item.Stats = map[string]float64{}
			}
			profiles = append(profiles, ItemProfile{ID: id, Name: item.Name, Description: description,
				IconURL: "https://ddragon.leagueoflegends.com/cdn/" + version + "/img/item/" + item.Image.Full,
				Mode:    mode, Categories: itemCategories(item.Tags), Price: item.Gold.Total, SellPrice: item.Gold.Sell,
				Purchasable: item.Gold.Purchasable, Recipe: item.From, Stats: item.Stats, RequiredChampion: item.RequiredChampion})
		}
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("KIWI item catalog contains no visible items")
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	return profiles, nil
}

func itemCategories(tags []string) []string {
	groups := []struct {
		name string
		tags []string
	}{
		{"攻击", []string{"Damage", "AttackSpeed", "CriticalStrike", "ArmorPenetration", "LifeSteal", "OnHit"}},
		{"法术", []string{"SpellDamage", "MagicPenetration"}},
		{"防御", []string{"Health", "Armor", "SpellBlock", "HealthRegen"}},
		{"功能", []string{"Aura", "Active", "Slow", "CooldownReduction", "Mana", "ManaRegen", "NonbootsMovement", "Trinket", "Vision"}},
		{"鞋子", []string{"Boots"}},
		{"消耗品", []string{"Consumable"}},
	}
	categories := make([]string, 0)
	for _, group := range groups {
		found := false
		for _, tag := range tags {
			for _, candidate := range group.tags {
				found = found || tag == candidate
			}
		}
		if found {
			categories = append(categories, group.name)
		}
	}
	if len(categories) == 0 {
		categories = append(categories, "其他")
	}
	return categories
}
