package importer

import "testing"

func TestValidateStats(t *testing.T) {
	tests := []struct {
		name    string
		record  map[string]any
		wantErr bool
	}{
		{"static profile", map[string]any{"tier": "gold"}, false},
		{"valid statistics", map[string]any{"sample_count": float64(20), "pick_rate": 0.25}, false},
		{"negative samples", map[string]any{"sample_count": float64(-1)}, true},
		{"fractional samples", map[string]any{"sample_count": 2.5}, true},
		{"missing sample size", map[string]any{"win_rate": 0.5}, true},
		{"invalid rate", map[string]any{"sample_count": float64(20), "pick_rate": 1.5}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateStats(test.record); (err != nil) != test.wantErr {
				t.Fatalf("validateStats() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestRecords(t *testing.T) {
	for _, input := range []any{
		[]any{map[string]any{"id": "A"}},
		map[string]any{"data": []any{map[string]any{"id": "A"}}},
	} {
		got, err := records(input)
		if err != nil || len(got) != 1 {
			t.Fatalf("records(%v) = %v, %v", input, got, err)
		}
	}
	if _, err := records([]any{"invalid"}); err == nil {
		t.Fatal("expected invalid record error")
	}
}
