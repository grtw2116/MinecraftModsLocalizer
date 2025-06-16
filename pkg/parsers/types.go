package parsers

import "strings"

type FileFormat int

const (
	FormatUnknown FileFormat = iota
	FormatJSON
	FormatLang
	FormatSNBT
)

type TranslationData map[string]string

func (f FileFormat) String() string {
	switch f {
	case FormatJSON:
		return "JSON"
	case FormatLang:
		return "LANG"
	case FormatSNBT:
		return "SNBT"
	default:
		return "Unknown"
	}
}

func GetExtensionForFormat(format FileFormat) string {
	switch format {
	case FormatJSON:
		return ".json"
	case FormatLang:
		return ".lang"
	case FormatSNBT:
		return ".snbt"
	default:
		return ".txt"
	}
}

// FormatLocaleCode formats a locale code according to Minecraft version requirements
// For version 1.11 and later: ja_jp (lowercase with underscore)
// For version 1.10.2 and earlier: ja_JP (mixed case with underscore)
func FormatLocaleCode(localeCode, minecraftVersion string) string {
	if IsLegacyMinecraftVersion(minecraftVersion) {
		return FormatLegacyLocaleCode(localeCode)
	}
	return FormatModernLocaleCode(localeCode)
}

// IsLegacyMinecraftVersion checks if the version is 1.10.2 or earlier
func IsLegacyMinecraftVersion(version string) bool {
	// Parse version to compare
	// 1.10.2 and earlier use mixed case (ja_JP)
	// 1.11 and later use lowercase (ja_jp)

	// Handle common version formats
	switch version {
	case "1.7", "1.7.2", "1.7.10":
		return true
	case "1.8", "1.8.8", "1.8.9":
		return true
	case "1.9", "1.9.1", "1.9.2", "1.9.3", "1.9.4":
		return true
	case "1.10", "1.10.1", "1.10.2":
		return true
	case "1.11", "1.11.1", "1.11.2":
		return false
	case "1.12", "1.12.1", "1.12.2":
		return false
	default:
		// For any version not explicitly listed, assume modern format
		// This includes 1.13+, 1.20+, etc.
		return false
	}
}

// FormatModernLocaleCode formats locale for Minecraft 1.11+ (.json files)
// Example: "ja_jp", "zh_cn"
func FormatModernLocaleCode(localeCode string) string {
	return strings.ToLower(localeCode)
}

// FormatLegacyLocaleCode formats locale for Minecraft 1.10.2- (.lang files)
// Example: "ja_JP", "zh_CN"
func FormatLegacyLocaleCode(localeCode string) string {
	parts := strings.Split(strings.ToLower(localeCode), "_")
	if len(parts) == 2 {
		return parts[0] + "_" + strings.ToUpper(parts[1])
	}
	return strings.ToLower(localeCode)
}
