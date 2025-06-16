package parsers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DetectFileFormat(filename string) FileFormat {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return FormatJSON
	case ".lang":
		return FormatLang
	case ".snbt":
		return FormatSNBT
	default:
		return FormatUnknown
	}
}

func ParseFile(filename string) (TranslationData, FileFormat, error) {
	format := DetectFileFormat(filename)

	switch format {
	case FormatJSON:
		return parseJSONFile(filename)
	case FormatLang:
		return parseLangFile(filename)
	case FormatSNBT:
		return parseSNBTFile(filename)
	default:
		return nil, FormatUnknown, fmt.Errorf("unsupported file format: %s", filepath.Ext(filename))
	}
}

func parseJSONFile(filename string) (TranslationData, FileFormat, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, FormatJSON, err
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, FormatJSON, err
	}

	result := make(TranslationData)
	flattenJSON(data, "", result)

	return result, FormatJSON, nil
}

func flattenJSON(data interface{}, prefix string, result TranslationData) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}
			flattenJSON(value, fullKey, result)
		}
	case string:
		result[prefix] = v
	default:
		result[prefix] = fmt.Sprintf("%v", v)
	}
}

func parseLangFile(filename string) (TranslationData, FileFormat, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, FormatLang, err
	}
	defer file.Close()

	result := make(TranslationData)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	return result, FormatLang, scanner.Err()
}

func parseSNBTFile(filename string) (TranslationData, FileFormat, error) {
	// TODO: Implement SNBT parsing
	return nil, FormatSNBT, fmt.Errorf("SNBT format not yet implemented")
}

func WriteFile(filename string, data TranslationData, format FileFormat) error {
	switch format {
	case FormatJSON:
		return writeJSONFile(filename, data)
	case FormatLang:
		return writeLangFile(filename, data)
	case FormatSNBT:
		return writeSNBTFile(filename, data)
	default:
		return fmt.Errorf("unsupported output format")
	}
}

func writeJSONFile(filename string, data TranslationData) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func writeLangFile(filename string, data TranslationData) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for key, value := range data {
		if _, err := fmt.Fprintf(file, "%s=%s\n", key, value); err != nil {
			return err
		}
	}

	return nil
}

func writeSNBTFile(filename string, data TranslationData) error {
	// TODO: Implement SNBT writing
	return fmt.Errorf("SNBT format not yet implemented")
}
