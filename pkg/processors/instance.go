package processors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
	"github.com/grtw2116/MinecraftModsLocalizer/pkg/translators"
)

type MinecraftInstance struct {
	RootPath string
	ModsPath string
	JARFiles []string
}

func DetectMinecraftInstance(inputPath string) (*MinecraftInstance, error) {
	// Check if the input path exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", inputPath)
	}

	// Check if it's a single JAR file (backward compatibility)
	if IsJARFile(inputPath) {
		return &MinecraftInstance{
			RootPath: filepath.Dir(inputPath),
			ModsPath: filepath.Dir(inputPath),
			JARFiles: []string{inputPath},
		}, nil
	}

	// Check if it's a single language file (backward compatibility)
	if isLanguageFile(inputPath) {
		return nil, fmt.Errorf("single language files not supported in instance mode, use individual file mode")
	}

	// Assume it's a Minecraft instance directory
	instance := &MinecraftInstance{
		RootPath: inputPath,
		ModsPath: filepath.Join(inputPath, "mods"),
	}

	// Verify this looks like a Minecraft instance
	if err := validateMinecraftInstance(instance); err != nil {
		return nil, err
	}

	// Scan for JAR files in mods directory
	jarFiles, err := findJARFiles(instance.ModsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan mods directory: %v", err)
	}

	instance.JARFiles = jarFiles
	return instance, nil
}

func validateMinecraftInstance(instance *MinecraftInstance) error {
	// Check if mods directory exists
	if _, err := os.Stat(instance.ModsPath); os.IsNotExist(err) {
		return fmt.Errorf("mods directory not found: %s (not a valid Minecraft instance)", instance.ModsPath)
	}

	// Optional: Check for other Minecraft instance indicators
	// These are common files/directories in Minecraft instances
	indicators := []string{
		"versions",
		"saves",
		"resourcepacks",
		"config",
		"logs",
	}

	foundIndicators := 0
	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(instance.RootPath, indicator)); err == nil {
			foundIndicators++
		}
	}

	// If we don't find any indicators, issue a warning but continue
	if foundIndicators == 0 {
		fmt.Printf("Warning: %s doesn't look like a typical Minecraft instance directory\n", instance.RootPath)
		fmt.Printf("Expected to find directories like: %s\n", strings.Join(indicators, ", "))
		fmt.Printf("Continuing anyway as mods directory was found...\n")
	}

	return nil
}

func findJARFiles(modsPath string) ([]string, error) {
	var jarFiles []string

	err := filepath.Walk(modsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a JAR file
		if IsJARFile(path) {
			jarFiles = append(jarFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return jarFiles, nil
}

func ProcessMinecraftInstance(instancePath, outputPath, targetLang, engine string, dryRun, extractOnly, resourcePack bool, similarityThreshold float64, batchSize int) error {
	fmt.Printf("Processing Minecraft instance: %s\n", instancePath)

	// Detect and validate the Minecraft instance
	instance, err := DetectMinecraftInstance(instancePath)
	if err != nil {
		return err
	}

	// Look for BetterQuesting files
	bqFiles, err := parsers.FindBetterQuestingFiles(instancePath)
	if err != nil {
		fmt.Printf("Warning: Failed to search for BetterQuesting files: %v\n", err)
	}

	if len(bqFiles) > 0 {
		fmt.Printf("Found %d BetterQuesting files:\n", len(bqFiles))
		for i, bqFile := range bqFiles {
			relPath, _ := filepath.Rel(instancePath, bqFile)
			fmt.Printf("  %d. %s\n", i+1, relPath)
		}
	}

	if len(instance.JARFiles) == 0 && len(bqFiles) == 0 {
		return fmt.Errorf("no JAR files or BetterQuesting files found in instance: %s", instancePath)
	}

	if len(instance.JARFiles) > 0 {
		fmt.Printf("Found %d JAR files in mods directory:\n", len(instance.JARFiles))
		for i, jarFile := range instance.JARFiles {
			relPath, _ := filepath.Rel(instance.RootPath, jarFile)
			fmt.Printf("  %d. %s\n", i+1, relPath)
		}
	}

	if dryRun {
		fmt.Println("\nDry run mode - would process the above files")
		return nil
	}

	// Set up output directory structure
	if outputPath == "" {
		if resourcePack {
			outputPath = filepath.Join(instance.RootPath, fmt.Sprintf("resource_pack_%s", targetLang))
		} else if extractOnly {
			outputPath = filepath.Join(instance.RootPath, "extracted_languages")
		} else {
			outputPath = filepath.Join(instance.RootPath, "translated_files")
		}
	}

	fmt.Printf("Output directory: %s\n", outputPath)

	// Process each JAR file first to ensure consistent terminology for BetterQuesting translation
	var allTranslatedFiles []JARLanguageFile
	var totalExtracted int

	for i, jarFile := range instance.JARFiles {
		fmt.Printf("\n=== Processing JAR %d/%d: %s ===\n", i+1, len(instance.JARFiles), filepath.Base(jarFile))

		if extractOnly {
			// Extract to separate directory for each mod
			modName := getModNameFromJAR(jarFile)
			modOutputPath := filepath.Join(outputPath, modName)

			if err := processSingleJARForExtraction(jarFile, modOutputPath); err != nil {
				fmt.Printf("Warning: Failed to process %s: %v\n", filepath.Base(jarFile), err)
				continue
			}
		} else {
			// Process for translation
			translatedFiles, err := processSingleJARForTranslation(jarFile, targetLang, engine, similarityThreshold, batchSize)
			if err != nil {
				fmt.Printf("Warning: Failed to process %s: %v\n", filepath.Base(jarFile), err)
				continue
			}

			allTranslatedFiles = append(allTranslatedFiles, translatedFiles...)
		}

		totalExtracted++
	}

	// Process BetterQuesting files after JAR files to benefit from consistent terminology
	if len(bqFiles) > 0 && !extractOnly {
		bqOutputPath := filepath.Join(outputPath, "betterquesting")
		if err := os.MkdirAll(bqOutputPath, 0755); err != nil {
			return fmt.Errorf("failed to create BetterQuesting output directory: %v", err)
		}

		for i, bqFile := range bqFiles {
			fmt.Printf("\n=== Processing BetterQuesting %d/%d: %s ===\n", i+1, len(bqFiles), filepath.Base(bqFile))

			bqOutputFile := filepath.Join(bqOutputPath, fmt.Sprintf("%s_%s", targetLang, filepath.Base(bqFile)))
			if err := ProcessBetterQuestingFile(bqFile, bqOutputFile, targetLang, engine, false, similarityThreshold, batchSize); err != nil {
				fmt.Printf("Warning: Failed to process %s: %v\n", filepath.Base(bqFile), err)
				continue
			}
		}
	}

	if extractOnly {
		fmt.Printf("\nExtraction completed: %d JAR files processed\n", totalExtracted)
		fmt.Printf("Files extracted to: %s\n", outputPath)
	} else if resourcePack {
		if err := generateCombinedResourcePack(allTranslatedFiles, outputPath, targetLang); err != nil {
			return fmt.Errorf("failed to generate resource pack: %v", err)
		}
		fmt.Printf("\nResource pack generated: %s\n", outputPath)
	} else {
		if err := saveCombinedTranslatedFiles(allTranslatedFiles, outputPath); err != nil {
			return fmt.Errorf("failed to save translated files: %v", err)
		}
		fmt.Printf("\nTranslated files saved to: %s\n", outputPath)
	}

	return nil
}

func getModNameFromJAR(jarPath string) string {
	basename := filepath.Base(jarPath)
	// Remove .jar extension and clean up version numbers
	modName := strings.TrimSuffix(basename, ".jar")

	// Try to remove common version patterns
	// e.g., "modname-1.2.3" -> "modname"
	if idx := strings.LastIndex(modName, "-"); idx > 0 {
		// Check if what follows the dash looks like a version
		potential := modName[idx+1:]
		if isVersionLike(potential) {
			modName = modName[:idx]
		}
	}

	return modName
}

func isVersionLike(s string) bool {
	// Simple check for version-like strings (contains numbers and dots/dashes)
	hasNumber := false
	for _, r := range s {
		if r >= '0' && r <= '9' {
			hasNumber = true
		} else if r != '.' && r != '-' && r != '_' && r != '+' && !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') {
			return false
		}
	}
	return hasNumber
}

func processSingleJARForExtraction(jarPath, outputPath string) error {
	langFiles, err := ExtractLanguageFiles(jarPath)
	if err != nil {
		return err
	}

	if len(langFiles) == 0 {
		fmt.Printf("  No language files found in %s\n", filepath.Base(jarPath))
		return nil
	}

	fmt.Printf("  Found %d language files\n", len(langFiles))
	return extractLanguageFilesToDirectory(langFiles, outputPath)
}

func processSingleJARForTranslation(jarPath, targetLang, engine string, similarityThreshold float64, batchSize int) ([]JARLanguageFile, error) {
	langFiles, err := ExtractLanguageFiles(jarPath)
	if err != nil {
		return nil, err
	}

	if len(langFiles) == 0 {
		fmt.Printf("  No language files found in %s\n", filepath.Base(jarPath))
		return nil, nil
	}

	// Find source language files
	var sourceFiles []JARLanguageFile
	for _, lf := range langFiles {
		if isSourceLanguage(lf.Language) {
			sourceFiles = append(sourceFiles, lf)
		}
	}

	if len(sourceFiles) == 0 {
		fmt.Printf("  No source language files found in %s\n", filepath.Base(jarPath))
		return nil, nil
	}

	// Create translator
	translator, err := translators.CreateTranslator(engine)
	if err != nil {
		return nil, err
	}

	var translatedFiles []JARLanguageFile
	for _, sourceFile := range sourceFiles {
		fmt.Printf("  Translating %s (%d keys)\n", sourceFile.Language, len(sourceFile.Data))

		translatedData, err := translators.TranslateDataWithSimilarity(sourceFile.Data, translator, targetLang, similarityThreshold, batchSize)
		if err != nil {
			return nil, err
		}

		translatedFile := JARLanguageFile{
			Path:     strings.Replace(sourceFile.Path, sourceFile.Language, targetLang, 1),
			Language: targetLang,
			Data:     translatedData,
			Format:   sourceFile.Format,
		}
		translatedFiles = append(translatedFiles, translatedFile)
	}

	return translatedFiles, nil
}

func generateCombinedResourcePack(allTranslatedFiles []JARLanguageFile, outputPath, targetLang string) error {
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return err
	}

	// Create pack.mcmeta
	packMcmeta := fmt.Sprintf(`{
  "pack": {
    "pack_format": 15,
    "description": "Combined translated language pack for %s"
  }
}`, targetLang)

	if err := os.WriteFile(filepath.Join(outputPath, "pack.mcmeta"), []byte(packMcmeta), 0644); err != nil {
		return err
	}

	// Save all translated files
	for _, tf := range allTranslatedFiles {
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

		if err := os.MkdirAll(filepath.Dir(resourcePackPath), 0755); err != nil {
			return err
		}

		if err := parsers.WriteFile(resourcePackPath, tf.Data, tf.Format); err != nil {
			return err
		}
	}

	return nil
}

func saveCombinedTranslatedFiles(allTranslatedFiles []JARLanguageFile, outputPath string) error {
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return err
	}

	// Group files by mod (namespace)
	modFiles := make(map[string][]JARLanguageFile)

	for _, tf := range allTranslatedFiles {
		parts := strings.Split(tf.Path, "/")
		if len(parts) < 3 {
			continue
		}

		// Find namespace
		namespace := "unknown"
		for i, part := range parts {
			if part == "assets" && i+1 < len(parts) {
				namespace = parts[i+1]
				break
			}
		}

		modFiles[namespace] = append(modFiles[namespace], tf)
	}

	// Save files grouped by mod
	for namespace, files := range modFiles {
		modDir := filepath.Join(outputPath, namespace)
		if err := os.MkdirAll(modDir, 0755); err != nil {
			return err
		}

		for _, tf := range files {
			filename := fmt.Sprintf("%s%s", tf.Language, parsers.GetExtensionForFormat(tf.Format))
			outputFile := filepath.Join(modDir, filename)

			if err := parsers.WriteFile(outputFile, tf.Data, tf.Format); err != nil {
				return err
			}
		}
	}

	return nil
}
