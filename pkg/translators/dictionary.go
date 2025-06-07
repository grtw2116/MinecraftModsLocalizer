package translators

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
)

type TermDictionary struct {
	Terms map[string]map[string]string `json:"terms"`
}

func NewTermDictionary() *TermDictionary {
	return &TermDictionary{
		Terms: make(map[string]map[string]string),
	}
}

func (td *TermDictionary) AddTerm(original, targetLang, translation string) {
	if td.Terms[targetLang] == nil {
		td.Terms[targetLang] = make(map[string]string)
	}
	td.Terms[targetLang][original] = translation
}

func (td *TermDictionary) GetTranslation(original, targetLang string) (string, bool) {
	if langTerms, exists := td.Terms[targetLang]; exists {
		if translation, found := langTerms[original]; found {
			return translation, true
		}
	}
	return "", false
}

func (td *TermDictionary) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, td)
}

func (td *TermDictionary) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(td, "", "  ")
	if err != nil {
		return err
	}
	
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

func levenshteinDistance(s1, s2 string) int {
	len1, len2 := utf8.RuneCountInString(s1), utf8.RuneCountInString(s2)
	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	runes1 := []rune(s1)
	runes2 := []rune(s2)

	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 1
			if runes1[i-1] == runes2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}
	return matrix[len1][len2]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func calculateSimilarity(s1, s2 string) float64 {
	maxLen := math.Max(float64(utf8.RuneCountInString(s1)), float64(utf8.RuneCountInString(s2)))
	if maxLen == 0 {
		return 1.0
	}
	distance := float64(levenshteinDistance(s1, s2))
	return 1.0 - distance/maxLen
}

func (td *TermDictionary) FindSimilarExamples(text, targetLang string, threshold float64, maxResults int) []SimilarityMatch {
	var matches []SimilarityMatch
	
	if langTerms, exists := td.Terms[targetLang]; exists {
		for original, translation := range langTerms {
			similarity := calculateSimilarity(text, original)
			if similarity >= threshold {
				matches = append(matches, SimilarityMatch{
					Example: TranslationExample{
						Original:    original,
						Translation: translation,
						Language:    targetLang,
					},
					Similarity: similarity,
				})
			}
		}
	}
	
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Similarity > matches[j].Similarity
	})
	
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}
	
	return matches
}

func TranslateDataWithSimilarity(data parsers.TranslationData, translator Translator, targetLang string, similarityThreshold float64) (parsers.TranslationData, error) {
	result := make(parsers.TranslationData)
	total := len(data)
	count := 0
	
	dict := NewTermDictionary()
	dictPath := "dictionary.json"
	if err := dict.LoadFromFile(dictPath); err != nil {
		fmt.Printf("Warning: Could not load term dictionary: %v\n", err)
	}
	
	maxExamples := 3
	
	for key, value := range data {
		count++
		fmt.Printf("Translating %d/%d: %s\n", count, total, key)
		
		if dictTranslation, found := dict.GetTranslation(value, targetLang); found {
			fmt.Printf("Using exact dictionary match for '%s'\n", value)
			result[key] = dictTranslation
		} else {
			similarExamples := dict.FindSimilarExamples(value, targetLang, similarityThreshold, maxExamples)
			
			var translatedValue string
			var err error
			
			if len(similarExamples) > 0 {
				fmt.Printf("Found %d similar examples for '%s'\n", len(similarExamples), value)
				if openaiTranslator, ok := translator.(*OpenAITranslator); ok {
					translatedValue, err = openaiTranslator.TranslateWithExamples(value, targetLang, similarExamples)
				} else {
					translatedValue, err = translator.Translate(value, targetLang)
				}
			} else {
				translatedValue, err = translator.Translate(value, targetLang)
			}
			
			if err != nil {
				fmt.Printf("Warning: Failed to translate '%s': %v\n", key, err)
				result[key] = value
			} else {
				result[key] = translatedValue
				dict.AddTerm(value, targetLang, translatedValue)
			}
			
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	if err := dict.SaveToFile(dictPath); err != nil {
		fmt.Printf("Warning: Could not save term dictionary: %v\n", err)
	}
	
	return result, nil
}