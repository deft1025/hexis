package staticdata

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"hexis/internal/model"
)

type SkillProfile struct {
	ID string `json:"id"`
	model.SkillsResult
}

var championIDPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)
var skillImagePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+\.png$`)

func DecodeSkills(id, version string, content []byte) (SkillProfile, error) {
	if _, err := PatchVersion(version); err != nil || !championIDPattern.MatchString(id) {
		return SkillProfile{}, fmt.Errorf("skills require pinned version and champion ID")
	}
	type ability struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Image       struct {
			Full string `json:"full"`
		} `json:"image"`
	}
	var document struct {
		Version string `json:"version"`
		Data    map[string]struct {
			ID      string    `json:"id"`
			Title   string    `json:"title"`
			Tags    []string  `json:"tags"`
			Passive ability   `json:"passive"`
			Spells  []ability `json:"spells"`
		} `json:"data"`
	}
	if err := json.Unmarshal(content, &document); err != nil {
		return SkillProfile{}, err
	}
	c, ok := document.Data[id]
	if !ok || c.ID != id || document.Version != version || len(c.Spells) != 4 || c.Title == "" {
		return SkillProfile{}, fmt.Errorf("skill catalog version or champion mismatch: %s", id)
	}
	result := SkillProfile{ID: id, SkillsResult: model.SkillsResult{Version: version, Title: c.Title, Roles: []string{}, Skills: []model.Skill{}, SourceURL: "https://ddragon.leagueoflegends.com/cdn/" + version + "/data/zh_CN/champion/" + id + ".json"}}
	roles := map[string]string{"Tank": "坦克", "Mage": "法师", "Fighter": "战士", "Assassin": "刺客", "Marksman": "射手", "Support": "辅助"}
	for _, tag := range c.Tags {
		if name := roles[tag]; name != "" {
			result.Roles = append(result.Roles, name)
		}
	}
	for index, a := range append([]ability{c.Passive}, c.Spells...) {
		if strings.TrimSpace(a.Name) == "" || strings.TrimSpace(a.Description) == "" || !skillImagePattern.MatchString(a.Image.Full) {
			return SkillProfile{}, fmt.Errorf("invalid skill %d for %s", index, id)
		}
		folder := "spell"
		if index == 0 {
			folder = "passive"
		}
		result.Skills = append(result.Skills, model.Skill{Key: []string{"被动", "Q", "W", "E", "R"}[index], Name: a.Name, Description: PlainDescription(a.Description), IconURL: "https://ddragon.leagueoflegends.com/cdn/" + version + "/img/" + folder + "/" + a.Image.Full})
	}
	return result, nil
}

// Downloads run only in the importer, never in request handlers.
func FetchSkills(ctx context.Context, client *http.Client, version, out string, ids []string) ([]SkillProfile, []SourceFile, error) {
	if _, err := PatchVersion(version); err != nil {
		return nil, nil, err
	}
	if len(ids) == 0 {
		return nil, nil, fmt.Errorf("empty skills catalog")
	}
	if err := os.MkdirAll(filepath.Join(out, "raw", "skills"), 0755); err != nil {
		return nil, nil, err
	}
	profiles := make([]SkillProfile, len(ids))
	sources := make([]SourceFile, len(ids))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error
	jobs := make(chan int)
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				id := ids[index]
				fail := func(err error) { once.Do(func() { firstErr = err; cancel() }) }
				if !championIDPattern.MatchString(id) {
					fail(fmt.Errorf("invalid champion ID %q", id))
					continue
				}
				url := "https://ddragon.leagueoflegends.com/cdn/" + version + "/data/zh_CN/champion/" + id + ".json"
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				if err != nil {
					fail(err)
					continue
				}
				response, err := client.Do(req)
				if err != nil {
					fail(err)
					continue
				}
				content, readErr := io.ReadAll(io.LimitReader(response.Body, (2<<20)+1))
				response.Body.Close()
				if readErr != nil {
					fail(readErr)
					continue
				}
				if response.StatusCode != http.StatusOK || len(content) > 2<<20 {
					fail(fmt.Errorf("download skills %s: HTTP %d or oversized body", id, response.StatusCode))
					continue
				}
				profile, err := DecodeSkills(id, version, content)
				if err != nil {
					fail(err)
					continue
				}
				file := filepath.Join("raw", "skills", id+".json")
				if err := os.WriteFile(filepath.Join(out, file), content, 0644); err != nil {
					fail(err)
					continue
				}
				hash := sha256.Sum256(content)
				profiles[index] = profile
				sources[index] = SourceFile{URL: url, File: file, SHA256: hex.EncodeToString(hash[:])}
			}
		}()
	}
	for index := range ids {
		if ctx.Err() != nil {
			break
		}
		select {
		case jobs <- index:
		case <-ctx.Done():
		}
	}
	close(jobs)
	wg.Wait()
	if firstErr != nil {
		return nil, nil, firstErr
	}
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	content, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	if err := os.WriteFile(filepath.Join(out, "skills.json"), append(content, '\n'), 0644); err != nil {
		return nil, nil, err
	}
	return profiles, sources, nil
}
