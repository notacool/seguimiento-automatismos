"""Tests for configuration module."""
import os
import pytest


class TestConfig:
    """Test suite for configuration."""

    def test_default_api_base_url(self):
        """Test default API_BASE_URL value."""
        # Remove env var if exists
        if 'API_BASE_URL' in os.environ:
            del os.environ['API_BASE_URL']

        # Reimport to get default value
        import importlib
        import config
        importlib.reload(config)

        assert config.API_BASE_URL == 'http://localhost:8080'

    def test_custom_api_base_url(self, monkeypatch):
        """Test custom API_BASE_URL from environment."""
        custom_url = 'http://production:9090'
        monkeypatch.setenv('API_BASE_URL', custom_url)

        import importlib
        import config
        importlib.reload(config)

        assert config.API_BASE_URL == custom_url

    def test_default_api_timeout(self):
        """Test default API_TIMEOUT value."""
        if 'API_TIMEOUT' in os.environ:
            del os.environ['API_TIMEOUT']

        import importlib
        import config
        importlib.reload(config)

        assert config.API_TIMEOUT == 30

    def test_custom_api_timeout(self, monkeypatch):
        """Test custom API_TIMEOUT from environment."""
        monkeypatch.setenv('API_TIMEOUT', '60')

        import importlib
        import config
        importlib.reload(config)

        assert config.API_TIMEOUT == 60

    def test_verbose_flag_default(self):
        """Test VERBOSE default value."""
        if 'CLI_VERBOSE' in os.environ:
            del os.environ['CLI_VERBOSE']

        import importlib
        import config
        importlib.reload(config)

        assert config.VERBOSE is False

    def test_verbose_flag_true(self, monkeypatch):
        """Test VERBOSE when set to true."""
        monkeypatch.setenv('CLI_VERBOSE', 'true')

        import importlib
        import config
        importlib.reload(config)

        assert config.VERBOSE is True
