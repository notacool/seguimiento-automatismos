"""Tests for CLI commands."""
import pytest
import responses
from click.testing import CliRunner
from main import cli


class TestHealthCommand:
    """Test suite for health command."""

    @responses.activate
    def test_health_command_success(self, mock_health_response):
        """Test health command with healthy service."""
        responses.add(
            responses.GET,
            "http://localhost:8080/health",
            json=mock_health_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['health'])

        assert result.exit_code == 0
        assert "HEALTHY" in result.output
        assert "OK" in result.output

    @responses.activate
    def test_health_command_unhealthy(self):
        """Test health command with unhealthy service."""
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

        runner = CliRunner()
        result = runner.invoke(cli, ['health'])

        assert result.exit_code == 1
        assert "Error" in result.output

    @responses.activate
    def test_health_command_custom_url(self, mock_health_response):
        """Test health command with custom URL."""
        custom_url = "http://production:8080"
        responses.add(
            responses.GET,
            f"{custom_url}/health",
            json=mock_health_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['health', '--url', custom_url])

        assert result.exit_code == 0


class TestTaskGetCommand:
    """Test suite for task get command."""

    @responses.activate
    def test_task_get_success(self, mock_task_response):
        """Test get task command with valid ID."""
        task_id = "550e8400-e29b-41d4-a716-446655440000"
        responses.add(
            responses.GET,
            f"http://localhost:8080/Automatizacion/{task_id}",
            json=mock_task_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'get', task_id])

        assert result.exit_code == 0
        assert "Test Task" in result.output
        assert "IN_PROGRESS" in result.output

    @responses.activate
    def test_task_get_not_found(self, mock_error_response):
        """Test get task command with invalid ID."""
        task_id = "550e8400-e29b-41d4-a716-446655440000"
        responses.add(
            responses.GET,
            f"http://localhost:8080/Automatizacion/{task_id}",
            json=mock_error_response,
            status=404
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'get', task_id])

        assert result.exit_code == 1
        assert "Error" in result.output

    @responses.activate
    def test_task_get_json_output(self, mock_task_response):
        """Test get task command with JSON output."""
        task_id = "550e8400-e29b-41d4-a716-446655440000"
        responses.add(
            responses.GET,
            f"http://localhost:8080/Automatizacion/{task_id}",
            json=mock_task_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'get', task_id, '--json'])

        assert result.exit_code == 0
        # JSON output should contain the raw data
        assert task_id in result.output


class TestTaskListCommand:
    """Test suite for task list command."""

    @responses.activate
    def test_task_list_success(self, mock_task_list_response):
        """Test list tasks command."""
        responses.add(
            responses.GET,
            "http://localhost:8080/AutomatizacionListado",
            json=mock_task_list_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'list'])

        assert result.exit_code == 0
        assert "Task 1" in result.output
        assert "Task 2" in result.output

    @responses.activate
    def test_task_list_with_state_filter(self, mock_task_list_response):
        """Test list tasks with state filter."""
        responses.add(
            responses.GET,
            "http://localhost:8080/AutomatizacionListado",
            json=mock_task_list_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'list', '--state', 'IN_PROGRESS'])

        assert result.exit_code == 0

    @responses.activate
    def test_task_list_with_name_filter(self, mock_task_list_response):
        """Test list tasks with name filter."""
        responses.add(
            responses.GET,
            "http://localhost:8080/AutomatizacionListado",
            json=mock_task_list_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'list', '--name', 'Test'])

        assert result.exit_code == 0

    @responses.activate
    def test_task_list_with_pagination(self, mock_task_list_response):
        """Test list tasks with pagination."""
        responses.add(
            responses.GET,
            "http://localhost:8080/AutomatizacionListado",
            json=mock_task_list_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'list', '--page', '2', '--limit', '10'])

        assert result.exit_code == 0

    @responses.activate
    def test_task_list_empty(self):
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

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'list'])

        assert result.exit_code == 0
        assert "No tasks found" in result.output

    @responses.activate
    def test_task_list_json_output(self, mock_task_list_response):
        """Test list tasks with JSON output."""
        responses.add(
            responses.GET,
            "http://localhost:8080/AutomatizacionListado",
            json=mock_task_list_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'list', '--json'])

        assert result.exit_code == 0


class TestTaskSubtasksCommand:
    """Test suite for task subtasks command."""

    @responses.activate
    def test_task_subtasks_success(self, mock_task_response):
        """Test subtasks command."""
        task_id = "550e8400-e29b-41d4-a716-446655440000"
        responses.add(
            responses.GET,
            f"http://localhost:8080/Automatizacion/{task_id}",
            json=mock_task_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'subtasks', task_id])

        assert result.exit_code == 0
        assert "Subtask 1" in result.output

    @responses.activate
    def test_task_subtasks_empty(self):
        """Test subtasks command with task having no subtasks."""
        task_id = "550e8400-e29b-41d4-a716-446655440000"
        task_response = {
            "id": task_id,
            "name": "Task Without Subtasks",
            "state": "PENDING",
            "subtasks": [],
            "created_by": "Team",
            "created_at": "2025-11-27T10:00:00Z",
            "updated_at": "2025-11-27T10:00:00Z"
        }
        responses.add(
            responses.GET,
            f"http://localhost:8080/Automatizacion/{task_id}",
            json=task_response,
            status=200
        )

        runner = CliRunner()
        result = runner.invoke(cli, ['task', 'subtasks', task_id])

        assert result.exit_code == 0
        assert "No subtasks" in result.output
