package translators

import (
	"fmt"
	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
)

type Translator interface {
	Translate(text, targetLang string) (string, error)
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

func TranslateData(data parsers.TranslationData, translator Translator, targetLang string) (parsers.TranslationData, error) {
	return TranslateDataWithSimilarity(data, translator, targetLang, 0.6)
}