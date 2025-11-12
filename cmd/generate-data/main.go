package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/godeps/en2cn"
	pinyin "github.com/mozillazg/go-pinyin"
)

var (
	defaultCMUURL   = "https://svn.code.sf.net/p/cmusphinx/code/trunk/cmudict/cmudict-0.7b"
	defaultJiebaURL = "https://raw.githubusercontent.com/fxsjy/jieba/master/extra_dict/dict.txt.big"
)

func main() {
	outDir := flag.String("out", "data", "directory to place generated JSON files")
	cmuURL := flag.String("cmu-url", defaultCMUURL, "url for CMU dictionary")
	jiebaURL := flag.String("jieba-url", defaultJiebaURL, "url for Chinese lexicon")
	workDir := flag.String("work", "third_party", "directory to store downloaded sources")
	candidateLimit := flag.Int("candidate-limit", 120000, "maximum number of Chinese candidates to keep")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(*workDir, 0o755); err != nil {
		panic(err)
	}

	cmuPath := filepath.Join(*workDir, "cmudict-0.7b")
	if err := downloadIfNeeded(*cmuURL, cmuPath); err != nil {
		panic(fmt.Errorf("download cmu dict: %w", err))
	}

	jiebaPath := filepath.Join(*workDir, "dict.txt.big")
	if err := downloadIfNeeded(*jiebaURL, jiebaPath); err != nil {
		panic(fmt.Errorf("download jieba dict: %w", err))
	}

	fmt.Println("Building eng_ipa.json …")
	engIPA, err := buildEngIPA(cmuPath)
	if err != nil {
		panic(fmt.Errorf("build eng ipa: %w", err))
	}
	if err := writeJSON(filepath.Join(*outDir, "eng_ipa.json"), engIPA); err != nil {
		panic(fmt.Errorf("write eng ipa: %w", err))
	}

	fmt.Println("Building candidate_db.json …")
	candidateDB, err := buildCandidateDB(jiebaPath, *candidateLimit)
	if err != nil {
		panic(fmt.Errorf("build candidate db: %w", err))
	}
	if err := writeJSON(filepath.Join(*outDir, "candidate_db.json"), candidateDB); err != nil {
		panic(fmt.Errorf("write candidate db: %w", err))
	}
}

func downloadIfNeeded(url, dest string) error {
	if _, err := os.Stat(dest); err == nil {
		return nil
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}
	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dest)
}

var (
	wordLine = regexp.MustCompile(`^[A-Z]`)
)

func buildEngIPA(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string, 150000)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !wordLine.MatchString(line) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		word := sanitizeWord(fields[0])
		if word == "" {
			continue
		}
		if _, exists := result[word]; exists {
			continue
		}

		ipa := convertArpabetToIPA(fields[1:])
		if ipa == "" {
			continue
		}
		result[word] = ipa
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	for word, ipa := range manualIPA {
		result[word] = ipa
	}
	return result, nil
}

func sanitizeWord(s string) string {
	if idx := strings.IndexByte(s, '('); idx >= 0 {
		s = s[:idx]
	}
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return s
}

var arpabetToIPA = map[string]string{
	"AA":  "ɑ",
	"AE":  "æ",
	"AH":  "ʌ",
	"AO":  "ɔ",
	"AW":  "aʊ",
	"AY":  "aɪ",
	"EH":  "ɛ",
	"ER":  "ɝ",
	"EY":  "eɪ",
	"IH":  "ɪ",
	"IY":  "i",
	"OW":  "oʊ",
	"OY":  "ɔɪ",
	"UH":  "ʊ",
	"UW":  "u",
	"AX":  "ə",
	"AXR": "ɚ",
	"IX":  "ɪ",

	"B":  "b",
	"CH": "tʃ",
	"D":  "d",
	"DH": "ð",
	"F":  "f",
	"G":  "g",
	"HH": "h",
	"JH": "dʒ",
	"K":  "k",
	"L":  "l",
	"M":  "m",
	"N":  "n",
	"NG": "ŋ",
	"P":  "p",
	"R":  "r",
	"S":  "s",
	"SH": "ʃ",
	"T":  "t",
	"TH": "θ",
	"V":  "v",
	"W":  "w",
	"Y":  "y",
	"Z":  "z",
	"ZH": "ʒ",
	"Q":  "ʔ",
}

func convertArpabetToIPA(tokens []string) string {
	var buf bytes.Buffer
	buf.WriteByte('/')
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		tok = trimDigits(tok)
		if tok == "" {
			continue
		}
		ipa, ok := arpabetToIPA[tok]
		if !ok {
			return ""
		}
		buf.WriteString(ipa)
	}
	buf.WriteByte('/')
	return buf.String()
}

func trimDigits(s string) string {
	return strings.TrimRightFunc(s, func(r rune) bool {
		return unicode.IsDigit(r)
	})
}

var manualIPA = map[string]string{
	"openai": "/oʊpənaɪ/",
}

func buildCandidateDB(dictPath string, limit int) (map[string][]en2cn.PinyinPart, error) {
	file, err := os.Open(dictPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	args := pinyin.NewArgs()
	args.Heteronym = false
	args.Style = pinyin.Normal

	result := make(map[string][]en2cn.PinyinPart, limit)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(result) >= limit {
			break
		}
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		word := fields[0]
		if !isMostlyHan(word) {
			continue
		}

		py := pinyin.Pinyin(word, args)
		if len(py) == 0 {
			continue
		}
		parts := make([]en2cn.PinyinPart, 0, len(py))
		valid := true
		for _, syllables := range py {
			if len(syllables) == 0 {
				valid = false
				break
			}
			sm, ym := splitPinyin(syllables[0])
			if sm == "" && ym == "" {
				valid = false
				break
			}
			parts = append(parts, en2cn.PinyinPart{Shengmu: sm, Yunmu: ym})
		}
		if !valid || len(parts) == 0 {
			continue
		}
		result[word] = parts
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func isMostlyHan(word string) bool {
	count := 0
	for _, r := range word {
		if unicode.Is(unicode.Scripts["Han"], r) {
			count++
		}
	}
	return count*2 >= len([]rune(word))*2 && count > 0
}

var shengmuOrder = []string{
	"zh", "ch", "sh", "b", "p", "m", "f", "d", "t", "n", "l",
	"g", "k", "h", "j", "q", "x", "r", "z", "c", "s", "y", "w",
}

func splitPinyin(s string) (string, string) {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "ü", "v")
	for _, sm := range shengmuOrder {
		if strings.HasPrefix(s, sm) {
			return sm, strings.TrimPrefix(s, sm)
		}
	}
	return "", s
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
