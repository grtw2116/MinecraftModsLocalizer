package parsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BetterQuesting file structure
type BetterQuestingFile struct {
	Format        string                     `json:"format"`
	QuestDatabase map[string]*Quest          `json:"questDatabase,omitempty"`
	QuestLines    map[string]*QuestLine      `json:"questLines,omitempty"`
}

type Quest struct {
	QuestID         int                    `json:"questID"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	IsMain          bool                   `json:"isMain"`
	IsSilent        bool                   `json:"isSilent"`
	LockedProgress  bool                   `json:"lockedProgress"`
	AutoClaim       bool                   `json:"autoClaim"`
	RepeatTime      int                    `json:"repeatTime"`
	Logic           string                 `json:"logic"`
	TaskLogic       string                 `json:"taskLogic"`
	PreRequisites   []int                  `json:"preRequisites,omitempty"`
	Icon            *ItemStack             `json:"icon,omitempty"`
	Tasks           map[string]*Task       `json:"tasks,omitempty"`
	Rewards         map[string]*Reward     `json:"rewards,omitempty"`
}

type Task struct {
	TaskID        string      `json:"taskID"`
	Index         int         `json:"index"`
	Name          string      `json:"name,omitempty"`
	Description   string      `json:"description,omitempty"`
	RequiredItems []*ItemStack `json:"requiredItems,omitempty"`
}

type Reward struct {
	RewardID string      `json:"rewardID"`
	Index    int         `json:"index"`
	Rewards  []*ItemStack `json:"rewards,omitempty"`
}

type QuestLine struct {
	LineID      int                    `json:"lineID"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Icon        *ItemStack             `json:"icon,omitempty"`
	Quests      map[string]*QuestPos   `json:"quests,omitempty"`
}

type QuestPos struct {
	ID int `json:"id"`
	X  int `json:"x"`
	Y  int `json:"y"`
}

type ItemStack struct {
	ID     string                 `json:"id"`
	Count  int                    `json:"Count"`
	Tag    map[string]interface{} `json:"tag"`
	Damage int                    `json:"Damage"`
}

func IsBetterQuestingFile(filename string) bool {
	// Check if it's a JSON file that might be BetterQuesting
	if !strings.HasSuffix(strings.ToLower(filename), ".json") {
		return false
	}
	
	// Check for common BetterQuesting file names
	baseName := strings.ToLower(filepath.Base(filename))
	bqNames := []string{
		"defaultquests.json",
		"quests.json",
		"betterquesting.json",
		"questbook.json",
	}
	
	for _, name := range bqNames {
		if baseName == name {
			return true
		}
	}
	
	// Try to parse and check for BetterQuesting format
	if data, err := os.ReadFile(filename); err == nil {
		var bqFile BetterQuestingFile
		if err := json.Unmarshal(data, &bqFile); err == nil {
			// Check if it has BetterQuesting format marker
			return strings.Contains(bqFile.Format, "bq_standard") || 
				   bqFile.QuestDatabase != nil || 
				   bqFile.QuestLines != nil
		}
	}
	
	return false
}

func ParseBetterQuestingFile(filename string) (*BetterQuestingFile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var bqFile BetterQuestingFile
	if err := json.Unmarshal(data, &bqFile); err != nil {
		return nil, fmt.Errorf("failed to parse BetterQuesting file: %v", err)
	}

	return &bqFile, nil
}

func ExtractBetterQuestingTranslations(bqFile *BetterQuestingFile) TranslationData {
	translations := make(TranslationData)
	
	// Extract quest translations
	if bqFile.QuestDatabase != nil {
		for questID, quest := range bqFile.QuestDatabase {
			if quest.Name != "" {
				key := fmt.Sprintf("quest.%s.name", questID)
				translations[key] = quest.Name
			}
			if quest.Description != "" {
				key := fmt.Sprintf("quest.%s.description", questID)
				translations[key] = quest.Description
			}
			
			// Extract task translations
			if quest.Tasks != nil {
				for taskID, task := range quest.Tasks {
					if task.Name != "" {
						key := fmt.Sprintf("quest.%s.task.%s.name", questID, taskID)
						translations[key] = task.Name
					}
					if task.Description != "" {
						key := fmt.Sprintf("quest.%s.task.%s.description", questID, taskID)
						translations[key] = task.Description
					}
				}
			}
		}
	}
	
	// Extract quest line translations
	if bqFile.QuestLines != nil {
		for lineID, questLine := range bqFile.QuestLines {
			if questLine.Name != "" {
				key := fmt.Sprintf("questline.%s.name", lineID)
				translations[key] = questLine.Name
			}
			if questLine.Description != "" {
				key := fmt.Sprintf("questline.%s.description", lineID)
				translations[key] = questLine.Description
			}
		}
	}
	
	return translations
}

func ApplyBetterQuestingTranslations(bqFile *BetterQuestingFile, translations TranslationData) {
	// Apply quest translations
	if bqFile.QuestDatabase != nil {
		for questID, quest := range bqFile.QuestDatabase {
			nameKey := fmt.Sprintf("quest.%s.name", questID)
			if translated, exists := translations[nameKey]; exists {
				quest.Name = translated
			}
			
			descKey := fmt.Sprintf("quest.%s.description", questID)
			if translated, exists := translations[descKey]; exists {
				quest.Description = translated
			}
			
			// Apply task translations
			if quest.Tasks != nil {
				for taskID, task := range quest.Tasks {
					taskNameKey := fmt.Sprintf("quest.%s.task.%s.name", questID, taskID)
					if translated, exists := translations[taskNameKey]; exists {
						task.Name = translated
					}
					
					taskDescKey := fmt.Sprintf("quest.%s.task.%s.description", questID, taskID)
					if translated, exists := translations[taskDescKey]; exists {
						task.Description = translated
					}
				}
			}
		}
	}
	
	// Apply quest line translations
	if bqFile.QuestLines != nil {
		for lineID, questLine := range bqFile.QuestLines {
			nameKey := fmt.Sprintf("questline.%s.name", lineID)
			if translated, exists := translations[nameKey]; exists {
				questLine.Name = translated
			}
			
			descKey := fmt.Sprintf("questline.%s.description", lineID)
			if translated, exists := translations[descKey]; exists {
				questLine.Description = translated
			}
		}
	}
}

func WriteBetterQuestingFile(filename string, bqFile *BetterQuestingFile) error {
	data, err := json.MarshalIndent(bqFile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func FindBetterQuestingFiles(instancePath string) ([]string, error) {
	var bqFiles []string
	
	// Common locations for BetterQuesting files in Minecraft instances
	searchPaths := []string{
		filepath.Join(instancePath, "config", "betterquesting"),
		filepath.Join(instancePath, "config"),
		filepath.Join(instancePath, "saves"),
		instancePath, // Root directory
	}
	
	for _, searchPath := range searchPaths {
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue
		}
		
		err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			
			if !info.IsDir() && IsBetterQuestingFile(path) {
				bqFiles = append(bqFiles, path)
			}
			
			return nil
		})
		
		if err != nil {
			return nil, err
		}
	}
	
	return bqFiles, nil
}