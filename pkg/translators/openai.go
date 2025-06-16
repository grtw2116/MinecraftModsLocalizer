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

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/logger"
	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
)

type APIClient struct {
	APIKey  string
	Model   string
	BaseURL string
	client  *http.Client
}

type BatchConfig struct {
	Keys             []string
	BatchSize        int
	ProgressCallback ProgressCallback
	BatchCallback    BatchResultCallback
}

type OpenAITranslator struct {
	client *APIClient
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

func NewAPIClient() *APIClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	return &APIClient{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: baseURL,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func NewOpenAITranslator() *OpenAITranslator {
	return &OpenAITranslator{
		client: NewAPIClient(),
	}
}

func (t *OpenAITranslator) Translate(text, targetLang string) (string, error) {
	return t.TranslateWithExamples(text, targetLang, nil)
}

func (t *OpenAITranslator) TranslateBatch(texts []string, targetLang string, config *BatchConfig) ([]BatchTranslationResult, error) {
	if config == nil {
		config = &BatchConfig{BatchSize: 10}
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 10
	}
	return t.translateBatch(texts, targetLang, config)
}

func (t *OpenAITranslator) translateBatch(texts []string, targetLang string, config *BatchConfig) ([]BatchTranslationResult, error) {
	if t.client.APIKey == "" {
		return nil, fmt.Errorf("API key not found. Set OPENAI_API_KEY or ANTHROPIC_API_KEY environment variable")
	}

	if config.Keys != nil && len(config.Keys) != len(texts) {
		return nil, fmt.Errorf("number of keys (%d) must match number of texts (%d)", len(config.Keys), len(texts))
	}

	var results []BatchTranslationResult
	originalTotalTexts := len(texts)

	keys := config.Keys
	for len(texts) > 0 {
		currentBatchSize := config.BatchSize
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

		if config.BatchCallback != nil {
			config.BatchCallback(batchResults)
		}

		if config.ProgressCallback != nil {
			config.ProgressCallback(len(results), originalTotalTexts)
		}
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
	prompt := t.buildBatchPrompt(keys, texts, targetLang)
	response, err := t.client.sendRequest(prompt)
	if err != nil {
		return nil, err
	}
	return t.parseBatchResponse(keys, texts, response)
}

func (t *OpenAITranslator) buildBatchPrompt(keys, texts []string, targetLang string) string {
	var textList strings.Builder
	var prompt string

	if keys != nil && len(keys) == len(texts) {
		for i, text := range texts {
			textList.WriteString(fmt.Sprintf("%d. \"%s\": \"%s\"\n", i+1, keys[i], text))
		}
		prompt = fmt.Sprintf(`Translate the following Minecraft mod key-value pairs from English to %s. Keep translations natural and appropriate for gaming context. Only translate the VALUES (after the colon), keep the KEYS unchanged.

Please respond with the translated key-value pairs in the same numbered format:

%s

Return the translations in the exact same numbered format (1., 2., etc.), with keys unchanged and only values translated.`, parsers.GetLanguageNameForPrompt(targetLang), textList.String())
	} else {
		for i, text := range texts {
			textList.WriteString(fmt.Sprintf("%d. %s\n", i+1, text))
		}
		prompt = fmt.Sprintf(`Translate the following Minecraft mod texts from English to %s. Keep translations natural and appropriate for gaming context.

Please respond with ONLY the translated texts in the same numbered format:

%s

Return the translations in the exact same numbered format (1., 2., etc.), one per line.`, parsers.GetLanguageNameForPrompt(targetLang), textList.String())
	}

	logger.Debug("Batch translation prompt for %d texts:\n%s", len(texts), prompt)
	return prompt
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
	if t.client.APIKey == "" {
		return "", fmt.Errorf("API key not found. Set OPENAI_API_KEY or ANTHROPIC_API_KEY environment variable")
	}

	prompt := t.buildSinglePrompt(text, targetLang, examples)
	return t.client.sendRequest(prompt)
}

func (t *OpenAITranslator) buildSinglePrompt(text, targetLang string, examples []SimilarityMatch) string {
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
	logger.Debug("Single translation prompt:\n%s", prompt)
	return prompt
}

func (c *APIClient) sendRequest(prompt string) (string, error) {
	reqBody := OpenAIRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read API response: %v", err)
	}

	logger.Debug("API response status: %d, body length: %d", resp.StatusCode, len(body))
	logger.Debug("Raw API response body: %s", string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %v. Response body: %s", err, string(body))
	}

	if openaiResp.Error != nil {
		return "", fmt.Errorf("API error (%s): %s", openaiResp.Error.Type, openaiResp.Error.Message)
	}

	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("no translation received")
	}

	response := strings.TrimSpace(openaiResp.Choices[0].Message.Content)
	logger.Debug("API response: %s", response)
	return response, nil
}
