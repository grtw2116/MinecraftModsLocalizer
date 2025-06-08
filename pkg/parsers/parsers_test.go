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