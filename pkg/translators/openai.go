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

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
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
	Choices []Choice     `json:"choices"`
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
	return t.TranslateBatchWithKeys(nil, texts, targetLang, batchSize)
}

func (t *OpenAITranslator) TranslateBatchWithKeys(keys, texts []string, targetLang string, batchSize int) ([]BatchTranslationResult, error) {
	if t.APIKey == "" {
		return nil, fmt.Errorf("API key not found. Set OPENAI_API_KEY or ANTHROPIC_API_KEY environment variable")
	}

	if batchSize <= 0 {
		batchSize = 1
	}

	// If keys are provided, they should match the number of texts
	if keys != nil && len(keys) != len(texts) {
		return nil, fmt.Errorf("number of keys (%d) must match number of texts (%d)", len(keys), len(texts))
	}

	var results []BatchTranslationResult

	for len(texts) > 0 {
		currentBatchSize := batchSize
		if len(texts) < currentBatchSize {
			currentBatchSize = len(texts)
		}

		textBatch := texts[:currentBatchSize]
		texts = texts[currentBatchSize:]

		var keyBatch []string
		if keys != nil {
			keyBatch = keys[:currentBatchSize]
			keys = keys[currentBatchSize:]
		}

		batchResults, err := t.translateBatchChunk(keyBatch, textBatch, targetLang)
		if err != nil {
			for _, text := range textBatch {
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

func (t *OpenAITranslator) translateBatchChunk(keys, texts []string, targetLang string) ([]BatchTranslationResult, error) {
	var textList strings.Builder
	var prompt string
	
	if keys != nil && len(keys) == len(texts) {
		// Format as key-value pairs
		for i, text := range texts {
			textList.WriteString(fmt.Sprintf("%d. \"%s\": \"%s\"\n", i+1, keys[i], text))
		}
		
		prompt = fmt.Sprintf(`Translate the following Minecraft mod key-value pairs from English to %s. Keep translations natural and appropriate for gaming context. Only translate the VALUES (after the colon), keep the KEYS unchanged.

Please respond with the translated key-value pairs in the same numbered format:

%s

Return the translations in the exact same numbered format (1., 2., etc.), with keys unchanged and only values translated.`, parsers.GetLanguageNameForPrompt(targetLang), textList.String())
	} else {
		// Fallback to text-only format
		for i, text := range texts {
			textList.WriteString(fmt.Sprintf("%d. %s\n", i+1, text))
		}
		
		prompt = fmt.Sprintf(`Translate the following Minecraft mod texts from English to %s. Keep translations natural and appropriate for gaming context.

Please respond with ONLY the translated texts in the same numbered format:

%s

Return the translations in the exact same numbered format (1., 2., etc.), one per line.`, parsers.GetLanguageNameForPrompt(targetLang), textList.String())
	}

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

	return t.parseBatchResponse(keys, texts, strings.TrimSpace(openaiResp.Choices[0].Message.Content))
}

func (t *OpenAITranslator) parseBatchResponse(keys, inputs []string, response string) ([]BatchTranslationResult, error) {
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
				lineContent := strings.TrimSpace(line[len(numberPrefix):])
				
				// Check if we have keys (key-value format) or just text
				if keys != nil && len(keys) == len(inputs) {
					// Parse key-value format: "key": "translated_value"
					colonIndex := strings.Index(lineContent, ":")
					if colonIndex > 0 {
						// Extract the value part after the colon
						valuePart := strings.TrimSpace(lineContent[colonIndex+1:])
						// Remove surrounding quotes if present
						if len(valuePart) >= 2 && valuePart[0] == '"' && valuePart[len(valuePart)-1] == '"' {
							output = valuePart[1 : len(valuePart)-1]
						} else {
							output = valuePart
						}
						isValid = output != "" && output != input
					} else {
						// Fallback: treat as plain text if parsing fails
						output = lineContent
						isValid = output != "" && output != input
					}
				} else {
					// Plain text format
					output = lineContent
					isValid = output != "" && output != input
				}
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

	prompt := fmt.Sprintf(`Translate the following Minecraft mod text from English to %s. Keep the translation natural and appropriate for gaming context.`, parsers.GetLanguageNameForPrompt(targetLang))

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

