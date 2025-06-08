package processors

import (
	"fmt"
)

// MinecraftInstanceProcessor handles Minecraft instance directories
type MinecraftInstanceProcessor struct{}

func (p *MinecraftInstanceProcessor) Process(inputPath, outputPath, targetLang, engine string, dryRun, extractOnly, resourcePack bool, similarityThreshold float64, batchSize int) error {
	fmt.Printf("Processing Minecraft instance: %s\n", inputPath)
	
	// Delegate to existing ProcessMinecraftInstance function
	return ProcessMinecraftInstance(inputPath, outputPath, targetLang, engine, dryRun, extractOnly, resourcePack, similarityThreshold, batchSize)
}