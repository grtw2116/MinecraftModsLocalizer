from init import USER, REPO, VERSION
import requests
import logging
from typing import Optional


def get_latest_release_tag(user: str, repo: str) -> Optional[str]:
    """Get the latest release tag from GitHub repository.
    
    Args:
        user: GitHub username
        repo: Repository name
        
    Returns:
        Latest release tag name, or None if request fails
    """
    url = f"https://api.github.com/repos/{user}/{repo}/releases/latest"
    
    try:
        response = requests.get(url, timeout=10)
        if response.status_code == 200:
            return response.json()['tag_name']
        else:
            logging.error(f"GitHub API error {response.status_code}: {response.text}")
            return None
    except requests.RequestException as e:
        logging.error(f"Failed to fetch latest release: {e}")
        return None
    except KeyError:
        logging.error("Invalid response format from GitHub API")
        return None


def check_version() -> Optional[bool]:
    """Compare current VERSION constant with latest GitHub release tag.
    
    Returns:
        True if versions match, False if different, None if check failed
    """
    latest_tag = get_latest_release_tag(USER, REPO)
    if latest_tag is None:
        logging.warning("Could not check for updates")
        return None
        
    return VERSION == latest_tag
