# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MinecraftModsLocalizer is a cross-platform GUI application for translating Minecraft mod files. The project is in early development stage, transitioning from a basic Go project to a Wails v2-based application.

## Current Development Status

This is an early-stage project with only basic Go module initialization completed. The main development tasks ahead are implementing the Wails v2 framework and building the core translation functionality.

## Common Commands

### Basic Go Commands
- `go run main.go` - Run the current basic CLI version
- `go mod tidy` - Clean up module dependencies
- `go build` - Build the application

### Current CLI Usage
- `./localizer -input [path] -lang [locale]` - Basic translation
- `./localizer -input [path] -minecraft-version [version] -lang [locale]` - Version-aware locale formatting
- `./localizer -input [path] -dry-run` - Preview without translating
- `./localizer -input [path] -extract-only` - Extract without translating

### Locale Code Formatting
The application now supports version-aware locale code generation:
- **Minecraft 1.11+**: Uses lowercase format (e.g., `ja_jp.json`, `ja_jp.lang`)
- **Minecraft 1.10.2 and earlier**: Uses mixed-case format (e.g., `ja_JP.json`, `ja_JP.lang`)
- Use `--minecraft-version` flag to specify target version (default: 1.20)

### Future Wails Commands (after Wails initialization)
- `wails dev` - Run development server with hot reload
- `wails build` - Build production application for current platform
- `wails build -platform windows,darwin,linux` - Cross-platform build

## Architecture Overview

### Planned Architecture (as per SPEC.md)

**GUI Framework**: Wails v2 for OS-native UI across Windows/Mac/Linux

**File Format Support**:
- JSON format (`{"key": "value"}`) for modern Minecraft versions
- .lang format (`key=value`) for legacy Minecraft versions  
- SNBT format for Structure NBT data

**Translation Engine Integration**:
- Google Translate API
- DeepL API
- OpenAI API-compatible LLMs (default: gpt-4.1-mini)

**Core Components** (to be implemented):
1. File format parsers for JSON/LANG/SNBT
2. Translation engine abstraction layer
3. GUI components for file selection, engine choice, progress tracking
4. Cross-platform native UI integration via Wails

## Development Priority Order

1. **Wails v2 Project Initialization** - Replace current basic Go structure
2. **File Format Parsers** - Implement JSON, .lang, and SNBT parsing
3. **Translation API Integration** - Build abstraction for multiple translation services
4. **GUI Implementation** - Create native UI for all supported platforms

## Important Notes

- The project specification is documented in `SPEC.md`
- Translation source language defaults to English with future auto-detection planned
- Target language will be user-selectable
- Focus on OS-native look and feel across all platforms