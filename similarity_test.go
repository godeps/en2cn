package en2cn

import "testing"

func TestCalculateSimilarityPrefersCloserMatch(t *testing.T) {
	engine := &Engine{
		shengmuScores: map[string]map[string]float64{
			"t": {"t": 1, "d": 0.8},
			"d": {"d": 1},
			"":  {"": 1},
		},
		yunmuScores: map[string]map[string]float64{
			"a": {"a": 1, "o": 0.6},
			"o": {"o": 1},
			"":  {"": 1},
		},
	}

	target := []PinyinPart{
		{Shengmu: "t", Yunmu: "a"},
	}

	perfect := []PinyinPart{
		{Shengmu: "t", Yunmu: "a"},
	}

	closeButDifferent := []PinyinPart{
		{Shengmu: "d", Yunmu: "o"},
	}

	far := []PinyinPart{
		{Shengmu: "", Yunmu: ""},
	}

	scorePerfect := engine.CalculateSimilarity(target, perfect)
	scoreClose := engine.CalculateSimilarity(target, closeButDifferent)
	scoreFar := engine.CalculateSimilarity(target, far)

	if !(scorePerfect > scoreClose && scoreClose > scoreFar) {
		t.Fatalf("unexpected similarity ordering perfect=%f close=%f far=%f", scorePerfect, scoreClose, scoreFar)
	}
}
