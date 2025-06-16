package processors

import (
	"fmt"

	"github.com/grtw2116/MinecraftModsLocalizer/internal/parsers"
	"github.com/grtw2116/MinecraftModsLocalizer/internal/translators"
)

// ProcessLanguageFile handles individual language files (.json, .lang, .snbt)
func ProcessLanguageFile(inputPath, outputPath, targetLang, engine, minecraftVersion string, dryRun bool, similarityThreshold float64, batchSize int) error {
	fmt.Printf("Processing language file: %s\n", inputPath)

	// Parse input file
	data, format, err := parsers.ParseFile(inputPath)
	if err != nil {
		return fmt.Errorf("error parsing file: %v", err)
	}

	fmt.Printf("Detected format: %v\n", format.String())
	fmt.Printf("Found %d translation keys\n", len(data))

	if dryRun {
		fmt.Println("\nDry run mode - showing sample keys:")
		count := 0
		for key, value := range data {
			if count >= 3 {
				break
			}
			fmt.Printf("  %s: %s\n", key, value)
			count++
		}
		if len(data) > 3 {
			fmt.Printf("  ... and %d more keys\n", len(data)-3)
		}
		fmt.Println("Use --dry-run=false to perform actual translation")
		return nil
	}

	// Create translator
	translator, err := translators.CreateTranslator(engine)
	if err != nil {
		return fmt.Errorf("error creating translator: %v", err)
	}

	// Perform translation
	fmt.Printf("Starting translation with %s engine (similarity threshold: %.1f, batch size: %d)...\n", engine, similarityThreshold, batchSize)
	translatedData, err := translators.TranslateDataWithSimilarity(data, translator, targetLang, similarityThreshold, batchSize)
	if err != nil {
		return fmt.Errorf("error during translation: %v", err)
	}

	// Write output file
	if err := parsers.WriteFile(outputPath, translatedData, format); err != nil {
		return fmt.Errorf("error writing output file: %v", err)
	}

	fmt.Printf("Processing completed: %s\n", outputPath)
	return nil
}
