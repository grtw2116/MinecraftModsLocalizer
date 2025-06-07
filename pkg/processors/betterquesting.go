package processors

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
	"github.com/grtw2116/MinecraftModsLocalizer/pkg/translators"
)

func ProcessBetterQuestingFile(inputPath, outputPath, targetLang, engine string, dryRun bool, similarityThreshold float64) error {
	fmt.Printf("Processing BetterQuesting file: %s\n", inputPath)
	
	// Parse BetterQuesting file
	bqFile, err := parsers.ParseBetterQuestingFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse BetterQuesting file: %v", err)
	}
	
	// Extract translatable text
	translations := parsers.ExtractBetterQuestingTranslations(bqFile)
	
	fmt.Printf("Found %d translatable strings\n", len(translations))
	
	if dryRun {
		fmt.Println("\nDry run mode - showing sample translatable strings:")
		count := 0
		for key, value := range translations {
			if count >= 5 {
				break
			}
			fmt.Printf("  %s: %s\n", key, value)
			count++
		}
		if len(translations) > 5 {
			fmt.Printf("  ... and %d more strings\n", len(translations)-5)
		}
		fmt.Println("Use --dry-run=false to perform actual translation")
		return nil
	}
	
	// Create translator
	translator, err := translators.CreateTranslator(engine)
	if err != nil {
		return fmt.Errorf("error creating translator: %v", err)
	}
	
	// Translate strings
	fmt.Printf("Starting translation with %s engine...\n", engine)
	translatedData, err := translators.TranslateDataWithSimilarity(translations, translator, targetLang, similarityThreshold)
	if err != nil {
		return fmt.Errorf("error during translation: %v", err)
	}
	
	// Apply translations to BetterQuesting structure
	parsers.ApplyBetterQuestingTranslations(bqFile, translatedData)
	
	// Generate output filename if not specified
	if outputPath == "" {
		ext := filepath.Ext(inputPath)
		base := strings.TrimSuffix(inputPath, ext)
		outputPath = fmt.Sprintf("%s_%s%s", base, targetLang, ext)
	}
	
	// Write translated file
	if err := parsers.WriteBetterQuestingFile(outputPath, bqFile); err != nil {
		return fmt.Errorf("error writing output file: %v", err)
	}
	
	fmt.Printf("Translation completed: %s\n", outputPath)
	return nil
}