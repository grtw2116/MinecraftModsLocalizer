import itertools
import json
import logging
import os
import re
from typing import Dict, List, Optional, Tuple

from init import MAX_ATTEMPTS
from chatgpt import translate_with_chatgpt
from provider import get_config


def extract_map_from_lang(filepath: str) -> Dict[str, str]:
    """Extract key-value pairs from .lang file format.
    
    Args:
        filepath: Path to the .lang file
        
    Returns:
        Dictionary mapping keys to values
    """
    collected_map = {}
    try:
        with open(filepath, 'r', encoding='utf-8') as file:
            for line in file:
                line = line.strip()
                if line and not line.startswith('#'):
                    if '=' in line:
                        key, value = line.split('=', 1)
                        collected_map[key.strip()] = value.strip()
    except (FileNotFoundError, IOError) as e:
        logging.error(f"Failed to read lang file {filepath}: {e}")
    return collected_map


def extract_map_from_json(file_path: str) -> Dict[str, str]:
    """Extract English text entries from JSON language file.
    
    Args:
        file_path: Path to the JSON language file
        
    Returns:
        Dictionary mapping keys to English text values
    """
    collected_map = {}

    if not os.path.exists(file_path):
        logging.info(f"Could not find {file_path}. Skipping this mod for translation.")
        return collected_map

    logging.info(f"Extract keys in en_us.json(or ja_jp.json) in {file_path}")
    try:
        with open(file_path, 'r', encoding="utf-8") as f:
            content = json.load(f)

        collected_map = _filter_english_entries(content)

    except json.JSONDecodeError as e:
        logging.error(
            f"Failed to load or process JSON from {file_path}: {e}. "
            f"Skipping this mod for translation. Please check the file for syntax errors.")
    except (FileNotFoundError, IOError) as e:
        logging.error(f"Failed to read JSON file {file_path}: {e}")

    return collected_map


def _filter_english_entries(content: Dict) -> Dict[str, str]:
    """Filter JSON content to extract only English text entries.
    
    Args:
        content: Parsed JSON content
        
    Returns:
        Dictionary with English text entries only
    """
    collected_map = {}
    japanese_pattern = re.compile(r'[\u3040-\u30FF\u3400-\u4DBF\u4E00-\u9FFF]')
    
    for key, value in content.items():
        if (not key.startswith("_comment") and 
            isinstance(value, str) and 
            not japanese_pattern.search(value)):
            collected_map[key] = value
            
    return collected_map


def split_list(big_list: List[str]) -> List[List[str]]:
    """Split a large list into smaller chunks based on configuration.
    
    Args:
        big_list: List to be split into chunks
        
    Returns:
        List of chunks, each containing at most chunk_size elements
    """
    list_of_chunks = []
    config = get_config()

    for i in range(0, len(big_list), config.chunk_size):
        chunk = big_list[i:i + config.chunk_size]
        list_of_chunks.append(chunk)

    return list_of_chunks


def create_map_with_none_filling(split_target: List[str], translated_split_target: List[str]) -> Dict[str, Optional[str]]:
    """Create a mapping between original and translated text, handling mismatched lengths.
    
    Args:
        split_target: Original text list
        translated_split_target: Translated text list
        
    Returns:
        Dictionary mapping original to translated text, with None for missing translations
    """
    result_map = {}
    for key, value in itertools.zip_longest(split_target, translated_split_target):
        if value == '':
            value = None
        result_map[key] = value

    return result_map


class TranslationError(Exception):
    """Custom exception for translation errors."""
    pass


def _validate_translation_result(original: List[str], translated: List[str]) -> bool:
    """Validate that translation result has correct line count.
    
    Args:
        original: Original text lines
        translated: Translated text lines
        
    Returns:
        True if line counts match or filtered counts match
    """
    if len(original) == len(translated):
        return True
        
    # Try filtering empty lines and check again
    filtered_original = [item for item in original if item.strip()]
    filtered_translated = [item for item in translated if item.strip()]
    
    return len(filtered_original) == len(filtered_translated)


def _translate_chunk_with_retry(split_target: List[str], timeout: int = 180) -> Dict[str, str]:
    """Translate a chunk with retry logic.
    
    Args:
        split_target: List of text to translate
        timeout: Timeout in seconds for translation
        
    Returns:
        Dictionary mapping original to translated text
        
    Raises:
        TranslationError: If translation fails after all retries
    """
    for attempt in range(MAX_ATTEMPTS):
        try:
            translated_split_target = translate_with_chatgpt(split_target, timeout)
            
            if _validate_translation_result(split_target, translated_split_target):
                return dict(zip(split_target, translated_split_target))
                
        except Exception as e:
            logging.warning(f"Translation attempt {attempt + 1} failed: {e}")
            
        if attempt == MAX_ATTEMPTS - 1:
            raise TranslationError(
                f"Failed to translate chunk after {MAX_ATTEMPTS} attempts. "
                f"Original lines: {len(split_target)}, "
                f"Translated lines: {len(translated_split_target) if 'translated_split_target' in locals() else 'unknown'}"
            )
    
    return {}


def prepare_translation(targets: List[str]) -> Dict[str, str]:
    """Prepare and execute translation of text targets.
    
    Args:
        targets: List of text strings to translate
        
    Returns:
        Dictionary mapping original text to translated text
    """
    split_targets = split_list(targets)
    result_map = {}

    logging.info(f"The file contains {len(targets)} lines")
    logging.info(f"Splitting the file into {len(split_targets)} chunks for translation...")

    for index, split_target in enumerate(split_targets, 1):
        logging.info(f"Translating chunk {index}/{len(split_targets)}...")
        
        try:
            chunk_result = _translate_chunk_with_retry(split_target)
            result_map.update(chunk_result)
            
        except TranslationError as e:
            logging.error(f"Failed to translate chunk {index}/{len(split_targets)}: {e}")
            logging.error(f"Original chunk content: {split_target}")
            # Continue with remaining chunks despite failure

    logging.info("Translation completed!")
    return result_map
