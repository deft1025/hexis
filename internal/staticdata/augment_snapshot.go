package staticdata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RebuildAugments uses the original pinned, hash-verified inputs without downloading
// or overwriting the snapshot that was previously imported.
func RebuildAugments(dir string) (*Manifest, []Profile, []Profile, []VariableDescription, error) {
	var manifest Manifest
	readJSON := func(name string, target any) error {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		return json.Unmarshal(data, target)
	}
	if err := readJSON("manifest.json", &manifest); err != nil {
		return nil, nil, nil, nil, err
	}
	patch, err := PatchVersion(manifest.Version)
	if err != nil || patch != manifest.Patch || manifest.Mode != mode || manifest.Locale != "zh_CN" {
		return nil, nil, nil, nil, fmt.Errorf("snapshot must have a pinned KIWI zh_CN version")
	}
	for _, name := range []string{"augment-lists.json", "kiwi.bin.json", "strings-zh.json"} {
		path := filepath.Join("raw", name)
		var expected string
		for _, source := range manifest.Sources {
			if source.File == path {
				expected = source.SHA256
				if !strings.HasPrefix(source.URL, "https://raw.communitydragon.org/"+patch+"/") {
					return nil, nil, nil, nil, fmt.Errorf("unpinned source %s", path)
				}
			}
		}
		data, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			return nil, nil, nil, nil, err
		}
		hash := sha256.Sum256(data)
		if expected == "" || hex.EncodeToString(hash[:]) != expected {
			return nil, nil, nil, nil, fmt.Errorf("snapshot hash mismatch: %s", path)
		}
	}
	var lists []augmentList
	var raw map[string]json.RawMessage
	var translations struct {
		Entries map[string]string `json:"entries"`
	}
	var previous []Profile
	for _, input := range []struct {
		name   string
		target any
	}{
		{"raw/augment-lists.json", &lists}, {"raw/kiwi.bin.json", &raw}, {"raw/strings-zh.json", &translations}, {"augments.json", &previous},
	} {
		if err := readJSON(input.name, input.target); err != nil {
			return nil, nil, nil, nil, err
		}
	}
	bin, err := decodeBin(raw)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	profiles, unresolved, err := convertAugments(patch, lists, bin, translations.Entries)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if len(profiles) != manifest.Augments || len(previous) != len(profiles) {
		return nil, nil, nil, nil, fmt.Errorf("augment snapshot count mismatch")
	}
	return &manifest, previous, profiles, unresolved, nil
}
