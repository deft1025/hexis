package service

import "testing"

func TestParsePage(t *testing.T) {
	for _, test := range []struct {
		limit   string
		offset  string
		wantL   int
		wantO   int
		wantErr bool
	}{
		{"", "", 100, 0, false},
		{"20", "5", 20, 5, false},
		{"101", "", 0, 0, true},
		{"", "-1", 0, 0, true},
	} {
		limit, offset, err := ParsePage(test.limit, test.offset)
		if limit != test.wantL || offset != test.wantO || (err != nil) != test.wantErr {
			t.Fatalf("ParsePage(%q, %q) = %d, %d, %v", test.limit, test.offset, limit, offset, err)
		}
	}
}
