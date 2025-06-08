package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/processors"
)

var (
	inputFile           = flag.String("input", "", "Input path (Minecraft instance directory, JAR file, or individual language file)")
	outputFile          = flag.String("output", "", "Output file path (optional, defaults to input_translated.ext or resource pack)")
	targetLang          = flag.String("lang", "ja", "Target language code (default: ja)")
	engine              = flag.String("engine", "openai", "Translation engine: openai, google, deepl (default: openai)")
	dryRun              = flag.Bool("dry-run", false, "Parse file and show statistics without translating")
	similarityThreshold = flag.Float64("similarity", 0.6, "Similarity threshold for finding similar examples (0.0-1.0, default: 0.6)")
	extractOnly         = flag.Bool("extract-only", false, "Extract language files from JAR without translating")
	batchSize           = flag.Int("batch-size", 1, "Number of texts to translate per API request (default: 1 for individual processing, 10+ for batch processing)")
	help                = flag.Bool("help", false, "Show help")
)

func main() {
	flag.Parse()

	if *help || *inputFile == "" {
		showUsage()
		return
	}

	fmt.Printf("MinecraftModsLocalizer CLI\n")
	fmt.Printf("Input: %s\n", *inputFile)

	// Detect input type
	inputType := processors.DetectInputType(*inputFile)
	if inputType == processors.InputTypeUnknown {
		fmt.Fprintf(os.Stderr, "Error: Unable to determine input type for: %s\n", *inputFile)
		fmt.Fprintf(os.Stderr, "Supported inputs: Minecraft instance directories, .jar files, .json/.lang/.snbt files, BetterQuesting files\n")
		os.Exit(1)
	}

	fmt.Printf("Input type: %s\n", inputType.String())
	fmt.Printf("Target Language: %s\n", *targetLang)
	fmt.Printf("Engine: %s\n", *engine)

	fmt.Printf("Output: %s\n", *outputFile)

	// Process the input
	if err := processors.ProcessInput(inputType, *inputFile, *outputFile, *targetLang, *engine, *dryRun, *extractOnly, *similarityThreshold, *batchSize); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", inputType.String(), err)
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "MinecraftModsLocalizer - Translate Minecraft mod files\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nSupported inputs: Minecraft instance directories, .jar files, .json/.lang/.snbt files, BetterQuesting files\n")
	fmt.Fprintf(os.Stderr, "Supported languages: ja, ko, zh-cn, zh-tw, fr, de, es, etc.\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  %s -input /path/to/minecraft/instance -lang ja\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input /path/to/minecraft/instance -extract-only\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input mod.jar -lang ja\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input defaultquests.json -lang ja\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input en_us.json -lang ja -engine openai -similarity 0.7\n", os.Args[0])
}
