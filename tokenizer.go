package en2cn

import (
	"strings"
)

var ipaCleaner = strings.NewReplacer("/", "", "ˈ", "", "ˌ", "", " ", "", "ː", "", "'", "", "\"", "")

// TokenizeIPA converts an IPA sequence into a slice of pinyin-like parts.
func (e *Engine) TokenizeIPA(ipa string) []PinyinPart {
	clean := cleanIPA(ipa)
	runes := []rune(clean)
	var parts []PinyinPart

	for idx := 0; idx < len(runes); {
		matchLen := 0
		var matched PinyinPart

		for length := len(runes) - idx; length > 0; length-- {
			segment := string(runes[idx : idx+length])
			if part, ok := e.ipaMap[segment]; ok {
				matchLen = length
				matched = part
				break
			}
		}

		if matchLen == 0 {
			idx++
			continue
		}

		parts = append(parts, matched)
		idx += matchLen
	}

	return normalizeParts(parts)
}

func cleanIPA(ipa string) string {
	lowered := strings.ToLower(ipa)
	return ipaCleaner.Replace(lowered)
}

func normalizeParts(parts []PinyinPart) []PinyinPart {
	if len(parts) == 0 {
		return nil
	}

	normalized := make([]PinyinPart, 0, len(parts))
	for _, part := range parts {
		switch {
		case part.Shengmu != "" && part.Yunmu != "":
			normalized = append(normalized, part)
		case part.Shengmu != "":
			normalized = append(normalized, part)
		case part.Yunmu != "":
			if len(normalized) > 0 && normalized[len(normalized)-1].Yunmu == "" {
				normalized[len(normalized)-1].Yunmu = part.Yunmu
			} else {
				normalized = append(normalized, part)
			}
		default:
			normalized = append(normalized, part)
		}
	}

	return normalized
}
