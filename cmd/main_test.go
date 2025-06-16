package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMainDryRun tests the main application with dry run mode
func TestMainDryRun(t *testing.T) {
	// Build the application first
	cmd := exec.Command("go", "build", "-o", "localizer_test", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("localizer_test")

	tests := []struct {
		name      string
		inputFile string
		args      []string
		wantError bool
	}{
		{
			name:      "JSON file dry run",
			inputFile: "../testdata/test_en_us.json",
			args:      []string{"-input", "../testdata/test_en_us.json", "-dry-run"},
			wantError: false,
		},
		{
			name:      "LANG file dry run",
			inputFile: "../testdata/test_en_us.lang",
			args:      []string{"-input", "../testdata/test_en_us.lang", "-dry-run"},
			wantError: false,
		},
		{
			name:      "JAR file dry run",
			inputFile: "../testdata/test_mod.jar",
			args:      []string{"-input", "../testdata/test_mod.jar", "-dry-run"},
			wantError: false,
		},
		{
			name:      "Instance dry run",
			inputFile: "../testdata/minecraft_instance_example",
			args:      []string{"-input", "../testdata/minecraft_instance_example", "-dry-run"},
			wantError: false,
		},
		{
			name:      "Non-existent file",
			inputFile: "non_existent.json",
			args:      []string{"-input", "non_existent.json", "-dry-run"},
			wantError: true,
		},
		{
			name:      "Help flag",
			inputFile: "",
			args:      []string{"-help"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip test if input file doesn't exist (except for help and non-existent file tests)
			if tt.inputFile != "" && tt.inputFile != "non_existent.json" {
				if _, err := os.Stat(tt.inputFile); os.IsNotExist(err) {
					t.Skipf("Test file does not exist: %s", tt.inputFile)
				}
			}

			cmd := exec.Command("./localizer_test", tt.args...)
			output, err := cmd.CombinedOutput()

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but command succeeded. Output: %s", string(output))
				}
			} else {
				if err != nil {
					t.Errorf("Command failed: %v. Output: %s", err, string(output))
				}
			}

			// Check that output contains expected strings
			outputStr := string(output)
			if !tt.wantError && tt.inputFile != "" {
				if !strings.Contains(outputStr, "MinecraftModsLocalizer CLI") {
					t.Errorf("Output should contain 'MinecraftModsLocalizer CLI'")
				}
				if !strings.Contains(outputStr, "Input type:") && tt.inputFile != "non_existent.json" {
					t.Errorf("Output should contain 'Input type:'")
				}
			}

			if tt.name == "Help flag" {
				if !strings.Contains(outputStr, "Usage:") {
					t.Errorf("Help output should contain 'Usage:'")
				}
			}
		})
	}
}

func TestMainFlags(t *testing.T) {
	// Build the application first
	cmd := exec.Command("go", "build", "-o", "localizer_test", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("localizer_test")

	testFile := "../testdata/test_en_us.json"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file does not exist")
	}

	tests := []struct {
		name           string
		args           []string
		expectInOutput []string
	}{
		{
			name:           "Custom target language",
			args:           []string{"-input", testFile, "-lang", "ko_kr", "-dry-run"},
			expectInOutput: []string{"Target Language: ko"},
		},
		{
			name:           "Custom engine",
			args:           []string{"-input", testFile, "-engine", "google", "-dry-run"},
			expectInOutput: []string{"Engine: google"},
		},
		{
			name:           "Custom batch size",
			args:           []string{"-input", testFile, "-batch-size", "5", "-dry-run"},
			expectInOutput: []string{"MinecraftModsLocalizer CLI"},
		},
		{
			name:           "Custom similarity threshold",
			args:           []string{"-input", testFile, "-similarity", "0.8", "-dry-run"},
			expectInOutput: []string{"MinecraftModsLocalizer CLI"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./localizer_test", tt.args...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Command failed: %v. Output: %s", err, string(output))
				return
			}

			outputStr := string(output)
			for _, expected := range tt.expectInOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Output should contain '%s'. Got: %s", expected, outputStr)
				}
			}
		})
	}
}

// TestApplicationIntegration tests full integration with real files
func TestApplicationIntegration(t *testing.T) {
	// Build the application first
	cmd := exec.Command("go", "build", "-o", "localizer_test", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("localizer_test")

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test JSON file
	testData := map[string]string{
		"item.test.sword":  "Test Sword",
		"block.test.stone": "Test Stone",
	}

	testFile := filepath.Join(tmpDir, "test_input.json")
	jsonData, _ := json.Marshal(testData)
	if err := os.WriteFile(testFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test dry run
	t.Run("Dry run integration", func(t *testing.T) {
		cmd := exec.Command("./localizer_test", "-input", testFile, "-dry-run")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Errorf("Dry run failed: %v. Output: %s", err, string(output))
			return
		}

		outputStr := string(output)
		expectedStrings := []string{
			"MinecraftModsLocalizer CLI",
			"Input type: Language File",
			"Detected format: JSON",
			"Found 2 translation keys",
			"Dry run mode",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(outputStr, expected) {
				t.Errorf("Output should contain '%s'. Got: %s", expected, outputStr)
			}
		}
	})
}

