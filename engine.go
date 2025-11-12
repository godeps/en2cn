package en2cn

import (
	"fmt"
	"strings"
)

// Convert transforms an English word into the most similar sounding Chinese word.
func (e *Engine) Convert(word string) (string, error) {
	if len(e.candidateDB) == 0 {
		return "", ErrNoCandidate
	}

	lowered := strings.ToLower(strings.TrimSpace(word))
	if override, ok := e.overrides[lowered]; ok && override != "" {
		return override, nil
	}
	ipa, ok := e.engIPADB[lowered]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrIPAUnavailable, word)
	}

	ipaParts := e.TokenizeIPA(ipa)
	if len(ipaParts) == 0 {
		return "", fmt.Errorf("empty ipa parts for %s", word)
	}

	best := ""
	bestScore := -1.0
	for hanzi, candidateParts := range e.candidateDB {
		score := e.CalculateSimilarity(ipaParts, candidateParts)
		if score > bestScore {
			bestScore = score
			best = hanzi
		}
	}

	if best == "" {
		return "", ErrNoCandidate
	}
	return best, nil
}
