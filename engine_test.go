package en2cn

import "testing"

func TestConvertExamples(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("init engine: %v", err)
	}

	want := map[string]string{
		"hello":     "哈喽",
		"coffee":    "咖啡",
		"tiger":     "太格",
		"banana":    "巴娜娜",
		"tesla":     "特斯拉",
		"apple":     "苹果",
		"google":    "谷歌",
		"microsoft": "迈克软",
		"openai":    "欧朋爱",
	}

	for word, expected := range want {
		got, err := engine.Convert(word)
		if err != nil {
			t.Fatalf("convert %s: %v", word, err)
		}
		if got != expected {
			t.Fatalf("convert %s: expected %s got %s", word, expected, got)
		}
	}
}
