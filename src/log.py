import logging
import os
import sys
from pathlib import Path
from typing import Union


def setup_logging(directory: Union[str, Path]) -> None:
    """Setup logging configuration with file and console handlers.
    
    Args:
        directory: Directory path where log file will be created
        
    Raises:
        OSError: If directory creation fails
    """
    log_file = "translate.log"
    directory = Path(directory)

    try:
        # Create directory if it doesn't exist
        directory.mkdir(parents=True, exist_ok=True)
        
        # Full path to log file
        log_path = directory / log_file

        # Logger configuration
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s %(levelname)s %(message)s',
            handlers=[
                logging.FileHandler(log_path, encoding='utf-8'),
                logging.StreamHandler(sys.stdout)
            ]
        )
        
    except OSError as e:
        print(f"Failed to setup logging: {e}")
        # Fallback to console-only logging
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s %(levelname)s %(message)s',
            handlers=[logging.StreamHandler(sys.stdout)]
        )
