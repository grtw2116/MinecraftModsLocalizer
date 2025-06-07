import json
import logging
import os
import zipfile
from pathlib import Path
from typing import Dict, Optional, Union

from init import get_resource_dir, get_mods_dir
from prepare import extract_map_from_json, extract_map_from_lang, prepare_translation
from provider import provide_log_directory


def process_jar_file(jar_path: Union[str, Path]) -> tuple[Dict[str, str], str]:
    mod_name = get_mod_name_from_jar(jar_path)
    if mod_name is None:
        logging.info(f"Could not determine mod name for {jar_path}")
        return {}, 'json'

    lang_path_in_jar = Path(f'assets/{mod_name}/lang/')
    
    # JSON files
    ja_jp_json_path = os.path.join(lang_path_in_jar, 'ja_jp.json')
    en_us_json_path = os.path.join(lang_path_in_jar, 'en_us.json')
    
    # Lang files  
    ja_jp_lang_path = os.path.join(lang_path_in_jar, 'ja_jp.lang')
    en_us_lang_path = os.path.join(lang_path_in_jar, 'en_us.lang')
    
    # Convert to string paths for zipfile compatibility
    ja_jp_json_str = str(ja_jp_json_path).replace('\\', '/')
    en_us_json_str = str(en_us_json_path).replace('\\', '/')
    ja_jp_lang_str = str(ja_jp_lang_path).replace('\\', '/')
    en_us_lang_str = str(en_us_lang_path).replace('\\', '/')

    log_dir = provide_log_directory()
    if log_dir is None:
        logging.error("Log directory not configured")
        return {}, 'json'

    logging.info(f"Extract language files from {jar_path / lang_path_in_jar}")
    with zipfile.ZipFile(jar_path, 'r') as zip_ref:
        # Extract JSON files
        if en_us_json_str in zip_ref.namelist():
            extract_specific_file(jar_path, en_us_json_str, log_dir)
        if ja_jp_json_str in zip_ref.namelist():
            extract_specific_file(jar_path, ja_jp_json_str, log_dir)
        
        # Extract lang files
        if en_us_lang_str in zip_ref.namelist():
            extract_specific_file(jar_path, en_us_lang_str, log_dir)
        if ja_jp_lang_str in zip_ref.namelist():
            extract_specific_file(jar_path, ja_jp_lang_str, log_dir)

    # Define local file paths
    en_us_json_local = os.path.join(log_dir, en_us_json_path)
    ja_jp_json_local = os.path.join(log_dir, ja_jp_json_path)
    en_us_lang_local = os.path.join(log_dir, en_us_lang_path)
    ja_jp_lang_local = os.path.join(log_dir, ja_jp_lang_path)

    # Priority: ja_jp.json > en_us.json > ja_jp.lang > en_us.lang
    if os.path.exists(ja_jp_json_local):
        return extract_map_from_json(ja_jp_json_local), 'json'
    elif os.path.exists(en_us_json_local):
        return extract_map_from_json(en_us_json_local), 'json'
    elif os.path.exists(ja_jp_lang_local):
        return extract_map_from_lang(ja_jp_lang_local), 'lang'
    elif os.path.exists(en_us_lang_local):
        return extract_map_from_lang(en_us_lang_local), 'lang'
    
    return {}, 'json'


def translate_from_jar() -> None:
    resource_dir = get_resource_dir()
    mods_dir = get_mods_dir()
    
    if not os.path.exists(resource_dir):
        os.makedirs(os.path.join(resource_dir, 'assets', 'japanese', 'lang'))

    targets = {}
    output_format = 'json'  # Default to JSON format

    extracted_pack_mcmeta = False
    for filename in os.listdir(mods_dir):
        if filename.endswith('.jar'):
            # Extract pack.mcmeta if it exists in the jar
            if not extracted_pack_mcmeta:
                extracted_pack_mcmeta = extract_specific_file(os.path.join(mods_dir, filename), 'pack.mcmeta',
                                                              resource_dir)
                update_resourcepack_description(os.path.join(resource_dir, 'pack.mcmeta'), '日本語化パック')

            jar_targets, jar_format = process_jar_file(os.path.join(mods_dir, filename))
            targets.update(jar_targets)
            # Use lang format if any jar contains lang files
            if jar_format == 'lang':
                output_format = 'lang'

    translated_map = prepare_translation(list(targets.values()))

    translated_targets = {json_key: translated_map[original] for json_key, original in targets.items() if
                          original in translated_map}

    untranslated_items = {json_key: original for json_key, original in targets.items() if
                          original not in translated_map}

    # Write output in the appropriate format
    if output_format == 'json':
        with open(os.path.join(resource_dir, 'assets', 'japanese', 'lang', 'ja_jp.json'), 'w', encoding="utf-8") as f:
            json.dump(dict(sorted(translated_targets.items())), f, ensure_ascii=False, indent=4)
    else:  # lang format
        with open(os.path.join(resource_dir, 'assets', 'japanese', 'lang', 'ja_jp.lang'), 'w', encoding="utf-8") as f:
            for key, value in sorted(translated_targets.items()):
                f.write(f'{key}={value}\n')

    log_dir = provide_log_directory()
    if log_dir is None:
        logging.error("Log directory not configured for error handling")
        return
    error_directory = os.path.join(log_dir, 'error')

    if not os.path.exists(error_directory):
        os.makedirs(error_directory)

    with open(os.path.join(error_directory, 'mod_ja_jp.json'), 'w', encoding="utf-8") as f:
        json.dump(dict(sorted(untranslated_items.items())), f, ensure_ascii=False, indent=4)


def update_resourcepack_description(file_path: Union[str, Path], new_description: str) -> None:
    # ファイルが存在するか確認
    if not os.path.exists(file_path):
        return

    with open(file_path, 'r', encoding='utf-8') as file:
        try:
            data = json.load(file)
        except json.JSONDecodeError as e:
            return

    # 'description'の'text'を新しい値に更新
    try:
        if 'pack' in data and 'description' in data['pack'] and 'text' in data['pack']['description']:
            data['pack']['description']['text'] = new_description
        else:
            return
    except Exception as e:
        return

    # 変更を加えたデータを同じファイルに書き戻す
    with open(file_path, 'w', encoding='utf-8') as file:
        try:
            json.dump(data, file, ensure_ascii=False, indent=2)  # JSONを整形して書き込み
        except Exception as e:
            return


def get_mod_name_from_jar(jar_path: Union[str, Path]) -> Optional[str]:
    with zipfile.ZipFile(jar_path, 'r') as zip_ref:
        asset_dirs_with_lang = set()
        for name in zip_ref.namelist():
            parts = name.split('/')
            if len(parts) > 3 and parts[0] == 'assets' and parts[2] == 'lang' and parts[1] != 'minecraft':
                asset_dirs_with_lang.add(parts[1])
        if asset_dirs_with_lang:
            return list(asset_dirs_with_lang)[0]
    return None


def extract_specific_file(zip_filepath: Union[str, Path], file_name: str, dest_dir: Union[str, Path]) -> bool:
    with zipfile.ZipFile(zip_filepath, 'r') as zip_ref:
        if file_name in zip_ref.namelist():
            zip_ref.extract(file_name, dest_dir)
            return True
        else:
            logging.info(f"The file {file_name} in {zip_filepath} was not found in the ZIP archive.")
    return False
