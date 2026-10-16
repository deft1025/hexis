package model

type Skill struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
}

type SkillsResult struct {
	Version   string   `json:"version"`
	Title     string   `json:"title"`
	Roles     []string `json:"roles"`
	Skills    []Skill  `json:"skills"`
	SourceURL string   `json:"source_url"`
}
