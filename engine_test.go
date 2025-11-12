package en2cn

import "testing"

func TestConvertExamples(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("init engine: %v", err)
	}

	tests := map[string]string{
		"hello":  "哈喽",
		"coffee": "咖啡",
		"tiger":  "太格",
	}

	for word, want := range tests {
		got, err := engine.Convert(word)
		if err != nil {
			t.Fatalf("convert %s: %v", word, err)
		}
		if got != want {
			ipa := engine.engIPADB[word]
			parts := engine.TokenizeIPA(ipa)
			t.Fatalf("convert %s: want %s, got %s (ipa=%s parts=%+v)", word, want, got, ipa, parts)
		}
	}
}
