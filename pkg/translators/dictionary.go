package translators

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// formatDuration formats a duration in a human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), math.Mod(d.Seconds(), 60))
	}
	return fmt.Sprintf("%.0fh%.0fm", d.Hours(), math.Mod(d.Minutes(), 60))
}

// showProgress displays a progress bar with elapsed time and ETA
func showProgress(current, total int, startTime time.Time) {
	if total == 0 {
		return
	}

	progress := float64(current) / float64(total)
	percentage := int(progress * 100)

	// Create progress bar (40 characters wide)
	barWidth := 40
	filledWidth := int(progress * float64(barWidth))
	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

	elapsed := time.Since(startTime)

	var eta string
	var rate string
	if current > 0 {
		avgTimePerItem := elapsed / time.Duration(current)
		remaining := time.Duration(total-current) * avgTimePerItem
		eta = formatDuration(remaining)
		rate = fmt.Sprintf("%.1f items/min", float64(current)/elapsed.Minutes())
	} else {
		eta = "calculating..."
		rate = "calculating..."
	}

	// Clear the line and print progress
	fmt.Printf("\r\033[K[%s] %3d%% (%d/%d) | Elapsed: %s | ETA: %s | Rate: %s",
		bar, percentage, current, total, formatDuration(elapsed), eta, rate)

	if current == total {
		fmt.Println() // New line when complete
	}
}

func TranslateDataWithSimilarity(data parsers.TranslationData, translator Translator, targetLang string, similarityThreshold float64, batchSize int) (parsers.TranslationData, error) {
	result := make(parsers.TranslationData)
	total := len(data)
	count := 0
	startTime := time.Now()

	dict := NewTermDictionary()
	dictPath := "dictionary.json"
	if err := dict.LoadFromFile(dictPath); err != nil {
		fmt.Printf("Warning: Could not load term dictionary: %v\n", err)
	}

	maxExamples := 3

	fmt.Printf("Starting translation of %d items...\n", total)
	showProgress(0, total, startTime)

	if batchSize > 1 {
		// Batch processing mode
		var pendingTexts []string
		var pendingKeys []string
		var pendingValues []string

		for key, value := range data {
			if dictTranslation, found := dict.GetTranslation(value, targetLang); found {
				result[key] = dictTranslation
				count++
				showProgress(count, total, startTime)
			} else {
				pendingTexts = append(pendingTexts, value)
				pendingKeys = append(pendingKeys, key)
				pendingValues = append(pendingValues, value)
			}
		}

		if len(pendingTexts) > 0 {
			if openaiTranslator, ok := translator.(*OpenAITranslator); ok {
				// Use the new method that accepts keys for key-value pairs in prompts
				// Create progress callback to update progress during batch processing
				progressCallback := func(completed, batchTotal int) {
					// Update count based on actual progress from batch translation
					currentCount := count + completed
					showProgress(currentCount, total, startTime)
				}
				
				// Create batch result callback to update dictionary after each batch
				batchResultCallback := func(batchResults []BatchTranslationResult) {
					// Save dictionary after each batch completion
					for _, batchResult := range batchResults {
						if batchResult.IsValid {
							dict.AddTerm(batchResult.Input, targetLang, batchResult.Output)
						}
					}
					if err := dict.SaveToFile(dictPath); err != nil {
						fmt.Printf("\nWarning: Failed to save dictionary: %v\n", err)
					}
				}
				
				batchResults, err := openaiTranslator.TranslateBatchWithKeysProgressAndCallback(pendingKeys, pendingTexts, targetLang, batchSize, progressCallback, batchResultCallback)
				if err != nil {
					return nil, fmt.Errorf("batch translation failed: %v", err)
				}

				// Create a map of input text to translation result
				resultMap := make(map[string]BatchTranslationResult)
				for _, batchResult := range batchResults {
					resultMap[batchResult.Input] = batchResult
				}

				// Process results in the same order as pendingKeys
				for i, key := range pendingKeys {
					if i < len(pendingValues) {
						value := pendingValues[i]
						if batchResult, found := resultMap[value]; found {
							if batchResult.IsValid {
								result[key] = batchResult.Output
								// Dictionary is already updated in batchResultCallback
							} else {
								fmt.Printf("\nError: Failed to translate key '%s' (text: '%s'): %s\n", key, value, batchResult.Error)
								result[key] = value
							}
						} else {
							fmt.Printf("\nError: No translation result found for key '%s' (text: '%s')\n", key, value)
							result[key] = value
						}
						count++
						// Progress is already updated by the progress callback during batch processing
					}
				}
			} else {
				// Fallback to individual processing for non-OpenAI translators
				for i, value := range pendingTexts {
					translatedValue, err := translator.Translate(value, targetLang)
					if err != nil {
						fmt.Printf("\nError: Failed to translate key '%s' (text: '%s'): %v\n", pendingKeys[i], value, err)
						result[pendingKeys[i]] = value
					} else {
						result[pendingKeys[i]] = translatedValue
						dict.AddTerm(value, targetLang, translatedValue)
					}
					count++
					showProgress(count, total, startTime)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	} else {
		// Individual processing mode (original behavior)
		for key, value := range data {
			count++

			if dictTranslation, found := dict.GetTranslation(value, targetLang); found {
				result[key] = dictTranslation
			} else {
				similarExamples := dict.FindSimilarExamples(value, targetLang, similarityThreshold, maxExamples)

				var translatedValue string
				var err error

				if len(similarExamples) > 0 {
					if openaiTranslator, ok := translator.(*OpenAITranslator); ok {
						translatedValue, err = openaiTranslator.TranslateWithExamples(value, targetLang, similarExamples)
					} else {
						translatedValue, err = translator.Translate(value, targetLang)
					}
				} else {
					translatedValue, err = translator.Translate(value, targetLang)
				}

				if err != nil {
					fmt.Printf("\nError: Failed to translate key '%s' (text: '%s'): %v\n", key, value, err)
					result[key] = value
				} else {
					result[key] = translatedValue
					dict.AddTerm(value, targetLang, translatedValue)
				}

				time.Sleep(100 * time.Millisecond)
			}

			// Update progress bar
			showProgress(count, total, startTime)
		}
	}

	if err := dict.SaveToFile(dictPath); err != nil {
		fmt.Printf("\nWarning: Could not save term dictionary: %v\n", err)
	}

	fmt.Printf("\nTranslation completed! Processed %d items in %s\n", total, formatDuration(time.Since(startTime)))

	return result, nil
}
