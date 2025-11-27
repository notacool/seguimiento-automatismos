"""HTTP client for API interaction."""
import json
from typing import Any, Dict, Optional
import requests
from rich.console import Console

from config import API_BASE_URL, API_TIMEOUT

console = Console()


class APIError(Exception):
    """Custom exception for API errors."""
    def __init__(self, status_code: int, problem_details: Dict[str, Any]):
        self.status_code = status_code
        self.problem_details = problem_details
        super().__init__(problem_details.get('detail', 'Unknown API error'))


class APIClient:
    """Client for interacting with the automatizaciones API."""

    def __init__(self, base_url: str = API_BASE_URL, timeout: int = API_TIMEOUT):
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

    def _handle_response(self, response: requests.Response) -> Any:
        """Handle API response and errors."""
        if response.status_code == 204:
            return None

        try:
            data = response.json()
        except json.JSONDecodeError:
            data = {'detail': response.text or 'Empty response'}

        if not response.ok:
            raise APIError(response.status_code, data)

        return data

    def health_check(self) -> Dict[str, Any]:
        """Check API health status."""
        response = self.session.get(f'{self.base_url}/health', timeout=self.timeout)
        return self._handle_response(response)

    def create_task(self, name: str, created_by: str, state: Optional[str] = None,
                    subtasks: Optional[list] = None) -> Dict[str, Any]:
        """Create a new task."""
        payload = {
            'name': name,
            'created_by': created_by
        }
        if state:
            payload['state'] = state
        if subtasks:
            payload['subtasks'] = subtasks

        response = self.session.post(
            f'{self.base_url}/Automatizacion',
            json=payload,
            timeout=self.timeout
        )
        return self._handle_response(response)

    def update_task(self, task_id: str, updated_by: str, name: Optional[str] = None,
                    state: Optional[str] = None, subtasks: Optional[list] = None) -> Dict[str, Any]:
        """Update an existing task."""
        payload = {
            'id': task_id,
            'updated_by': updated_by
        }
        if name:
            payload['name'] = name
        if state:
            payload['state'] = state
        if subtasks:
            payload['subtasks'] = subtasks

        response = self.session.put(
            f'{self.base_url}/Automatizacion',
            json=payload,
            timeout=self.timeout
        )
        return self._handle_response(response)

    def get_task(self, task_id: str) -> Dict[str, Any]:
        """Get task by ID."""
        response = self.session.get(
            f'{self.base_url}/Automatizacion/{task_id}',
            timeout=self.timeout
        )
        return self._handle_response(response)

    def list_tasks(self, state: Optional[str] = None, name: Optional[str] = None,
                   page: int = 1, limit: int = 20) -> Dict[str, Any]:
        """List tasks with optional filters."""
        params = {
            'page': page,
            'limit': limit
        }
        if state:
            params['state'] = state
        if name:
            params['name'] = name

        response = self.session.get(
            f'{self.base_url}/AutomatizacionListado',
            params=params,
            timeout=self.timeout
        )
        return self._handle_response(response)

    def update_subtask(self, subtask_id: str, updated_by: str, name: Optional[str] = None,
                       state: Optional[str] = None) -> Dict[str, Any]:
        """Update a subtask."""
        payload = {'updated_by': updated_by}
        if name:
            payload['name'] = name
        if state:
            payload['state'] = state

        response = self.session.put(
            f'{self.base_url}/Subtask/{subtask_id}',
            json=payload,
            timeout=self.timeout
        )
        return self._handle_response(response)

    def delete_subtask(self, subtask_id: str, deleted_by: str) -> None:
        """Delete a subtask (soft delete)."""
        payload = {'deleted_by': deleted_by}
        response = self.session.delete(
            f'{self.base_url}/Subtask/{subtask_id}',
            json=payload,
            timeout=self.timeout
        )
        return self._handle_response(response)
