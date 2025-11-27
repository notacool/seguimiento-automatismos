"""Tests for API client."""
import pytest
import responses
from api_client import APIClient, APIError


class TestAPIClient:
    """Test suite for APIClient class."""

    @responses.activate
    def test_health_check_success(self, api_client, mock_health_response):
        """Test successful health check."""
        responses.add(
            responses.GET,
            "http://localhost:8080/health",
            json=mock_health_response,
            status=200
        )

        result = api_client.health_check()

        assert result == mock_health_response
        assert result["status"] == "healthy"
        assert result["database"] == "ok"

    @responses.activate
    def test_health_check_unhealthy(self, api_client):
        """Test health check with unhealthy service."""
        unhealthy_response = {
            "status": "unhealthy",
            "database": "error",
            "timestamp": "2025-11-27T12:00:00Z"
        }
        responses.add(
            responses.GET,
            "http://localhost:8080/health",
            json=unhealthy_response,
            status=503
        )

        with pytest.raises(APIError) as exc_info:
            api_client.health_check()

        assert exc_info.value.status_code == 503

    @responses.activate
    def test_get_task_success(self, api_client, mock_task_response):
        """Test successful get task by ID."""
        task_id = "550e8400-e29b-41d4-a716-446655440000"
        responses.add(
            responses.GET,
            f"http://localhost:8080/Automatizacion/{task_id}",
            json=mock_task_response,
            status=200
        )

        result = api_client.get_task(task_id)

        assert result["id"] == task_id
        assert result["name"] == "Test Task"
        assert result["state"] == "IN_PROGRESS"
        assert len(result["subtasks"]) == 1

    @responses.activate
    def test_get_task_not_found(self, api_client, mock_error_response):
        """Test get task with non-existent ID."""
        task_id = "550e8400-e29b-41d4-a716-446655440000"
        responses.add(
            responses.GET,
            f"http://localhost:8080/Automatizacion/{task_id}",
            json=mock_error_response,
            status=404
        )

        with pytest.raises(APIError) as exc_info:
            api_client.get_task(task_id)

        assert exc_info.value.status_code == 404
        assert exc_info.value.problem_details["title"] == "Task Not Found"

    @responses.activate
    def test_list_tasks_success(self, api_client, mock_task_list_response):
        """Test successful list tasks."""
        responses.add(
            responses.GET,
            "http://localhost:8080/AutomatizacionListado",
            json=mock_task_list_response,
            status=200
        )

        result = api_client.list_tasks()

        assert "tasks" in result
        assert "pagination" in result
        assert len(result["tasks"]) == 2
        assert result["pagination"]["total"] == 2

    @responses.activate
    def test_list_tasks_with_filters(self, api_client, mock_task_list_response):
        """Test list tasks with filters."""
        responses.add(
            responses.GET,
            "http://localhost:8080/AutomatizacionListado",
            json=mock_task_list_response,
            status=200
        )

        result = api_client.list_tasks(
            state="IN_PROGRESS",
            name="Test",
            page=2,
            limit=10
        )

        assert "tasks" in result
        # Verify request was made with correct params
        assert len(responses.calls) == 1
        request_params = responses.calls[0].request.params
        assert request_params["state"] == "IN_PROGRESS"
        assert request_params["name"] == "Test"
        assert request_params["page"] == "2"
        assert request_params["limit"] == "10"

    @responses.activate
    def test_list_tasks_empty(self, api_client):
        """Test list tasks with no results."""
        empty_response = {
            "tasks": [],
            "pagination": {
                "page": 1,
                "limit": 20,
                "total": 0,
                "total_pages": 0
            }
        }
        responses.add(
            responses.GET,
            "http://localhost:8080/AutomatizacionListado",
            json=empty_response,
            status=200
        )

        result = api_client.list_tasks()

        assert len(result["tasks"]) == 0
        assert result["pagination"]["total"] == 0

    @responses.activate
    def test_create_task_success(self, api_client, mock_task_response):
        """Test successful task creation."""
        responses.add(
            responses.POST,
            "http://localhost:8080/Automatizacion",
            json=mock_task_response,
            status=201
        )

        result = api_client.create_task(
            name="Test Task",
            created_by="Test Team",
            subtasks=[{"name": "Subtask 1"}]
        )

        assert result["name"] == "Test Task"
        assert result["created_by"] == "Test Team"

    @responses.activate
    def test_update_task_success(self, api_client, mock_task_response):
        """Test successful task update."""
        task_id = "550e8400-e29b-41d4-a716-446655440000"
        responses.add(
            responses.PUT,
            "http://localhost:8080/Automatizacion",
            json=mock_task_response,
            status=200
        )

        result = api_client.update_task(
            task_id=task_id,
            updated_by="Test Team",
            state="COMPLETED"
        )

        assert result["id"] == task_id

    @responses.activate
    def test_update_subtask_success(self, api_client):
        """Test successful subtask update."""
        subtask_id = "660e8400-e29b-41d4-a716-446655440001"
        subtask_response = {
            "id": subtask_id,
            "name": "Updated Subtask",
            "state": "COMPLETED",
            "created_at": "2025-11-27T10:00:00Z",
            "updated_at": "2025-11-27T12:00:00Z"
        }
        responses.add(
            responses.PUT,
            f"http://localhost:8080/Subtask/{subtask_id}",
            json=subtask_response,
            status=200
        )

        result = api_client.update_subtask(
            subtask_id=subtask_id,
            updated_by="Test Team",
            state="COMPLETED"
        )

        assert result["id"] == subtask_id
        assert result["state"] == "COMPLETED"

    @responses.activate
    def test_delete_subtask_success(self, api_client):
        """Test successful subtask deletion."""
        subtask_id = "660e8400-e29b-41d4-a716-446655440001"
        responses.add(
            responses.DELETE,
            f"http://localhost:8080/Subtask/{subtask_id}",
            status=204
        )

        result = api_client.delete_subtask(
            subtask_id=subtask_id,
            deleted_by="Test Team"
        )

        assert result is None  # 204 No Content returns None

    @responses.activate
    def test_api_connection_error(self, api_client):
        """Test API connection error handling."""
        responses.add(
            responses.GET,
            "http://localhost:8080/health",
            body=Exception("Connection refused")
        )

        with pytest.raises(Exception):
            api_client.health_check()
