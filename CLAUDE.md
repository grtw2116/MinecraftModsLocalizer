# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MinecraftModsLocalizer is a Python application that translates Minecraft mods, quests, and guidebooks to Japanese using OpenAI's ChatGPT API. The tool supports:
- Mod translation (extracting and translating lang files from .jar files)
- FTBQuests and BetterQuesting translation
- Patchouli guidebook translation

## Architecture

The application follows a modular architecture with clear separation of concerns:

### Core Components
- `main.py` - GUI entry point using TkEasyGUI for user interaction
- `provider.py` - Global configuration management (API keys, models, prompts, directories)
- `prepare.py` - Text preprocessing, chunking, and translation coordination
- `chatgpt.py` - OpenAI API integration for translation

### Translation Modules
- `mod.py` - Minecraft mod .jar file processing and resource pack generation
- `quests.py` - FTBQuests and BetterQuesting file handling
- `patchouli.py` - Patchouli guidebook translation

### Support Modules
- `init.py` - Directory paths and constants configuration
- `log.py` - Logging setup and management
- `update.py` - Version checking functionality

### Key Architectural Patterns
- **Translation Pipeline**: Text extraction → chunking → translation → reassembly
- **Error Recovery**: Multiple translation attempts (up to 5) with line count validation
- **File Structure Preservation**: Maintains original file structures and formats
- **Chunked Processing**: Configurable chunk sizes for balancing speed vs accuracy

## Common Development Commands

### Local Development
```bash
# Install dependencies
pip install pyzipper requests pyinstaller TkEasyGUI openai

# Run the application
python src/main.py
```

### Building Executables
The project uses PyInstaller with Docker for cross-platform builds:

```bash
# Build for Linux
docker build -f linux/Dockerfile .

# Build for Windows  
docker build -f windows/Dockerfile .

# Development build
docker build -f dev/Dockerfile .
```

### Testing
- Place the executable in a Minecraft directory with `mods`, `config`, `resourcepacks` folders
- Requires OpenAI API key for functionality testing

## Key Configuration

### Directory Structure Expected
```
minecraft-directory/
├── minecraft-mods-localizer.exe
├── mods/                    # Mod .jar files
├── config/ftbquests/        # Quest configuration
├── kubejs/assets/           # KubeJS language files
├── resourcepacks/           # Output location
└── logs/localizer/          # Application logs and backups
```

### Translation Parameters
- **Chunk Size**: Configurable line count per translation request (1 for accuracy, 100+ for speed)
- **Model**: Default `gpt-4o-mini-2024-07-18`
- **Retry Logic**: Up to 5 attempts with line count validation
- **Output Validation**: Ensures pre/post translation line counts match

### File Processing Logic
- **Mods**: Extracts `en_us.json`/`ja_jp.json` from .jar files, creates resource packs
- **Quests**: Handles both JSON-based and SNBT direct modification
- **Patchouli**: Processes guidebook JSON files within mod archives

## Important Implementation Details

- Translation preserves special formatting characters (§, backslashes, programming variables)
- Failed translations are logged to `logs/localizer/error/` for manual review
- Resource pack generation includes automatic pack.mcmeta creation
- Line-by-line translation with strict line count preservation
- Automatic backup creation before file modification