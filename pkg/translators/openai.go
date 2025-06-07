package translators

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

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

func (t *OpenAITranslator) TranslateBatch(texts []string, targetLang string) ([]BatchTranslationResult, error) {
	return t.TranslateBatchWithSize(texts, targetLang, 10)
}

func (t *OpenAITranslator) TranslateBatchWithSize(texts []string, targetLang string, batchSize int) ([]BatchTranslationResult, error) {
	if t.APIKey == "" {
		return nil, fmt.Errorf("API key not found. Set OPENAI_API_KEY or ANTHROPIC_API_KEY environment variable")
	}

	if batchSize <= 0 {
		batchSize = 1
	}

	var results []BatchTranslationResult
	
	for len(texts) > 0 {
		currentBatchSize := batchSize
		if len(texts) < currentBatchSize {
			currentBatchSize = len(texts)
		}
		
		batch := texts[:currentBatchSize]
		texts = texts[currentBatchSize:]
		
		batchResults, err := t.translateBatchChunk(batch, targetLang)
		if err != nil {
			for _, text := range batch {
				results = append(results, BatchTranslationResult{
					Input:   text,
					Output:  text,
					IsValid: false,
					Error:   err.Error(),
				})
			}
			continue
		}
		
		results = append(results, batchResults...)
	}
	
	originalTexts := make([]string, len(results))
	for i, result := range results {
		originalTexts[i] = result.Input
	}
	validation := ValidateBatchResults(originalTexts, results)
	if !validation.IsValid {
		for _, missingInput := range validation.MissingInputs {
			singleResult, err := t.Translate(missingInput, targetLang)
			if err != nil {
				results = append(results, BatchTranslationResult{
					Input:   missingInput,
					Output:  missingInput,
					IsValid: false,
					Error:   err.Error(),
				})
			} else {
				results = append(results, BatchTranslationResult{
					Input:   missingInput,
					Output:  singleResult,
					IsValid: true,
				})
			}
		}
	}
	
	return results, nil
}

func (t *OpenAITranslator) translateBatchChunk(texts []string, targetLang string) ([]BatchTranslationResult, error) {
	var textList strings.Builder
	for i, text := range texts {
		textList.WriteString(fmt.Sprintf("%d. %s\n", i+1, text))
	}
	
	prompt := fmt.Sprintf(`Translate the following Minecraft mod texts from English to %s. Keep translations natural and appropriate for gaming context.

Please respond with ONLY the translated texts in the same numbered format:

%s

Return the translations in the exact same numbered format (1., 2., etc.), one per line.`, getLanguageName(targetLang), textList.String())

	reqBody := OpenAIRequest{
		Model: t.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", t.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, err
	}

	if openaiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openaiResp.Error.Message)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no translation received")
	}

	return t.parseBatchResponse(texts, strings.TrimSpace(openaiResp.Choices[0].Message.Content))
}

func (t *OpenAITranslator) parseBatchResponse(inputs []string, response string) ([]BatchTranslationResult, error) {
	lines := strings.Split(response, "\n")
	var results []BatchTranslationResult
	
	for i, input := range inputs {
		var output string
		var isValid bool
		var errorMsg string
		
		if i < len(lines) {
			line := strings.TrimSpace(lines[i])
			numberPrefix := fmt.Sprintf("%d.", i+1)
			if strings.HasPrefix(line, numberPrefix) {
				output = strings.TrimSpace(line[len(numberPrefix):])
				isValid = output != "" && output != input
			} else {
				output = input
				isValid = false
				errorMsg = "Failed to parse translation from response"
			}
		} else {
			output = input
			isValid = false
			errorMsg = "Missing translation in response"
		}
		
		results = append(results, BatchTranslationResult{
			Input:   input,
			Output:  output,
			IsValid: isValid,
			Error:   errorMsg,
		})
	}
	
	return results, nil
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