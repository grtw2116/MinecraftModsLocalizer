package processors

import (
	"os"

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
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

// InputProcessor defines the interface for processing different input types
type InputProcessor interface {
	Process(inputPath, outputPath, targetLang, engine string, dryRun, extractOnly, resourcePack bool, similarityThreshold float64, batchSize int) error
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

// CreateProcessor creates the appropriate processor for the given input type
func CreateProcessor(inputType InputType) InputProcessor {
	switch inputType {
	case InputTypeLanguageFile:
		return &LanguageFileProcessor{}
	case InputTypeJARFile:
		return &JARFileProcessor{}
	case InputTypeBetterQuesting:
		return &BetterQuestingProcessor{}
	case InputTypeMinecraftInstance:
		return &MinecraftInstanceProcessor{}
	default:
		return nil
	}
}