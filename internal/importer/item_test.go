package importer

import (
	"encoding/json"
	"testing"
)

func TestValidateItem(t *testing.T) {
	const input = `{"id":"3031","name":"无尽之刃","mode":"KIWI","icon_url":"https://example.com/3031.png","categories":["攻击"],"price":3500,"sell_price":2450,"purchasable":true,"recipe":["1038"],"stats":{"FlatPhysicalDamageMod":75}}`
	for _, test := range []struct {
		name string
		edit func(map[string]any)
		bad  bool
	}{
		{"valid", func(map[string]any) {}, false},
		{"wrong mode", func(r map[string]any) { r["mode"] = "CHERRY" }, true},
		{"negative price", func(r map[string]any) { r["price"] = float64(-1) }, true},
		{"fractional price", func(r map[string]any) { r["price"] = 2.5 }, true},
		{"zero id", func(r map[string]any) { r["id"] = "0" }, true},
		{"numeric id", func(r map[string]any) { r["id"] = float64(3031) }, true},
		{"missing categories", func(r map[string]any) { delete(r, "categories") }, true},
		{"bad recipe", func(r map[string]any) { r["recipe"] = []any{"bad"} }, true},
		{"bad attributes", func(r map[string]any) { r["stats"] = map[string]any{"damage": "75"} }, true},
		{"fabricated match stats", func(r map[string]any) { r["win_rate"] = 0.5 }, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var record map[string]any
			if err := json.Unmarshal([]byte(input), &record); err != nil {
				t.Fatal(err)
			}
			test.edit(record)
			if err := validateItem(record); (err != nil) != test.bad {
				t.Fatalf("validateItem = %v, want error %v", err, test.bad)
			}
		})
	}
}
