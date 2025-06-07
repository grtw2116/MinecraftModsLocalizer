import logging
import re
import time
from typing import List, Optional
from openai import OpenAI
from openai.types.chat import ChatCompletion
from openai import AuthenticationError, RateLimitError, APIError, APITimeoutError

from provider import get_config
from provider import TranslationConfig


class ChatGPTTranslationError(Exception):
    """Custom exception for ChatGPT translation errors."""
    pass


class ChatGPTTranslator:
    """ChatGPT translation service with improved error handling."""
    
    def __init__(self, config: Optional[TranslationConfig] = None):
        """Initialize the translator with configuration.
        
        Args:
            config: Translation configuration, uses global config if None
        """
        self.config = config or get_config()
        self._validate_config()
        self.client = OpenAI(api_key=self.config.api_key)
    
    def _validate_config(self) -> None:
        """Validate the configuration.
        
        Raises:
            ChatGPTTranslationError: If configuration is invalid
        """
        if not self.config.api_key:
            raise ChatGPTTranslationError("OpenAI API key is not configured")
        if not self.config.model:
            raise ChatGPTTranslationError("OpenAI model is not configured")
        if not self.config.prompt:
            raise ChatGPTTranslationError("Translation prompt is not configured")
    
    def _preprocess_text(self, split_target: List[str]) -> List[str]:
        """Preprocess text by removing problematic newlines.
        
        Args:
            split_target: List of text to preprocess
            
        Returns:
            Preprocessed text list
        """
        if len(split_target) <= 1:
            return split_target
        return [line.replace('\\n', '').replace('\n', '') for line in split_target]
    
    def _postprocess_text(self, translated_text: str, original_count: int) -> List[str]:
        """Postprocess translated text.
        
        Args:
            translated_text: Raw translated text from API
            original_count: Number of original lines
            
        Returns:
            Processed list of translated lines
        """
        if original_count > 1:
            result = translated_text.splitlines()
        else:
            result = [translated_text.replace('\n', '')]
        
        # Escape unescaped quotes
        result = [re.sub(r'(?<!\\)"', r'\\"', line) for line in result]
        return result
    
    def _make_api_request(self, processed_target: List[str]) -> ChatCompletion:
        """Make the actual API request to OpenAI.
        
        Args:
            processed_target: Preprocessed text to translate
            
        Returns:
            ChatCompletion response
            
        Raises:
            ChatGPTTranslationError: For various API errors
        """
        try:
            prompt_text = self.config.prompt.replace('{line_count}', str(len(processed_target)))
            
            response = self.client.chat.completions.create(
                model=self.config.model,
                messages=[
                    {
                        "role": "system",
                        "content": [{"type": "text", "text": prompt_text}]
                    },
                    {
                        "role": "user",
                        "content": [{"type": "text", "text": '\n'.join(processed_target)}]
                    }
                ],
            )
            return response
            
        except AuthenticationError as e:
            raise ChatGPTTranslationError(f"Authentication failed: {e}")
        except RateLimitError as e:
            raise ChatGPTTranslationError(f"Rate limit exceeded: {e}")
        except APITimeoutError as e:
            raise ChatGPTTranslationError(f"API request timed out: {e}")
        except APIError as e:
            raise ChatGPTTranslationError(f"OpenAI API error: {e}")
        except Exception as e:
            raise ChatGPTTranslationError(f"Unexpected error during API request: {e}")
    
    def translate(self, split_target: List[str], timeout: int = 180) -> List[str]:
        """Translate text using ChatGPT API.
        
        Args:
            split_target: List of text strings to translate
            timeout: Timeout in seconds (for logging purposes)
            
        Returns:
            List of translated text strings
            
        Raises:
            ChatGPTTranslationError: If translation fails
        """
        start_time = time.time()
        
        if not split_target:
            return []
        
        try:
            # Preprocess input
            processed_target = self._preprocess_text(split_target)
            
            # Make API request
            response = self._make_api_request(processed_target)
            
            # Extract and validate response
            if not response.choices or not response.choices[0].message:
                raise ChatGPTTranslationError("No valid response received from ChatGPT")
            
            translated_text = response.choices[0].message.content
            if not translated_text:
                raise ChatGPTTranslationError("Empty response received from ChatGPT")
            
            # Postprocess result
            result = self._postprocess_text(translated_text, len(split_target))
            
            elapsed_time = time.time() - start_time
            logging.info(f"Translation completed in {elapsed_time:.2f} seconds")
            
            return result
            
        except ChatGPTTranslationError:
            raise
        except Exception as e:
            elapsed_time = time.time() - start_time
            if elapsed_time > timeout:
                logging.error("Timeout reached while waiting for translation")
            raise ChatGPTTranslationError(f"Unexpected error during translation: {e}")


# Legacy function for backward compatibility
def translate_with_chatgpt(split_target: List[str], timeout: int = 180) -> List[str]:
    """Legacy function for ChatGPT translation.
    
    Args:
        split_target: List of text strings to translate
        timeout: Timeout in seconds
        
    Returns:
        List of translated text strings, empty list on error
    """
    try:
        translator = ChatGPTTranslator()
        return translator.translate(split_target, timeout)
    except ChatGPTTranslationError as e:
        logging.error(f"Translation failed: {e}")
        return []
