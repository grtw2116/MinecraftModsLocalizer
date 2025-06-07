package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/grtw2116/MinecraftModsLocalizer/pkg/parsers"
	"github.com/grtw2116/MinecraftModsLocalizer/pkg/processors"
	"github.com/grtw2116/MinecraftModsLocalizer/pkg/translators"
)

var (
	inputFile           = flag.String("input", "", "Input path (Minecraft instance directory, JAR file, or individual language file)")
	outputFile          = flag.String("output", "", "Output file path (optional, defaults to input_translated.ext or resource pack)")
	targetLang          = flag.String("lang", "ja", "Target language code (default: ja)")
	engine              = flag.String("engine", "openai", "Translation engine: openai, google, deepl (default: openai)")
	dryRun              = flag.Bool("dry-run", false, "Parse file and show statistics without translating")
	similarityThreshold = flag.Float64("similarity", 0.6, "Similarity threshold for finding similar examples (0.0-1.0, default: 0.6)")
	extractOnly         = flag.Bool("extract-only", false, "Extract language files from JAR without translating")
	resourcePack        = flag.Bool("resource-pack", false, "Generate resource pack format output")
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
	fmt.Printf("Target Language: %s\n", *targetLang)
	fmt.Printf("Engine: %s\n", *engine)

	if *outputFile == "" {
		*outputFile = generateOutputPath(*inputFile)
	}
	fmt.Printf("Output: %s\n", *outputFile)

	// Check if input is a JAR file
	if processors.IsJARFile(*inputFile) {
		if err := processors.ProcessJARFile(*inputFile, *outputFile, *targetLang, *engine, *dryRun, *extractOnly, *resourcePack, *similarityThreshold); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing JAR file: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if input is a BetterQuesting file
	if parsers.IsBetterQuestingFile(*inputFile) {
		if err := processors.ProcessBetterQuestingFile(*inputFile, *outputFile, *targetLang, *engine, *dryRun, *similarityThreshold); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing BetterQuesting file: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if input is a directory (potential Minecraft instance)
	if fileInfo, err := os.Stat(*inputFile); err == nil && fileInfo.IsDir() {
		if err := processors.ProcessMinecraftInstance(*inputFile, *outputFile, *targetLang, *engine, *dryRun, *extractOnly, *resourcePack, *similarityThreshold); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing Minecraft instance: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Parse input file
	data, format, err := parsers.ParseFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Detected format: %v\n", format.String())
	fmt.Printf("Found %d translation keys\n", len(data))

	if *dryRun {
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
		return
	}

	// Create translator
	translator, err := translators.CreateTranslator(*engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating translator: %v\n", err)
		os.Exit(1)
	}

	// Perform translation
	fmt.Printf("Starting translation with %s engine (similarity threshold: %.1f)...\n", *engine, *similarityThreshold)
	translatedData, err := translators.TranslateDataWithSimilarity(data, translator, *targetLang, *similarityThreshold)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during translation: %v\n", err)
		os.Exit(1)
	}

	// Write output file
	if err := parsers.WriteFile(*outputFile, translatedData, format); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Processing completed: %s\n", *outputFile)
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "MinecraftModsLocalizer - Translate Minecraft mod files\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nSupported inputs: Minecraft instance directories, .jar files, .json/.lang/.snbt files, BetterQuesting files\n")
	fmt.Fprintf(os.Stderr, "Supported languages: ja, ko, zh-cn, zh-tw, fr, de, es, etc.\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  %s -input /path/to/minecraft/instance -lang ja -resource-pack\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input /path/to/minecraft/instance -extract-only\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input mod.jar -lang ja -resource-pack\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input defaultquests.json -lang ja\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -input en_us.json -lang ja -engine openai -similarity 0.7\n", os.Args[0])
}

func generateOutputPath(input string) string {
	// Extract extension and create output filename
	if len(input) > 4 {
		ext := input[len(input)-5:]
		base := input[:len(input)-5]
		return base + "_translated" + ext
	}
	return input + "_translated"
}

