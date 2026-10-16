package staticdata

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestItemsSelectModeListsAndRetainTransformations(t *testing.T) {
	var catalog itemDocument
	if err := json.Unmarshal([]byte(`{"version":"16.20.1","data":{
		"3031":{"name":"无尽之刃","description":"<stats>75攻击力<br>25%暴击几率</stats>","image":{"full":"3031.png"},"tags":["Damage","CriticalStrike"],"gold":{"total":3500,"sell":2450,"purchasable":true},"from":["1038"]},
		"3040":{"name":"炽天使之拥","image":{"full":"3040.png"},"gold":{"total":2900,"sell":2030,"purchasable":false}},
		"1500":{"name":"内部物品","image":{"full":"1500.png"},"inStore":false},
		"7050":{"name":"隐藏物品","image":{"full":"7050.png"},"hideFromAll":true},
		"771001":{"name":"其他模式装备","image":{"full":"771001.png"}}
	}}`), &catalog); err != nil {
		t.Fatal(err)
	}
	mapData := map[string]mapEntry{
		"Maps/Shipping/Map12/Modes/KIWI":      {Type: "GameModeMapData", ModeName: "Kiwi", ItemLists: []string{"base", "extra"}},
		"Maps/Shipping/Map12/Modes/KIWI_JADE": {Type: "GameModeMapData", ItemLists: []string{"jade"}},
		"base":                                {Type: "GameModeItemList", Items: []string{"Items/3031", "Items/3040", "Items/1500", "Items/7050"}},
		"extra":                               {Type: "GameModeItemList", Items: []string{"Items/3031"}},
		"jade":                                {Type: "GameModeItemList", Items: []string{"Items/771001"}},
	}
	items, err := convertItems("16.20.1", catalog, mapData)
	if err != nil || len(items) != 2 {
		t.Fatalf("convertItems = %v, %v", items, err)
	}
	if items[0].Price != 3500 || items[0].Description != "75攻击力 25%暴击几率" || len(items[0].Recipe) != 1 || items[0].Mode != "KIWI" {
		t.Fatalf("incorrect item profile: %+v", items[0])
	}
	if items[1].Purchasable || items[1].Recipe == nil || items[1].Stats == nil {
		t.Fatalf("lost transformed item or empty structured fields: %+v", items[1])
	}
	if _, err := convertItems("16.21.1", catalog, mapData); err == nil {
		t.Fatal("accepted mismatched version")
	}
	delete(mapData, "extra")
	if _, err := convertItems("16.20.1", catalog, mapData); err == nil {
		t.Fatal("accepted missing mode list")
	}
	mapData["extra"] = mapEntry{Type: "GameModeItemList", Items: []string{"Items/99999"}}
	if _, err := convertItems("16.20.1", catalog, mapData); err == nil {
		t.Fatal("accepted missing item")
	}
}

func TestItemCategoriesAllowMultipleRoles(t *testing.T) {
	categories := itemCategories([]string{"SpellDamage", "Health", "Mana", "CooldownReduction"})
	if strings.Join(categories, ",") != "法术,防御,功能" {
		t.Fatal(categories)
	}
}
