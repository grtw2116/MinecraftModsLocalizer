package translators

import (
	"fmt"
	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
)

type Translator interface {
	Translate(text, targetLang string) (string, error)
	TranslateBatch(texts []string, targetLang string) ([]BatchTranslationResult, error)
}

type BatchTranslationResult struct {
	Input      string            `json:"input"`
	Output     string            `json:"output"`
	IsValid    bool              `json:"is_valid"`
	Error      string            `json:"error,omitempty"`
	Validation *ValidationResult `json:"validation,omitempty"`
}

type ValidationResult struct {
	IsValid       bool     `json:"is_valid"`
	MissingInputs []string `json:"missing_inputs,omitempty"`
	MissingCount  int      `json:"missing_count"`
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

func CreateTranslator(engine string) (Translator, error) {
	switch engine {
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

func ValidateBatchResults(inputs []string, results []BatchTranslationResult) *ValidationResult {
	inputSet := make(map[string]bool)
	for _, input := range inputs {
		inputSet[input] = true
	}

	var missingInputs []string
	resultSet := make(map[string]bool)

	for _, result := range results {
		resultSet[result.Input] = true
	}

	for input := range inputSet {
		if !resultSet[input] {
			missingInputs = append(missingInputs, input)
		}
	}

	return &ValidationResult{
		IsValid:       len(missingInputs) == 0,
		MissingInputs: missingInputs,
		MissingCount:  len(missingInputs),
	}
}

func TranslateData(data parsers.TranslationData, translator Translator, targetLang string) (parsers.TranslationData, error) {
	return TranslateDataWithSimilarity(data, translator, targetLang, 0.6, 1)
}
