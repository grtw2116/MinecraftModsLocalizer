package processors

import (
	"fmt"
)

// JARFileProcessor handles JAR files
type JARFileProcessor struct{}

func (p *JARFileProcessor) Process(inputPath, outputPath, targetLang, engine string, dryRun, extractOnly, resourcePack bool, similarityThreshold float64, batchSize int) error {
	fmt.Printf("Processing JAR file: %s\n", inputPath)
	
	// Delegate to existing ProcessJARFile function
	return ProcessJARFile(inputPath, outputPath, targetLang, engine, dryRun, extractOnly, resourcePack, similarityThreshold, batchSize)
}