from pathlib import Path

# Directory paths for different mod types
RESOURCE_DIR: Path = Path('./resourcepacks/japanese')
MODS_DIR: Path = Path('./mods')
BETTERQUESTING_DIR: Path = Path('./resources/betterquesting/lang')
FTBQUESTS_DIR1: Path = Path('./kubejs/assets/kubejs/lang')
FTBQUESTS_DIR2: Path = Path('./kubejs/assets/ftbquests/lang')
FTBQUESTS_DIR3: Path = Path('./config/ftbquests/quests/chapters')
FTBQUESTS_DIR4: Path = Path('./config/ftbquests/quests/chapter_groups.snbt')

# Translation retry configuration
MAX_ATTEMPTS: int = 5

# GitHub repository information
USER: str = 'Y-RyuZU'
REPO: str = 'MinecraftModsLocalizer'
VERSION: str = 'v2.1.3'
