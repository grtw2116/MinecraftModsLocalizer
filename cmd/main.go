package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
	"github.com/grtw2116/MinecraftModsLocalizer/pkg/processors"
)

var (
	inputFile           = flag.String("input", "", "Input path (Minecraft instance directory, JAR file, or individual language file)")
	outputFile          = flag.String("output", "", "Output file path (optional, defaults to input_translated.ext or resource pack)")
	targetLang          = flag.String("lang", "ja_jp", "Target language code (default: ja_jp)")
	engine              = flag.String("engine", "openai", "Translation engine: openai, google, deepl (default: openai)")
	minecraftVersion    = flag.String("minecraft-version", "1.20", "Minecraft version for locale formatting (e.g., 1.10.2, 1.11, 1.20)")
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
	// Validate target language
	if !parsers.ValidateLanguageCode(*targetLang) {
		fmt.Fprintf(os.Stderr, "Error: Unsupported language code: %s\n", *targetLang)
		fmt.Fprintf(os.Stderr, "Use 'localizer -help' to see supported languages\n")
		os.Exit(1)
	}

	// Format language code for the specified Minecraft version
	formattedLang, err := parsers.FormatLanguageCodeForVersion(*targetLang, *minecraftVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting language code: %v\n", err)
		os.Exit(1)
	}

	langInfo, _ := parsers.GetLanguage(*targetLang)
	fmt.Printf("Target Language: %s (%s)\n", formattedLang, langInfo.English)
	fmt.Printf("Engine: %s\n", *engine)
	fmt.Printf("Minecraft Version: %s\n", *minecraftVersion)

	fmt.Printf("Output: %s\n", *outputFile)

	// Process the input
	if err := processors.ProcessInput(inputType, *inputFile, *outputFile, formattedLang, *engine, *minecraftVersion, *dryRun, *extractOnly, *similarityThreshold, *batchSize); err != nil {
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
	fmt.Fprintf(os.Stderr, "Supported languages: Use any Minecraft language code (e.g., ja_jp, ko_kr, zh_cn, zh_tw, fr_fr, de_de, es_es)\n")
	fmt.Fprintf(os.Stderr, "                     Full list: en_us, ja_jp, zh_cn, ko_kr, ru_ru, es_es, fr_fr, de_de, pt_br, it_it, and 113 more\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  %s -input /path/to/minecraft/instance -lang ja_jp\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input /path/to/minecraft/instance -extract-only\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input mod.jar -lang ja_jp\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input defaultquests.json -lang ja_jp\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input en_us.json -lang ja_jp -engine openai -similarity 0.7\n", os.Args[0])
}
