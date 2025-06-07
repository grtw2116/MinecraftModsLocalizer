from pathlib import Path
from provider import provide_minecraft_directory


def get_base_directory() -> Path:
    """Get the base directory for Minecraft instance."""
    minecraft_dir = provide_minecraft_directory()
    if minecraft_dir:
        return Path(minecraft_dir)
    return Path('.')


def get_resource_dir() -> Path:
    """Get the resource directory path."""
    return get_base_directory() / 'resourcepacks' / 'japanese'


def get_mods_dir() -> Path:
    """Get the mods directory path."""
    return get_base_directory() / 'mods'


def get_betterquesting_dir() -> Path:
    """Get the BetterQuesting directory path."""
    return get_base_directory() / 'resources' / 'betterquesting' / 'lang'


def get_ftbquests_dir1() -> Path:
    """Get the FTBQuests kubejs assets directory path."""
    return get_base_directory() / 'kubejs' / 'assets' / 'kubejs' / 'lang'


def get_ftbquests_dir2() -> Path:
    """Get the FTBQuests ftbquests assets directory path."""
    return get_base_directory() / 'kubejs' / 'assets' / 'ftbquests' / 'lang'


def get_ftbquests_dir3() -> Path:
    """Get the FTBQuests chapters directory path."""
    return get_base_directory() / 'config' / 'ftbquests' / 'quests' / 'chapters'


def get_ftbquests_dir4() -> Path:
    """Get the FTBQuests chapter groups file path."""
    return get_base_directory() / 'config' / 'ftbquests' / 'quests' / 'chapter_groups.snbt'


# Directory paths for different mod types (deprecated - use functions above)
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
