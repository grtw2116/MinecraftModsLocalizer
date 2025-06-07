package parsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"regexp"
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
	Properties      *QuestProperties       `json:"properties,omitempty"`
}

// QuestProperties represents the properties section of a quest
// This handles both standard format and NBT format with dynamic typed keys
type QuestProperties map[string]interface{}

// GetBetterQuestingData extracts BetterQuesting data from properties, handling both formats
func (qp QuestProperties) GetBetterQuestingData() map[string]interface{} {
	if qp == nil {
		return nil
	}
	
	// Check for standard format first
	if bqData, exists := qp["betterquesting"]; exists {
		if bqMap, ok := bqData.(map[string]interface{}); ok {
			return bqMap
		}
	}
	
	// Check for NBT format (betterquesting:XX)
	for key, value := range qp {
		if matched, _ := regexp.MatchString(`^betterquesting:\d+$`, key); matched {
			if bqMap, ok := value.(map[string]interface{}); ok {
				return bqMap
			}
		}
	}
	
	return nil
}

// SetBetterQuestingData sets BetterQuesting data in properties, preserving the format
func (qp QuestProperties) SetBetterQuestingData(data map[string]interface{}) {
	if qp == nil {
		return
	}
	
	// Check for standard format first
	if _, exists := qp["betterquesting"]; exists {
		qp["betterquesting"] = data
		return
	}
	
	// Check for NBT format (betterquesting:XX)
	for key := range qp {
		if matched, _ := regexp.MatchString(`^betterquesting:\d+$`, key); matched {
			qp[key] = data
			return
		}
	}
	
	// Default to standard format if no existing format found
	qp["betterquesting"] = data
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
		
		// Check for NBT-style format (e.g., "format:8")
		var rawData map[string]interface{}
		if err := json.Unmarshal(data, &rawData); err == nil {
			// Look for NBT-style keys like "format:8", "questDatabase:9"
			for key := range rawData {
				if matched, _ := regexp.MatchString(`^(format|questDatabase|questLines):\d+$`, key); matched {
					return true
				}
			}
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
			
			// Extract from properties if available
			if quest.Properties != nil {
				bqData := quest.Properties.GetBetterQuestingData()
				if bqData != nil {
					// Check for standard format name/desc
					if name, exists := bqData["name"]; exists {
						if nameStr, ok := name.(string); ok && nameStr != "" {
							key := fmt.Sprintf("quest.%s.name", questID)
							translations[key] = nameStr
						}
					}
					if desc, exists := bqData["desc"]; exists {
						if descStr, ok := desc.(string); ok && descStr != "" {
							key := fmt.Sprintf("quest.%s.description", questID)
							translations[key] = descStr
						}
					}
					
					// Check for NBT format name/desc
					if name, exists := bqData["name:8"]; exists {
						if nameStr, ok := name.(string); ok && nameStr != "" {
							key := fmt.Sprintf("quest.%s.name", questID)
							translations[key] = nameStr
						}
					}
					if desc, exists := bqData["desc:8"]; exists {
						if descStr, ok := desc.(string); ok && descStr != "" {
							key := fmt.Sprintf("quest.%s.description", questID)
							translations[key] = descStr
						}
					}
				}
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
				
				// Also apply to properties if they exist
				if quest.Properties != nil {
					bqData := quest.Properties.GetBetterQuestingData()
					if bqData != nil {
						// Update both standard and NBT format fields
						if _, exists := bqData["name"]; exists {
							bqData["name"] = translated
						}
						if _, exists := bqData["name:8"]; exists {
							bqData["name:8"] = translated
						}
						quest.Properties.SetBetterQuestingData(bqData)
					}
				}
			}
			
			descKey := fmt.Sprintf("quest.%s.description", questID)
			if translated, exists := translations[descKey]; exists {
				quest.Description = translated
				
				// Also apply to properties if they exist
				if quest.Properties != nil {
					bqData := quest.Properties.GetBetterQuestingData()
					if bqData != nil {
						// Update both standard and NBT format fields
						if _, exists := bqData["desc"]; exists {
							bqData["desc"] = translated
						}
						if _, exists := bqData["desc:8"]; exists {
							bqData["desc:8"] = translated
						}
						quest.Properties.SetBetterQuestingData(bqData)
					}
				}
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

// ExtractNBTBetterQuestingTranslations extracts translations directly from NBT-style BetterQuesting files
func ExtractNBTBetterQuestingTranslations(filename string) (TranslationData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse NBT BetterQuesting file: %v", err)
	}

	translations := make(TranslationData)
	
	// Find questDatabase
	var questDatabase map[string]interface{}
	for key, value := range rawData {
		if matched, _ := regexp.MatchString(`^questDatabase:\d+$`, key); matched {
			if qdb, ok := value.(map[string]interface{}); ok {
				questDatabase = qdb
				break
			}
		}
	}
	
	if questDatabase != nil {
		for questID, questData := range questDatabase {
			if questMap, ok := questData.(map[string]interface{}); ok {
				// Look for properties section
				var properties map[string]interface{}
				for key, value := range questMap {
					if matched, _ := regexp.MatchString(`^properties:\d+$`, key); matched {
						if props, ok := value.(map[string]interface{}); ok {
							properties = props
							break
						}
					}
				}
				
				if properties != nil {
					// Look for betterquesting section
					var bqSection map[string]interface{}
					for key, value := range properties {
						if matched, _ := regexp.MatchString(`^betterquesting:\d+$`, key); matched {
							if bqs, ok := value.(map[string]interface{}); ok {
								bqSection = bqs
								break
							}
						}
					}
					
					if bqSection != nil {
						// Extract name and description
						if name, exists := bqSection["name:8"]; exists {
							if nameStr, ok := name.(string); ok && nameStr != "" {
								key := fmt.Sprintf("quest.%s.name", questID)
								translations[key] = nameStr
							}
						}
						if desc, exists := bqSection["desc:8"]; exists {
							if descStr, ok := desc.(string); ok && descStr != "" {
								key := fmt.Sprintf("quest.%s.description", questID)
								translations[key] = descStr
							}
						}
					}
				}
			}
		}
	}
	
	return translations, nil
}

// ApplyNBTBetterQuestingTranslations applies translations to NBT-style BetterQuesting files
func ApplyNBTBetterQuestingTranslations(filename string, translations TranslationData) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("failed to parse NBT BetterQuesting file: %v", err)
	}
	
	// Find questDatabase
	var questDatabase map[string]interface{}
	var questDatabaseKey string
	for key, value := range rawData {
		if matched, _ := regexp.MatchString(`^questDatabase:\d+$`, key); matched {
			if qdb, ok := value.(map[string]interface{}); ok {
				questDatabase = qdb
				questDatabaseKey = key
				break
			}
		}
	}
	
	if questDatabase != nil {
		for questID, questData := range questDatabase {
			if questMap, ok := questData.(map[string]interface{}); ok {
				// Look for properties section
				var properties map[string]interface{}
				var propertiesKey string
				for key, value := range questMap {
					if matched, _ := regexp.MatchString(`^properties:\d+$`, key); matched {
						if props, ok := value.(map[string]interface{}); ok {
							properties = props
							propertiesKey = key
							break
						}
					}
				}
				
				if properties != nil {
					// Look for betterquesting section
					var bqSection map[string]interface{}
					var bqSectionKey string
					for key, value := range properties {
						if matched, _ := regexp.MatchString(`^betterquesting:\d+$`, key); matched {
							if bqs, ok := value.(map[string]interface{}); ok {
								bqSection = bqs
								bqSectionKey = key
								break
							}
						}
					}
					
					if bqSection != nil {
						// Apply translations
						nameKey := fmt.Sprintf("quest.%s.name", questID)
						if translated, exists := translations[nameKey]; exists {
							bqSection["name:8"] = translated
						}
						
						descKey := fmt.Sprintf("quest.%s.description", questID)
						if translated, exists := translations[descKey]; exists {
							bqSection["desc:8"] = translated
						}
						
						// Update the nested structure
						properties[bqSectionKey] = bqSection
						questMap[propertiesKey] = properties
						questDatabase[questID] = questMap
					}
				}
			}
		}
		rawData[questDatabaseKey] = questDatabase
	}
	
	// Write back to file
	outputData, err := json.MarshalIndent(rawData, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, outputData, 0644)
}