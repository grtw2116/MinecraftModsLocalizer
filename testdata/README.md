# Test Data

This directory contains test files for MinecraftModsLocalizer.

## Files

### Individual Language Files
- `test_en_us.json` - Sample JSON format language file
- `test_en_us.lang` - Sample legacy .lang format language file
- `test_en_us_translated.json` - Example of translated JSON output
- `test_en_us_translated.lang` - Example of translated .lang output

### JAR Files
- `test_mod.jar` - Sample mod JAR containing language files

### Minecraft Instance Example
- `minecraft_instance_example/` - Mock Minecraft instance structure
  - `mods/` - Contains sample mod JARs for batch processing tests
  - `config/`, `saves/`, `versions/`, `resourcepacks/` - Standard Minecraft directories

## Usage

### Test Individual File Translation
```bash
./MinecraftModsLocalizer -input testdata/test_en_us.json -lang ja -dry-run
```

### Test JAR Processing
```bash
./MinecraftModsLocalizer -input testdata/test_mod.jar -extract-only
```

### Test Minecraft Instance Processing
```bash
./MinecraftModsLocalizer -input testdata/minecraft_instance_example -dry-run
```