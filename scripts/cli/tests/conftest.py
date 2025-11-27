"""Pytest fixtures for CLI tests."""
import pytest
from api_client import APIClient


@pytest.fixture
def api_base_url():
    """Base URL for API tests."""
    return "http://localhost:8080"


@pytest.fixture
def api_client(api_base_url):
    """Create an API client instance."""
    return APIClient(base_url=api_base_url, timeout=10)


@pytest.fixture
def mock_health_response():
    """Mock health check response."""
    return {
        "status": "healthy",
        "database": "ok",
        "timestamp": "2025-11-27T12:00:00Z"
    }


@pytest.fixture
def mock_task_response():
    """Mock task response."""
    return {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Test Task",
        "state": "IN_PROGRESS",
        "subtasks": [
            {
                "id": "660e8400-e29b-41d4-a716-446655440001",
                "name": "Subtask 1",
                "state": "COMPLETED",
                "created_at": "2025-11-27T10:00:00Z",
                "updated_at": "2025-11-27T10:15:00Z"
            }
        ],
        "created_by": "Test Team",
        "created_at": "2025-11-27T10:00:00Z",
        "updated_at": "2025-11-27T10:15:00Z"
    }


@pytest.fixture
def mock_task_list_response():
    """Mock task list response."""
    return {
        "tasks": [
            {
                "id": "550e8400-e29b-41d4-a716-446655440000",
                "name": "Task 1",
                "state": "IN_PROGRESS",
                "subtasks": [],
                "created_by": "Team A",
                "created_at": "2025-11-27T10:00:00Z",
                "updated_at": "2025-11-27T10:00:00Z"
            },
            {
                "id": "550e8400-e29b-41d4-a716-446655440001",
                "name": "Task 2",
                "state": "COMPLETED",
                "subtasks": [],
                "created_by": "Team B",
                "created_at": "2025-11-27T09:00:00Z",
                "updated_at": "2025-11-27T11:00:00Z"
            }
        ],
        "pagination": {
            "page": 1,
            "limit": 20,
            "total": 2,
            "total_pages": 1
        }
    }


@pytest.fixture
def mock_error_response():
    """Mock RFC 7807 error response."""
    return {
        "type": "https://api.grupoapi.com/problems/task-not-found",
        "title": "Task Not Found",
        "status": 404,
        "detail": "Task with ID '550e8400-e29b-41d4-a716-446655440000' not found"
    }
