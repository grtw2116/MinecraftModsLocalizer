package processors

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
	"github.com/grtw2116/MinecraftModsLocalizer/pkg/translators"
)

func ProcessBetterQuestingFile(inputPath, outputPath, targetLang, engine, minecraftVersion string, dryRun bool, similarityThreshold float64, batchSize int) error {
	fmt.Printf("Processing BetterQuesting file: %s\n", inputPath)

	// Try to extract translations directly using NBT format first
	translations, err := parsers.ExtractNBTBetterQuestingTranslations(inputPath)
	isNBTFormat := err == nil && len(translations) > 0

	if !isNBTFormat {
		// Fallback to standard format parsing
		bqFile, parseErr := parsers.ParseBetterQuestingFile(inputPath)
		if parseErr != nil {
			return fmt.Errorf("failed to parse BetterQuesting file (tried both NBT and standard formats): NBT error: %v, Standard error: %v", err, parseErr)
		}

		// Extract translatable text using standard format
		translations = parsers.ExtractBetterQuestingTranslations(bqFile)
	}

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
	translatedData, err := translators.TranslateDataWithSimilarity(translations, translator, targetLang, similarityThreshold, batchSize)
	if err != nil {
		return fmt.Errorf("error during translation: %v", err)
	}

	// Generate output filename if not specified
	if outputPath == "" {
		ext := filepath.Ext(inputPath)
		base := strings.TrimSuffix(inputPath, ext)
		outputPath = fmt.Sprintf("%s_%s%s", base, targetLang, ext)
	}

	// Apply translations based on format
	if isNBTFormat {
		// Copy input file to output path first, then apply translations
		if err := copyFile(inputPath, outputPath); err != nil {
			return fmt.Errorf("error copying file: %v", err)
		}
		if err := parsers.ApplyNBTBetterQuestingTranslations(outputPath, translatedData); err != nil {
			return fmt.Errorf("error applying NBT translations: %v", err)
		}
	} else {
		// Apply translations to standard format
		bqFile, _ := parsers.ParseBetterQuestingFile(inputPath) // We know this works from earlier
		parsers.ApplyBetterQuestingTranslations(bqFile, translatedData)

		// Write translated file
		if err := parsers.WriteBetterQuestingFile(outputPath, bqFile); err != nil {
			return fmt.Errorf("error writing output file: %v", err)
		}
	}

	fmt.Printf("Translation completed: %s\n", outputPath)
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
