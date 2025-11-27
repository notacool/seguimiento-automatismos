"""Tests for utility functions."""
import pytest
from utils import STATE_COLORS, STATES


class TestConstants:
    """Test suite for constants."""

    def test_states_list(self):
        """Test STATES constant contains all valid states."""
        expected_states = ['PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED']
        assert STATES == expected_states

    def test_state_colors_mapping(self):
        """Test STATE_COLORS contains all states."""
        assert 'PENDING' in STATE_COLORS
        assert 'IN_PROGRESS' in STATE_COLORS
        assert 'COMPLETED' in STATE_COLORS
        assert 'FAILED' in STATE_COLORS
        assert 'CANCELLED' in STATE_COLORS

    def test_state_colors_values(self):
        """Test STATE_COLORS has valid color values."""
        assert STATE_COLORS['PENDING'] == 'yellow'
        assert STATE_COLORS['IN_PROGRESS'] == 'blue'
        assert STATE_COLORS['COMPLETED'] == 'green'
        assert STATE_COLORS['FAILED'] == 'red'
        assert STATE_COLORS['CANCELLED'] == 'magenta'


class TestPrintFunctions:
    """Test suite for print utility functions."""

    def test_print_error_basic(self, capsys):
        """Test print_error with basic message."""
        from utils import print_error

        print_error("Test error message")
        captured = capsys.readouterr()
        assert "Error:" in captured.out
        assert "Test error message" in captured.out

    def test_print_error_with_details(self, capsys, mock_error_response):
        """Test print_error with problem details."""
        from utils import print_error

        print_error("API error", mock_error_response)
        captured = capsys.readouterr()
        assert "Error:" in captured.out
        assert "Details:" in captured.out
        assert "Task Not Found" in captured.out

    def test_print_success(self, capsys):
        """Test print_success function."""
        from utils import print_success

        print_success("Operation completed")
        captured = capsys.readouterr()
        assert "Operation completed" in captured.out

    def test_print_json(self, capsys):
        """Test print_json function."""
        from utils import print_json

        test_data = {"key": "value", "number": 123}
        print_json(test_data)
        captured = capsys.readouterr()
        # JSON output should contain the data
        assert "key" in captured.out
        assert "value" in captured.out

    def test_print_task(self, capsys, mock_task_response):
        """Test print_task function."""
        from utils import print_task

        print_task(mock_task_response, show_subtasks=True)
        captured = capsys.readouterr()
        assert "Test Task" in captured.out
        assert "IN_PROGRESS" in captured.out
        assert "Test Team" in captured.out

    def test_print_task_without_subtasks(self, capsys, mock_task_response):
        """Test print_task without showing subtasks."""
        from utils import print_task

        print_task(mock_task_response, show_subtasks=False)
        captured = capsys.readouterr()
        assert "Test Task" in captured.out

    def test_print_subtasks_table(self, capsys, mock_task_response):
        """Test print_subtasks_table function."""
        from utils import print_subtasks_table

        print_subtasks_table(mock_task_response["subtasks"])
        captured = capsys.readouterr()
        assert "Subtask 1" in captured.out

    def test_print_subtasks_table_empty(self, capsys):
        """Test print_subtasks_table with empty list."""
        from utils import print_subtasks_table

        print_subtasks_table([])
        captured = capsys.readouterr()
        assert "No subtasks" in captured.out

    def test_print_tasks_table(self, capsys, mock_task_list_response):
        """Test print_tasks_table function."""
        from utils import print_tasks_table

        print_tasks_table(
            mock_task_list_response["tasks"],
            mock_task_list_response["pagination"]
        )
        captured = capsys.readouterr()
        assert "Task 1" in captured.out
        assert "Task 2" in captured.out
        assert "Page 1/1" in captured.out

    def test_print_tasks_table_empty(self, capsys):
        """Test print_tasks_table with empty list."""
        from utils import print_tasks_table

        print_tasks_table([], None)
        captured = capsys.readouterr()
        assert "No tasks found" in captured.out

    def test_print_health_status(self, capsys, mock_health_response):
        """Test print_health_status function."""
        from utils import print_health_status

        print_health_status(mock_health_response)
        captured = capsys.readouterr()
        assert "HEALTHY" in captured.out
        assert "OK" in captured.out


class TestSubtask:
    """Test suite for subtask display functions."""

    def test_print_subtask(self, capsys):
        """Test print_subtask function."""
        from utils import print_subtask

        subtask = {
            "id": "660e8400-e29b-41d4-a716-446655440001",
            "name": "Test Subtask",
            "state": "COMPLETED",
            "created_at": "2025-11-27T10:00:00Z",
            "updated_at": "2025-11-27T10:15:00Z"
        }

        print_subtask(subtask)
        captured = capsys.readouterr()
        assert "Test Subtask" in captured.out
        assert "COMPLETED" in captured.out
