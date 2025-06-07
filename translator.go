package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type Translator interface {
	Translate(text, targetLang string) (string, error)
}

type TermDictionary struct {
	Terms map[string]map[string]string `json:"terms"`
}

type TranslationExample struct {
	Original    string `json:"original"`
	Translation string `json:"translation"`
	Language    string `json:"language"`
}

type SimilarityMatch struct {
	Example    TranslationExample
	Similarity float64
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

type OpenAITranslator struct {
	APIKey  string
	Model   string
	BaseURL string
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
	Error   *OpenAIError `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func NewOpenAITranslator() *OpenAITranslator {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY") // For Claude or other compatible APIs
	}
	
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini" // Default to cost-effective model
	}
	
	return &OpenAITranslator{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: baseURL,
	}
}

func (t *OpenAITranslator) Translate(text, targetLang string) (string, error) {
	return t.TranslateWithExamples(text, targetLang, nil)
}

func (t *OpenAITranslator) TranslateWithExamples(text, targetLang string, examples []SimilarityMatch) (string, error) {
	if t.APIKey == "" {
		return "", fmt.Errorf("API key not found. Set OPENAI_API_KEY or ANTHROPIC_API_KEY environment variable")
	}

	prompt := fmt.Sprintf(`Translate the following Minecraft mod text from English to %s. Keep the translation natural and appropriate for gaming context.`, getLanguageName(targetLang))
	
	if len(examples) > 0 {
		prompt += "\n\nHere are some similar translation examples for reference:\n"
		for i, match := range examples {
			if i >= 3 {
				break
			}
			prompt += fmt.Sprintf("- \"%s\" → \"%s\" (similarity: %.1f%%)\n", 
				match.Example.Original, match.Example.Translation, match.Similarity*100)
		}
		prompt += "\nPlease maintain consistency with these examples when translating."
	}
	
	prompt += fmt.Sprintf(`

Text to translate: %s

Only return the translated text, nothing else.`, text)

	reqBody := OpenAIRequest{
		Model: t.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", t.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return "", err
	}

	if openaiResp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", openaiResp.Error.Message)
	}

	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("no translation received")
	}

	return strings.TrimSpace(openaiResp.Choices[0].Message.Content), nil
}

func getLanguageName(code string) string {
	langMap := map[string]string{
		"ja":    "Japanese",
		"ko":    "Korean",
		"zh-cn": "Simplified Chinese",
		"zh-tw": "Traditional Chinese",
		"fr":    "French",
		"de":    "German",
		"es":    "Spanish",
		"it":    "Italian",
		"pt":    "Portuguese",
		"ru":    "Russian",
	}
	
	if name, exists := langMap[strings.ToLower(code)]; exists {
		return name
	}
	return code // Return the code itself if not found
}

func createTranslator(engine string) (Translator, error) {
	switch strings.ToLower(engine) {
	case "openai":
		return NewOpenAITranslator(), nil
	case "google":
		return nil, fmt.Errorf("Google Translate not yet implemented")
	case "deepl":
		return nil, fmt.Errorf("DeepL not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported translation engine: %s", engine)
	}
}

func translateData(data TranslationData, translator Translator, targetLang string) (TranslationData, error) {
	return translateDataWithSimilarity(data, translator, targetLang, 0.6)
}

func translateDataWithSimilarity(data TranslationData, translator Translator, targetLang string, similarityThreshold float64) (TranslationData, error) {
	result := make(TranslationData)
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