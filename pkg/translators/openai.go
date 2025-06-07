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