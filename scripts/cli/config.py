"""Configuration module for CLI."""
import os
from pathlib import Path
from dotenv import load_dotenv

# Load environment variables from .env file if it exists
env_path = Path(__file__).parent.parent.parent / '.env'
if env_path.exists():
    load_dotenv(env_path)

# API Configuration
API_BASE_URL = os.getenv('API_BASE_URL', 'http://localhost:8080')
API_TIMEOUT = int(os.getenv('API_TIMEOUT', '30'))

# Output Configuration
OUTPUT_FORMAT = os.getenv('CLI_OUTPUT_FORMAT', 'table')  # table, json, yaml
VERBOSE = os.getenv('CLI_VERBOSE', 'false').lower() == 'true'
