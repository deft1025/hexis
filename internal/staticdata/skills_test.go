package staticdata

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func skillFixture(id string) []byte {
	type ability struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Image       map[string]string `json:"image"`
	}
	a := ability{Name: "技能", Description: "<b>恢复</b>生命值", Image: map[string]string{"full": "Ahri.png"}}
	content, _ := json.Marshal(map[string]any{"version": "16.20.1", "data": map[string]any{id: map[string]any{"id": id, "title": "阿狸", "tags": []string{"Mage"}, "passive": a, "spells": []ability{a, a, a, a}}}})
	return content
}

func TestSkillsVersionAndPlainDescriptions(t *testing.T) {
	p, err := DecodeSkills("Ahri", "16.20.1", skillFixture("Ahri"))
	if err != nil || len(p.Skills) != 5 || p.Skills[0].Description != "恢复生命值" || p.Roles[0] != "法师" {
		t.Fatalf("%+v %v", p, err)
	}
	for _, test := range []struct {
		id, version string
		content     []byte
	}{
		{"Ahri", "16.19.1", skillFixture("Ahri")}, {"Ahri", "latest", skillFixture("Ahri")}, {"../Ahri", "16.20.1", skillFixture("Ahri")},
		{"Ahri", "16.20.1", skillFixture("Lux")}, {"Ahri", "16.20.1", []byte(`{"version":"16.20.1","data":{}}`)},
		{"Ahri", "16.20.1", []byte(strings.ReplaceAll(string(skillFixture("Ahri")), "Ahri.png", "../Ahri.png"))},
	} {
		if _, err := DecodeSkills(test.id, test.version, test.content); err == nil {
			t.Fatalf("accepted invalid skills %+v", test)
		}
	}
}

type skillTransport func(*http.Request) (*http.Response, error)

func (f skillTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetchSkillsWritesPinnedSourceAndCompleteCatalog(t *testing.T) {
	out := t.TempDir()
	client := &http.Client{Transport: skillTransport(func(r *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(r.URL.String(), "https://ddragon.leagueoflegends.com/cdn/16.20.1/data/zh_CN/champion/") {
			t.Errorf("unpinned URL %s", r.URL)
		}
		id := strings.TrimSuffix(filepath.Base(r.URL.Path), ".json")
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(skillFixture(id))))}, nil
	})}
	profiles, sources, err := FetchSkills(context.Background(), client, "16.20.1", out, []string{"Ahri", "Lux"})
	if err != nil || len(profiles) != 2 || len(sources) != 2 {
		t.Fatalf("%v %v %v", profiles, sources, err)
	}
	for _, s := range sources {
		if len(s.SHA256) != 64 {
			t.Fatal("missing source hash")
		}
		if _, err := os.Stat(filepath.Join(out, s.File)); err != nil {
			t.Fatal(err)
		}
	}
	var stored []SkillProfile
	content, err := os.ReadFile(filepath.Join(out, "skills.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, &stored); err != nil || len(stored) != 2 {
		t.Fatalf("%s %v", content, err)
	}
	client.Transport = skillTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("missing"))}, nil
	})
	failureDir := t.TempDir()
	if _, _, err := FetchSkills(context.Background(), client, "16.20.1", failureDir, []string{"Ahri"}); err == nil {
		t.Fatal("accepted missing catalog")
	}
	if _, err := os.Stat(filepath.Join(failureDir, "skills.json")); !os.IsNotExist(err) {
		t.Fatal("wrote partial catalog")
	}
}
