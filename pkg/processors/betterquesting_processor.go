package processors

import (
	"fmt"
)

// BetterQuestingProcessor handles BetterQuesting files
type BetterQuestingProcessor struct{}

func (p *BetterQuestingProcessor) Process(inputPath, outputPath, targetLang, engine string, dryRun, extractOnly, resourcePack bool, similarityThreshold float64, batchSize int) error {
	fmt.Printf("Processing BetterQuesting file: %s\n", inputPath)
	
	// Delegate to existing ProcessBetterQuestingFile function
	return ProcessBetterQuestingFile(inputPath, outputPath, targetLang, engine, dryRun, similarityThreshold, batchSize)
}