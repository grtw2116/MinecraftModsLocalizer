package processors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessLanguageFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "processor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test data
	testData := map[string]string{
		"item.test.example": "Example Item",
		"block.test.stone":  "Test Stone",
		"gui.test.title":    "Test GUI",
	}

	// Create test JSON file
	jsonFile := filepath.Join(tmpDir, "test_en_us.json")
	jsonData, _ := json.Marshal(testData)
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to create test JSON file: %v", err)
	}

	// Create test LANG file
	langFile := filepath.Join(tmpDir, "test_en_us.lang")
	var langContent strings.Builder
	for k, v := range testData {
		langContent.WriteString(k + "=" + v + "\n")
	}
	if err := os.WriteFile(langFile, []byte(langContent.String()), 0644); err != nil {
		t.Fatalf("Failed to create test LANG file: %v", err)
	}

	tests := []struct {
		name      string
		inputFile string
		dryRun    bool
		wantError bool
	}{
		{
			name:      "JSON file dry run",
			inputFile: jsonFile,
			dryRun:    true,
			wantError: false,
		},
		{
			name:      "LANG file dry run",
			inputFile: langFile,
			dryRun:    true,
			wantError: false,
		},
		{
			name:      "Non-existent file",
			inputFile: filepath.Join(tmpDir, "non_existent.json"),
			dryRun:    true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputFile := tt.inputFile + "_output"
			defer os.Remove(outputFile)

			err := ProcessLanguageFile(
				tt.inputFile,
				outputFile,
				"ja",
				"openai",
				"1.20",
				tt.dryRun,
				0.6,
				1,
			)

			if tt.wantError {
				if err == nil {
					t.Errorf("ProcessLanguageFile() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ProcessLanguageFile() unexpected error: %v", err)
				}
			}

			// For dry run, output file should not be created
			if tt.dryRun && !tt.wantError {
				if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
					t.Errorf("ProcessLanguageFile() with dry run should not create output file")
				}
			}
		})
	}
}

func TestProcessLanguageFileWithRealData(t *testing.T) {
	testFile := "../../testdata/test_en_us.json"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test data file does not exist")
	}

	tmpDir, err := os.MkdirTemp("", "processor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "output.json")

	// Test dry run - should not create output file and should not error
	err = ProcessLanguageFile(
		testFile,
		outputFile,
		"ja",
		"openai",
		"1.20",
		true, // dry run
		0.6,
		1,
	)

	if err != nil {
		t.Errorf("ProcessLanguageFile() dry run failed: %v", err)
	}

	// Output file should not exist in dry run
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Errorf("ProcessLanguageFile() dry run should not create output file")
	}
}
