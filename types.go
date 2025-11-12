package en2cn

// PinyinPart represents a single consonant/vowel pair that forms a syllable-like chunk.
type PinyinPart struct {
	Shengmu string `json:"shengmu"`
	Yunmu   string `json:"yunmu"`
}

// Engine wires every component of the pipeline together.
type Engine struct {
	engIPADB      map[string]string
	ipaMap        map[string]PinyinPart
	candidateDB   map[string][]PinyinPart
	shengmuScores map[string]map[string]float64
	yunmuScores   map[string]map[string]float64
	overrides     map[string]string
}
