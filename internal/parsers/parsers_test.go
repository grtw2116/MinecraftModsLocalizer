package parsers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectFileFormat(t *testing.T) {
	tests := []struct {
		filename string
		expected FileFormat
	}{
		{"test.json", FormatJSON},
		{"test.JSON", FormatJSON},
		{"test.lang", FormatLang},
		{"test.LANG", FormatLang},
		{"test.snbt", FormatSNBT},
		{"test.SNBT", FormatSNBT},
		{"test.txt", FormatUnknown},
		{"", FormatUnknown},
		{"no-extension", FormatUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := DetectFileFormat(tt.filename)
			if result != tt.expected {
				t.Errorf("DetectFileFormat(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestGetExtensionForFormat(t *testing.T) {
	tests := []struct {
		format   FileFormat
		expected string
	}{
		{FormatJSON, ".json"},
		{FormatLang, ".lang"},
		{FormatSNBT, ".snbt"},
		{FormatUnknown, ".txt"},
	}

	for _, tt := range tests {
		t.Run(tt.format.String(), func(t *testing.T) {
			result := GetExtensionForFormat(tt.format)
			if result != tt.expected {
				t.Errorf("GetExtensionForFormat(%v) = %s, want %s", tt.format, result, tt.expected)
			}
		})
	}
}

func TestFormatLocaleCode(t *testing.T) {
	tests := []struct {
		localeCode       string
		minecraftVersion string
		expected         string
	}{
		{"ja_jp", "1.20", "ja_jp"},
		{"JA_JP", "1.20", "ja_jp"},
		{"zh_cn", "1.16", "zh_cn"},
		{"en_us", "1.12", "en_us"},
		{"ja_jp", "1.10.2", "ja_JP"},
		{"JA_JP", "1.10.2", "ja_JP"},
		{"zh_cn", "1.9", "zh_CN"},
		{"en_us", "1.8.9", "en_US"},
		{"ja_jp", "1.7.10", "ja_JP"},
		{"simple", "1.20", "simple"},
		{"simple", "1.10.2", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.localeCode+"_"+tt.minecraftVersion, func(t *testing.T) {
			result := FormatLocaleCode(tt.localeCode, tt.minecraftVersion)
			if result != tt.expected {
				t.Errorf("FormatLocaleCode(%s, %s) = %s, want %s", tt.localeCode, tt.minecraftVersion, result, tt.expected)
			}
		})
	}
}

func TestIsLegacyMinecraftVersion(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		{"1.7", true},
		{"1.7.2", true},
		{"1.7.10", true},
		{"1.8", true},
		{"1.8.8", true},
		{"1.8.9", true},
		{"1.9", true},
		{"1.9.4", true},
		{"1.10", true},
		{"1.10.1", true},
		{"1.10.2", true},
		{"1.11", false},
		{"1.11.1", false},
		{"1.11.2", false},
		{"1.12", false},
		{"1.12.1", false},
		{"1.12.2", false},
		{"1.13", false},
		{"1.16", false},
		{"1.20", false},
		{"1.21", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := IsLegacyMinecraftVersion(tt.version)
			if result != tt.expected {
				t.Errorf("IsLegacyMinecraftVersion(%s) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestFormatModernLocaleCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ja_jp", "ja_jp"},
		{"JA_JP", "ja_jp"},
		{"zh_cn", "zh_cn"},
		{"EN_US", "en_us"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatModernLocaleCode(tt.input)
			if result != tt.expected {
				t.Errorf("FormatModernLocaleCode(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatLegacyLocaleCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ja_jp", "ja_JP"},
		{"JA_JP", "ja_JP"},
		{"zh_cn", "zh_CN"},
		{"EN_US", "en_US"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatLegacyLocaleCode(tt.input)
			if result != tt.expected {
				t.Errorf("FormatLegacyLocaleCode(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFileFormatString(t *testing.T) {
	tests := []struct {
		format   FileFormat
		expected string
	}{
		{FormatJSON, "JSON"},
		{FormatLang, "LANG"},
		{FormatSNBT, "SNBT"},
		{FormatUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.format.String()
			if result != tt.expected {
				t.Errorf("FileFormat(%d).String() = %s, want %s", tt.format, result, tt.expected)
			}
		})
	}
}

func TestParseAndWriteJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parser_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test data
	testData := TranslationData{
		"item.test.example": "Example Item",
		"block.test.stone":  "Test Stone",
		"gui.test.title":    "Test GUI",
	}

	// Create test JSON file
	jsonFile := filepath.Join(tmpDir, "test.json")
	jsonBytes, _ := json.Marshal(testData)
	if err := os.WriteFile(jsonFile, jsonBytes, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	parsed, format, err := ParseFile(jsonFile)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	if format != FormatJSON {
		t.Errorf("ParseFile() format = %v, want %v", format, FormatJSON)
	}

	if len(parsed) != len(testData) {
		t.Errorf("ParseFile() returned %d items, want %d", len(parsed), len(testData))
	}

	for key, value := range testData {
		if parsed[key] != value {
			t.Errorf("ParseFile() key %s = %s, want %s", key, parsed[key], value)
		}
	}

	// Write the data back
	outputFile := filepath.Join(tmpDir, "output.json")
	if err := WriteFile(outputFile, parsed, FormatJSON); err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	// Parse the written file to verify
	reparsed, _, err := ParseFile(outputFile)
	if err != nil {
		t.Fatalf("ParseFile() on written file failed: %v", err)
	}

	if len(reparsed) != len(testData) {
		t.Errorf("Reparsed file has %d items, want %d", len(reparsed), len(testData))
	}

	for key, value := range testData {
		if reparsed[key] != value {
			t.Errorf("Reparsed key %s = %s, want %s", key, reparsed[key], value)
		}
	}
}

func TestParseAndWriteLang(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parser_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test data
	testData := TranslationData{
		"item.test.example": "Example Item",
		"block.test.stone":  "Test Stone",
		"gui.test.title":    "Test GUI",
	}

	// Create test LANG file
	langFile := filepath.Join(tmpDir, "test.lang")
	var langContent strings.Builder
	for k, v := range testData {
		langContent.WriteString(k + "=" + v + "\n")
	}
	if err := os.WriteFile(langFile, []byte(langContent.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	parsed, format, err := ParseFile(langFile)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	if format != FormatLang {
		t.Errorf("ParseFile() format = %v, want %v", format, FormatLang)
	}

	if len(parsed) != len(testData) {
		t.Errorf("ParseFile() returned %d items, want %d", len(parsed), len(testData))
	}

	for key, value := range testData {
		if parsed[key] != value {
			t.Errorf("ParseFile() key %s = %s, want %s", key, parsed[key], value)
		}
	}

	// Write the data back
	outputFile := filepath.Join(tmpDir, "output.lang")
	if err := WriteFile(outputFile, parsed, FormatLang); err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	// Parse the written file to verify
	reparsed, _, err := ParseFile(outputFile)
	if err != nil {
		t.Fatalf("ParseFile() on written file failed: %v", err)
	}

	if len(reparsed) != len(testData) {
		t.Errorf("Reparsed file has %d items, want %d", len(reparsed), len(testData))
	}

	for key, value := range testData {
		if reparsed[key] != value {
			t.Errorf("Reparsed key %s = %s, want %s", key, reparsed[key], value)
		}
	}
}

func TestParseFileWithRealData(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		format   FileFormat
	}{
		{
			name:     "Real JSON file",
			filepath: "../../testdata/test_en_us.json",
			format:   FormatJSON,
		},
		{
			name:     "Real LANG file",
			filepath: "../../testdata/test_en_us.lang",
			format:   FormatLang,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := os.Stat(tt.filepath); os.IsNotExist(err) {
				t.Skipf("Test file does not exist: %s", tt.filepath)
			}

			parsed, format, err := ParseFile(tt.filepath)
			if err != nil {
				t.Fatalf("ParseFile() failed: %v", err)
			}

			if format != tt.format {
				t.Errorf("ParseFile() format = %v, want %v", format, tt.format)
			}

			if len(parsed) == 0 {
				t.Errorf("ParseFile() returned empty data")
			}

			// Test that all keys and values are non-empty strings
			for key, value := range parsed {
				if key == "" {
					t.Errorf("ParseFile() returned empty key")
				}
				if value == "" {
					t.Errorf("ParseFile() returned empty value for key %s", key)
				}
			}
		})
	}
}

func TestParseFileErrors(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "Invalid JSON",
			filename: "invalid.json",
			content:  `{"invalid": json}`,
		},
		{
			name:     "Non-existent file",
			filename: "non_existent.json",
			content:  "", // Will not be created
		},
	}

	tmpDir, err := os.MkdirTemp("", "parser_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filepath := filepath.Join(tmpDir, tt.filename)

			// Create file with content (except for non-existent file test)
			if tt.content != "" {
				if err := os.WriteFile(filepath, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			_, _, err := ParseFile(filepath)
			if err == nil {
				t.Errorf("ParseFile() should have returned an error for %s", tt.name)
			}
		})
	}
}
