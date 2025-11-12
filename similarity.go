package en2cn

import "math"

// CalculateSimilarity runs a Levenshtein-style dynamic programming algorithm with phoneme aware substitution scores.
func (e *Engine) CalculateSimilarity(target, candidate []PinyinPart) float64 {
	la := len(target)
	lb := len(candidate)
	if la == 0 && lb == 0 {
		return 1
	}

	dp := make([][]float64, la+1)
	for i := 0; i <= la; i++ {
		dp[i] = make([]float64, lb+1)
		dp[i][0] = float64(i)
	}
	for j := 0; j <= lb; j++ {
		dp[0][j] = float64(j)
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			subCost := e.substitutionCost(target[i-1], candidate[j-1])
			del := dp[i-1][j] + 1
			ins := dp[i][j-1] + 1
			sub := dp[i-1][j-1] + subCost
			dp[i][j] = math.Min(del, math.Min(ins, sub))
		}
	}

	maxLen := float64(max(la, lb))
	if maxLen == 0 {
		return 1
	}
	distance := dp[la][lb]
	similarity := 1 - distance/maxLen
	if similarity < 0 {
		return 0
	}
	return similarity
}

func (e *Engine) substitutionCost(src, dst PinyinPart) float64 {
	shengScore := e.lookupScore(e.shengmuScores, src.Shengmu, dst.Shengmu)
	yunScore := e.lookupScore(e.yunmuScores, src.Yunmu, dst.Yunmu)
	avgScore := (shengScore + yunScore) / 2
	return 1 - avgScore
}

func (e *Engine) lookupScore(table map[string]map[string]float64, a, b string) float64 {
	if a == b && a != "" {
		return 1
	}
	if row, ok := table[a]; ok {
		if score, ok := row[b]; ok {
			return score
		}
	}
	if row, ok := table[b]; ok {
		if score, ok := row[a]; ok {
			return score
		}
	}
	if a == "" && b == "" {
		return 1
	}
	if a == "" || b == "" {
		return 0
	}
	return 0.2
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
