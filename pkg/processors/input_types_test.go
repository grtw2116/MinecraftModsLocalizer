package processors

import (
	"os"
	"testing"
)

func TestDetectInputType(t *testing.T) {
	tests := []struct {
		name         string
		inputPath    string
		setupFunc    func(string) error // Function to set up test files/dirs
		cleanupFunc  func(string)       // Function to clean up test files/dirs
		expectedType InputType
	}{
		{
			name:      "JAR file",
			inputPath: "test.jar",
			setupFunc: func(path string) error {
				file, err := os.Create(path)
				if err != nil {
					return err
				}
				file.Close()
				return nil
			},
			cleanupFunc: func(path string) {
				os.Remove(path)
			},
			expectedType: InputTypeJARFile,
		},
		{
			name:      "JSON language file",
			inputPath: "en_us.json",
			setupFunc: func(path string) error {
				file, err := os.Create(path)
				if err != nil {
					return err
				}
				file.Close()
				return nil
			},
			cleanupFunc: func(path string) {
				os.Remove(path)
			},
			expectedType: InputTypeLanguageFile,
		},
		{
			name:      "LANG language file",
			inputPath: "en_us.lang",
			setupFunc: func(path string) error {
				file, err := os.Create(path)
				if err != nil {
					return err
				}
				file.Close()
				return nil
			},
			cleanupFunc: func(path string) {
				os.Remove(path)
			},
			expectedType: InputTypeLanguageFile,
		},
		{
			name:      "SNBT language file",
			inputPath: "en_us.snbt",
			setupFunc: func(path string) error {
				file, err := os.Create(path)
				if err != nil {
					return err
				}
				file.Close()
				return nil
			},
			cleanupFunc: func(path string) {
				os.Remove(path)
			},
			expectedType: InputTypeLanguageFile,
		},
		{
			name:      "Directory (Minecraft instance)",
			inputPath: "test_instance",
			setupFunc: func(path string) error {
				return os.Mkdir(path, 0755)
			},
			cleanupFunc: func(path string) {
				os.RemoveAll(path)
			},
			expectedType: InputTypeMinecraftInstance,
		},
		{
			name:         "Non-existent file",
			inputPath:    "non_existent.txt",
			setupFunc:    nil,
			cleanupFunc:  nil,
			expectedType: InputTypeUnknown,
		},
		{
			name:      "Unknown file type",
			inputPath: "test.txt",
			setupFunc: func(path string) error {
				file, err := os.Create(path)
				if err != nil {
					return err
				}
				file.Close()
				return nil
			},
			cleanupFunc: func(path string) {
				os.Remove(path)
			},
			expectedType: InputTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test file/directory
			if tt.setupFunc != nil {
				if err := tt.setupFunc(tt.inputPath); err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			}

			// Cleanup after test
			if tt.cleanupFunc != nil {
				defer tt.cleanupFunc(tt.inputPath)
			}

			// Test input type detection
			result := DetectInputType(tt.inputPath)
			if result != tt.expectedType {
				t.Errorf("DetectInputType(%s) = %v, want %v", tt.inputPath, result, tt.expectedType)
			}
		})
	}
}

func TestDetectInputTypeWithRealTestData(t *testing.T) {
	tests := []struct {
		name         string
		inputPath    string
		expectedType InputType
	}{
		{
			name:         "Real JSON file",
			inputPath:    "../../testdata/test_en_us.json",
			expectedType: InputTypeLanguageFile,
		},
		{
			name:         "Real LANG file",
			inputPath:    "../../testdata/test_en_us.lang",
			expectedType: InputTypeLanguageFile,
		},
		{
			name:         "Real JAR file",
			inputPath:    "../../testdata/test_mod.jar",
			expectedType: InputTypeJARFile,
		},
		{
			name:         "Real BetterQuesting file",
			inputPath:    "../../testdata/sample_betterquesting.json",
			expectedType: InputTypeBetterQuesting,
		},
		{
			name:         "Real Minecraft instance",
			inputPath:    "../../testdata/minecraft_instance_example",
			expectedType: InputTypeMinecraftInstance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if test file exists
			if _, err := os.Stat(tt.inputPath); os.IsNotExist(err) {
				t.Skipf("Test file does not exist: %s", tt.inputPath)
			}

			result := DetectInputType(tt.inputPath)
			if result != tt.expectedType {
				t.Errorf("DetectInputType(%s) = %v, want %v", tt.inputPath, result, tt.expectedType)
			}
		})
	}
}

func TestIsJARFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.jar", true},
		{"test.JAR", true},
		{"mod-1.2.3.jar", true},
		{"test.json", false},
		{"test.lang", false},
		{"test.txt", false},
		{"", false},
		{"no-extension", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := IsJARFile(tt.filename)
			if result != tt.expected {
				t.Errorf("IsJARFile(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestIsLanguageFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"en_us.json", true},
		{"en_us.lang", true},
		{"en_us.snbt", true},
		{"EN_US.JSON", true},
		{"EN_US.LANG", true},
		{"EN_US.SNBT", true},
		{"test.jar", false},
		{"test.txt", false},
		{"", false},
		{"no-extension", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isLanguageFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isLanguageFile(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestInputTypeString(t *testing.T) {
	tests := []struct {
		inputType InputType
		expected  string
	}{
		{InputTypeLanguageFile, "Language File"},
		{InputTypeJARFile, "JAR File"},
		{InputTypeBetterQuesting, "BetterQuesting File"},
		{InputTypeMinecraftInstance, "Minecraft Instance"},
		{InputTypeUnknown, "Unknown"},
		{InputType(999), "Unknown"}, // Test unknown value
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.inputType.String()
			if result != tt.expected {
				t.Errorf("InputType(%d).String() = %s, want %s", tt.inputType, result, tt.expected)
			}
		})
	}
}
