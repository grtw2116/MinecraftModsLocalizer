package processors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grtw2116/MinecraftModsLocalizer/internal/parsers"
)

// InputType represents the type of input being processed
type InputType int

const (
	InputTypeUnknown InputType = iota
	InputTypeLanguageFile
	InputTypeJARFile
	InputTypeBetterQuesting
	InputTypeMinecraftInstance
)

func (it InputType) String() string {
	switch it {
	case InputTypeLanguageFile:
		return "Language File"
	case InputTypeJARFile:
		return "JAR File"
	case InputTypeBetterQuesting:
		return "BetterQuesting File"
	case InputTypeMinecraftInstance:
		return "Minecraft Instance"
	default:
		return "Unknown"
	}
}

// DetectInputType determines the type of input based on the path
func DetectInputType(inputPath string) InputType {
	// Check if path exists
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return InputTypeUnknown
	}

	// Check for JAR file first (highest priority)
	if !fileInfo.IsDir() && IsJARFile(inputPath) {
		return InputTypeJARFile
	}

	// Check for BetterQuesting file
	if !fileInfo.IsDir() && parsers.IsBetterQuestingFile(inputPath) {
		return InputTypeBetterQuesting
	}

	// Check for language file
	if !fileInfo.IsDir() && isLanguageFile(inputPath) {
		return InputTypeLanguageFile
	}

	// Check if it's a directory (potential Minecraft instance)
	if fileInfo.IsDir() {
		return InputTypeMinecraftInstance
	}

	return InputTypeUnknown
}

// IsJARFile checks if a file is a JAR file based on its extension
func IsJARFile(filename string) bool {
	return strings.ToLower(filepath.Ext(filename)) == ".jar"
}

// isLanguageFile checks if a file is a language file (.json, .lang, .snbt)
func isLanguageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".json" || ext == ".lang" || ext == ".snbt"
}

// ProcessInput processes the input based on its detected type
func ProcessInput(inputType InputType, inputPath, outputPath, targetLang, engine, minecraftVersion string, dryRun, extractOnly bool, similarityThreshold float64, batchSize int) error {
	switch inputType {
	case InputTypeLanguageFile:
		return ProcessLanguageFile(inputPath, outputPath, targetLang, engine, minecraftVersion, dryRun, similarityThreshold, batchSize)
	case InputTypeJARFile:
		return ProcessJARFile(inputPath, outputPath, targetLang, engine, minecraftVersion, dryRun, extractOnly, false, similarityThreshold, batchSize)
	case InputTypeBetterQuesting:
		return ProcessBetterQuestingFile(inputPath, outputPath, targetLang, engine, minecraftVersion, dryRun, similarityThreshold, batchSize)
	case InputTypeMinecraftInstance:
		return ProcessMinecraftInstance(inputPath, outputPath, targetLang, engine, minecraftVersion, dryRun, extractOnly, true, similarityThreshold, batchSize)
	default:
		return fmt.Errorf("unsupported input type: %s", inputType.String())
	}
}
