package en2cn

import "testing"

func TestTokenizeIPAGreedyMatch(t *testing.T) {
	engine := &Engine{
		ipaMap: map[string]PinyinPart{
			"tʃ": {Shengmu: "ch"},
			"t":  {Shengmu: "t"},
			"ʃ":  {Shengmu: "sh"},
			"oʊ": {Yunmu: "ou"},
			"o":  {Yunmu: "o"},
		},
	}

	tests := []struct {
		ipa      string
		expected []PinyinPart
	}{
		{
			ipa: "/tʃoʊ/",
			expected: []PinyinPart{
				{Shengmu: "ch", Yunmu: "ou"},
			},
		},
		{
			ipa: "/tʃo/",
			expected: []PinyinPart{
				{Shengmu: "ch", Yunmu: "o"},
			},
		},
	}

	for _, tt := range tests {
		got := engine.TokenizeIPA(tt.ipa)
		if len(got) != len(tt.expected) {
			t.Fatalf("tokenize %s: expected %d parts got %d", tt.ipa, len(tt.expected), len(got))
		}
		for i := range tt.expected {
			if got[i] != tt.expected[i] {
				t.Fatalf("tokenize %s: expected %+v got %+v", tt.ipa, tt.expected[i], got[i])
			}
		}
	}
}
