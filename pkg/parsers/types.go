package parsers

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