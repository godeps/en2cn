package en2cn

import (
	"embed"
	"encoding/json"
	"fmt"
)

var (
	//go:embed data/eng_ipa.json
	englishIPAFile embed.FS
	//go:embed data/ipa_to_parts.json
	ipaPartsFile embed.FS
	//go:embed data/candidate_db.json
	candidateFile embed.FS
	//go:embed data/shengmu_similarity.json
	shengmuSimFile embed.FS
	//go:embed data/yunmu_similarity.json
	yunmuSimFile embed.FS
)

// NewEngine loads all embedded data files and prepares an Engine instance.
func NewEngine() (*Engine, error) {
	e := &Engine{}
	if err := loadJSON(englishIPAFile, "data/eng_ipa.json", &e.engIPADB); err != nil {
		return nil, fmt.Errorf("load english ipa db: %w", err)
	}
	if err := loadJSON(ipaPartsFile, "data/ipa_to_parts.json", &e.ipaMap); err != nil {
		return nil, fmt.Errorf("load ipa map: %w", err)
	}
	if err := loadJSON(candidateFile, "data/candidate_db.json", &e.candidateDB); err != nil {
		return nil, fmt.Errorf("load candidate db: %w", err)
	}
	if err := loadJSON(shengmuSimFile, "data/shengmu_similarity.json", &e.shengmuScores); err != nil {
		return nil, fmt.Errorf("load shengmu score: %w", err)
	}
	if err := loadJSON(yunmuSimFile, "data/yunmu_similarity.json", &e.yunmuScores); err != nil {
		return nil, fmt.Errorf("load yunmu score: %w", err)
	}
	return e, nil
}

func loadJSON(fs embed.FS, name string, target any) error {
	data, err := fs.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
