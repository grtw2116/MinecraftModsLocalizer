package processors

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
	"github.com/grtw2116/MinecraftModsLocalizer/pkg/translators"
)

type JARLanguageFile struct {
	Path     string
	Language string
	Data     parsers.TranslationData
	Format   parsers.FileFormat
}

func IsJARFile(filename string) bool {
	return strings.ToLower(filepath.Ext(filename)) == ".jar"
}

func ExtractLanguageFiles(jarPath string) ([]JARLanguageFile, error) {
	reader, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JAR file: %v", err)
	}
	defer reader.Close()

	var langFiles []JARLanguageFile

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		// Check if this is a language file in assets/*/lang/
		if strings.Contains(file.Name, "/lang/") && 
		   (strings.HasSuffix(file.Name, ".json") || strings.HasSuffix(file.Name, ".lang")) {
			
			langFile, err := extractSingleLanguageFile(file)
			if err != nil {
				fmt.Printf("Warning: Failed to extract %s: %v\n", file.Name, err)
				continue
			}
			
			langFiles = append(langFiles, langFile)
		}
	}

	return langFiles, nil
}

func extractSingleLanguageFile(file *zip.File) (JARLanguageFile, error) {
	rc, err := file.Open()
	if err != nil {
		return JARLanguageFile{}, err
	}
	defer rc.Close()

	// Determine format from original filename
	format := parsers.DetectFileFormat(file.Name)
	if format == parsers.FormatUnknown {
		return JARLanguageFile{}, fmt.Errorf("unsupported file format for %s", file.Name)
	}

	// Create temp file with proper extension
	ext := parsers.GetExtensionForFormat(format)
	tempFile, err := os.CreateTemp("", "lang_*"+ext)
	if err != nil {
		return JARLanguageFile{}, err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy content to temp file
	_, err = io.Copy(tempFile, rc)
	if err != nil {
		return JARLanguageFile{}, err
	}

	// Parse the language file
	data, _, err := parsers.ParseFile(tempFile.Name())
	if err != nil {
		return JARLanguageFile{}, err
	}

	// Extract language code from filename
	basename := filepath.Base(file.Name)
	langCode := strings.TrimSuffix(basename, filepath.Ext(basename))

	return JARLanguageFile{
		Path:     file.Name,
		Language: langCode,
		Data:     data,
		Format:   format,
	}, nil
}

func ProcessJARFile(jarPath, outputPath, targetLang, engine string, dryRun, extractOnly, resourcePack bool, similarityThreshold float64) error {
	fmt.Printf("Processing JAR file: %s\n", jarPath)
	
	// Extract language files
	langFiles, err := ExtractLanguageFiles(jarPath)
	if err != nil {
		return err
	}

	if len(langFiles) == 0 {
		return fmt.Errorf("no language files found in JAR")
	}

	fmt.Printf("Found %d language files:\n", len(langFiles))
	for _, lf := range langFiles {
		fmt.Printf("  %s (%s) - %d keys\n", lf.Path, lf.Language, len(lf.Data))
	}

	if extractOnly {
		return extractLanguageFilesToDirectory(langFiles, outputPath)
	}

	if dryRun {
		fmt.Println("\nDry run mode - would translate the following files:")
		for _, lf := range langFiles {
			if isSourceLanguage(lf.Language) {
				fmt.Printf("  %s -> %s_%s%s\n", lf.Path, lf.Language, targetLang, parsers.GetExtensionForFormat(lf.Format))
			}
		}
		return nil
	}

	// Find source language files (typically en_us)
	var sourceFiles []JARLanguageFile
	for _, lf := range langFiles {
		if isSourceLanguage(lf.Language) {
			sourceFiles = append(sourceFiles, lf)
		}
	}

	if len(sourceFiles) == 0 {
		return fmt.Errorf("no source language files (en_us) found in JAR")
	}

	// Create translator
	translator, err := translators.CreateTranslator(engine)
	if err != nil {
		return fmt.Errorf("error creating translator: %v", err)
	}

	// Translate each source file
	var translatedFiles []JARLanguageFile
	for _, sourceFile := range sourceFiles {
		fmt.Printf("\nTranslating %s...\n", sourceFile.Path)
		
		translatedData, err := translators.TranslateDataWithSimilarity(sourceFile.Data, translator, targetLang, similarityThreshold)
		if err != nil {
			return fmt.Errorf("error translating %s: %v", sourceFile.Path, err)
		}

		translatedFile := JARLanguageFile{
			Path:     strings.Replace(sourceFile.Path, sourceFile.Language, targetLang, 1),
			Language: targetLang,
			Data:     translatedData,
			Format:   sourceFile.Format,
		}
		translatedFiles = append(translatedFiles, translatedFile)
	}

	// Generate output
	if resourcePack {
		return generateResourcePack(translatedFiles, outputPath, targetLang)
	} else {
		return saveTranslatedFiles(translatedFiles, outputPath)
	}
}

func isSourceLanguage(lang string) bool {
	sourceLangs := []string{"en_us", "en_US", "en-us", "en-US"}
	for _, sl := range sourceLangs {
		if lang == sl {
			return true
		}
	}
	return false
}

func extractLanguageFilesToDirectory(langFiles []JARLanguageFile, outputDir string) error {
	if outputDir == "" {
		outputDir = "extracted_languages"
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for _, lf := range langFiles {
		// Create directory structure
		outputPath := filepath.Join(outputDir, lf.Path)
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}

		// Write file
		if err := parsers.WriteFile(outputPath, lf.Data, lf.Format); err != nil {
			return fmt.Errorf("failed to write %s: %v", outputPath, err)
		}
	}

	fmt.Printf("Extracted %d language files to %s\n", len(langFiles), outputDir)
	return nil
}

func saveTranslatedFiles(translatedFiles []JARLanguageFile, outputDir string) error {
	if outputDir == "" {
		outputDir = "translated_files"
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for _, tf := range translatedFiles {
		outputPath := filepath.Join(outputDir, filepath.Base(tf.Path))
		if err := parsers.WriteFile(outputPath, tf.Data, tf.Format); err != nil {
			return fmt.Errorf("failed to write %s: %v", outputPath, err)
		}
		fmt.Printf("Saved translated file: %s\n", outputPath)
	}

	return nil
}

func generateResourcePack(translatedFiles []JARLanguageFile, outputPath, targetLang string) error {
	if outputPath == "" {
		outputPath = fmt.Sprintf("resource_pack_%s", targetLang)
	}

	// Create resource pack directory structure
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return err
	}

	// Create pack.mcmeta
	packMcmeta := fmt.Sprintf(`{
  "pack": {
    "pack_format": 15,
    "description": "Translated language pack for %s"
  }
}`, targetLang)

	if err := os.WriteFile(filepath.Join(outputPath, "pack.mcmeta"), []byte(packMcmeta), 0644); err != nil {
		return err
	}

	// Save translated language files in proper structure
	for _, tf := range translatedFiles {
		// Extract mod namespace and create proper assets structure
		parts := strings.Split(tf.Path, "/")
		if len(parts) < 3 {
			continue
		}

		// Find assets index
		assetsIndex := -1
		for i, part := range parts {
			if part == "assets" {
				assetsIndex = i
				break
			}
		}

		if assetsIndex == -1 || assetsIndex+2 >= len(parts) {
			continue
		}

		namespace := parts[assetsIndex+1]
		resourcePackPath := filepath.Join(outputPath, "assets", namespace, "lang", fmt.Sprintf("%s%s", targetLang, parsers.GetExtensionForFormat(tf.Format)))

		// Create directory
		if err := os.MkdirAll(filepath.Dir(resourcePackPath), 0755); err != nil {
			return err
		}

		// Write language file
		if err := parsers.WriteFile(resourcePackPath, tf.Data, tf.Format); err != nil {
			return fmt.Errorf("failed to write resource pack file %s: %v", resourcePackPath, err)
		}
	}

	fmt.Printf("Generated resource pack: %s\n", outputPath)
	return nil
}