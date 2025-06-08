# MinecraftModsLocalizer

Minecraft mod files (JAR/JSON/LANG) translator with AI-powered translation engines.

## Features

- **Multiple Input Formats**: JAR files, JSON language files, and legacy .lang files
- **JAR Extraction**: Extract language files directly from mod JAR files
- **Resource Pack Generation**: Create proper Minecraft resource packs with translated content
- **AI Translation**: Support for OpenAI API-compatible LLMs with consistency checking
- **Similarity Matching**: Reuse previous translations for consistent terminology
- **Cross-platform**: Works on Windows, macOS, and Linux

## Installation

### Prerequisites
- Go 1.21 or higher

### Build from Source
```bash
git clone https://github.com/grtw2116/MinecraftModsLocalizer.git
cd MinecraftModsLocalizer
go build -o MinecraftModsLocalizer ./cmd
```

## Usage

### Basic Translation
```bash
# Translate a JSON language file
./MinecraftModsLocalizer -input en_us.json -lang ja_jp

# Translate with custom similarity threshold
./MinecraftModsLocalizer -input en_us.json -lang ja_jp -similarity 0.8
```

### JAR File Processing
```bash
# Translate JAR and generate resource pack
./MinecraftModsLocalizer -input mod.jar -lang ja_jp

# Extract language files from a mod JAR
./MinecraftModsLocalizer -input mod.jar -extract-only

# Dry run to see what would be translated
./MinecraftModsLocalizer -input mod.jar -lang ja_jp -dry-run
```

### Command Line Options

```
Usage: ./MinecraftModsLocalizer [options]

Options:
  -input string
        Input file path (JAR file or individual language file)
  -output string
        Output file path (optional, defaults to input_translated.ext or resource pack)
  -lang string
        Target language code (default: ja)
  -engine string
        Translation engine: openai, google, deepl (default: openai)
  -similarity float
        Similarity threshold for finding similar examples (0.0-1.0, default: 0.6)
  -extract-only
        Extract language files from JAR without translating
  -resource-pack
        Generate resource pack format output
  -dry-run
        Parse file and show statistics without translating
  -help
        Show help
```

## Supported Formats

- **JAR files**: Minecraft mod files (`.jar`)
- **JSON files**: Modern Minecraft language files (`.json`)
- **LANG files**: Legacy Minecraft language files (`.lang`)
- **SNBT files**: Structure NBT text format (`.snbt`) - *planned*

## Supported Languages

- Japanese (`ja`)
- Korean (`ko`)
- Simplified Chinese (`zh-cn`)
- Traditional Chinese (`zh-tw`)
- French (`fr`)
- German (`de`)
- Spanish (`es`)
- And more...

## Configuration

### Environment Variables

Set your API key for translation services:

```bash
# For OpenAI or compatible APIs
export OPENAI_API_KEY="your-api-key-here"

# For custom API endpoints
export OPENAI_BASE_URL="https://api.custom-provider.com/v1"
export OPENAI_MODEL="gpt-4o-mini"
```

### Translation Consistency

The tool automatically builds a translation dictionary (`dictionary.json`) to ensure consistent translations across multiple files. Similar phrases are matched using edit distance algorithms to maintain terminology consistency.

## Examples

### Translating a Single Language File
```bash
./MinecraftModsLocalizer -input assets/mymod/lang/en_us.json -lang ja_jp -engine openai
```

### Processing Multiple Mods
```bash
# Extract all language files first
./MinecraftModsLocalizer -input mod1.jar -extract-only -output extracted/

# Translate and create resource packs
./MinecraftModsLocalizer -input mod1.jar -lang ja_jp -resource-pack -output resourcepack_ja/
./MinecraftModsLocalizer -input mod2.jar -lang ja_jp -resource-pack -output resourcepack_ja/
```

### Resource Pack Structure

When using `-resource-pack`, the tool generates a proper Minecraft resource pack:

```
resource_pack_ja/
├── pack.mcmeta
└── assets/
    └── modname/
        └── lang/
            └── ja.json
```

## Development Status

This project is currently in early development. The core translation engine and JAR processing features are implemented. Future plans include:

- Wails v2 GUI implementation
- Google Translate and DeepL API integration
- Batch processing capabilities
- Translation memory management

## Contributing

This project is under active development. Please check the [SPEC.md](SPEC.md) file for detailed technical specifications.

## License

[License information to be added]
