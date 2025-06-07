from typing import Optional
from dataclasses import dataclass


@dataclass
class TranslationConfig:
    """Configuration for translation settings."""
    api_key: Optional[str] = None
    chunk_size: int = 1
    model: str = 'gpt-4o-mini-2024-07-18'
    log_directory: Optional[str] = None
    _prompt: Optional[str] = None
    
    @property
    def prompt(self) -> str:
        """Get the translation prompt template."""
        if self._prompt is not None:
            return self._prompt
        return self._get_default_prompt()
    
    @prompt.setter
    def prompt(self, value: str) -> None:
        """Set a custom translation prompt."""
        self._prompt = value
    
    def _get_default_prompt(self) -> str:
        """Get the default translation prompt template."""
        return """You are a professional translator. Please translate the following English text into Japanese, one line at a time, step by step, in order
Make sure that the number of lines of text before and after translation is the same. Never add or subtract extra lines.

# The number of lines of text to pass: {line_count}

# Pay attention to the details below
- Never include any greeting other than the translation result!
- **Translate line by line, step by step, in order.**
- **Make sure that the number of lines of text before and after translation is the same. Never add or subtract extra lines.**
- **The meaning of the sentences before and after may be connected by chance, but if the lines are different, they are different sentences, so do not mix them up!**
- **If multiple sentences are written on a single line, please translate as is, with all sentences on a single line.**
- Proper nouns may be included and can be written in Katakana.
- The backslash may be used as an escape character. Please maintain.
- There might be programming variable characters such as %s, 1, or \\"; please retain these.
- Do not edit any other characters that may look like special symbols.

# Example

### input
§6Checks for ore behind the
§6walls, floors or ceilings.
Whether or not mining fatigue is applied to players in the temple
if it has not yet been cleared.

### incorrect output
§6壁、床、または天井の後ろにある鉱石をチェックします。
まだクリアされていない場合、寺院内のプレイヤーにマイニング疲労が適用されるかどうか。

### correct output
§6後ろにある鉱石をチェックします。
§6壁、床、または天井
寺院内のプレイヤーにマイニング疲労が適用されるかどうか。
もしクリアされていない場合


### input
Add a new requirement group.Requirement groups can hold multiplerequirements and basicallymake them one big requirement.Requirement groups have two modes.In §zAND §rmode, all requirements needto return TRUE (which means "Yes, load!"),but in §zOR §rmode, only one requirementneeds to return TRUE.

### incorrect output
新しい要件グループを追加します。
要件グループは複数の要件を保持でき、基本的にそれらを1つの大きな要件にまとめます。要件グループには2つのモードがあります。
§zAND §rモードでは、すべての要件がTRUE（「はい、ロードする！」を意味します）を返す必要がありますが、§zOR §rモードでは、1つの要件だけがTRUEを返す必要があります。

### correct output
新しい要件グループを追加します。要件グループは複数の要件を保持でき、基本的にそれらを1つの大きな要件にまとめます。要件グループには2つのモードがあります。§zAND §rモードでは、すべての要件がTRUE（「はい、ロードする！」を意味します）を返す必要がありますが、§zOR §rモードでは、1つの要件だけがTRUEを返す必要があります。"""


class ConfigManager:
    """Singleton configuration manager for the translation application."""
    _instance: Optional['ConfigManager'] = None
    _config: TranslationConfig
    
    def __new__(cls) -> 'ConfigManager':
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._config = TranslationConfig()
        return cls._instance
    
    @property
    def config(self) -> TranslationConfig:
        """Get the current configuration."""
        return self._config
    
    def update_config(self, **kwargs) -> None:
        """Update configuration with provided keyword arguments."""
        for key, value in kwargs.items():
            if hasattr(self._config, key):
                setattr(self._config, key, value)
            else:
                raise ValueError(f"Unknown configuration key: {key}")


_config_manager = ConfigManager()


def provide_api_key() -> Optional[str]:
    """Get the current API key."""
    return _config_manager.config.api_key


def set_api_key(api_key: str) -> None:
    """Set the API key."""
    _config_manager.config.api_key = api_key


def provide_chunk_size() -> int:
    """Get the current chunk size."""
    return _config_manager.config.chunk_size


def set_chunk_size(chunk_size: int) -> None:
    """Set the chunk size."""
    _config_manager.config.chunk_size = chunk_size


def provide_model() -> str:
    """Get the current model."""
    return _config_manager.config.model


def set_model(model: str) -> None:
    """Set the model."""
    _config_manager.config.model = model


def provide_prompt() -> str:
    """Get the current prompt template."""
    return _config_manager.config.prompt


def set_prompt(prompt: str) -> None:
    """Set a custom prompt template."""
    _config_manager.config.prompt = prompt


def provide_log_directory() -> Optional[str]:
    """Get the current log directory."""
    return _config_manager.config.log_directory


def set_log_directory(log_directory: str) -> None:
    """Set the log directory."""
    _config_manager.config.log_directory = log_directory


def get_config() -> TranslationConfig:
    """Get the configuration instance for direct access."""
    return _config_manager.config