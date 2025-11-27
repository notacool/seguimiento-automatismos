"""Utility functions for CLI."""
import json
from typing import Any, Dict, List
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich.json import JSON

console = Console()

STATES = ['PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED']

STATE_COLORS = {
    'PENDING': 'yellow',
    'IN_PROGRESS': 'blue',
    'COMPLETED': 'green',
    'FAILED': 'red',
    'CANCELLED': 'magenta'
}


def print_error(error_msg: str, problem_details: Dict[str, Any] = None):
    """Print error message in formatted style."""
    console.print(f"[bold red]Error:[/bold red] {error_msg}")

    if problem_details:
        console.print("\n[bold]Details:[/bold]")
        console.print(f"  Type: {problem_details.get('type', 'N/A')}")
        console.print(f"  Title: {problem_details.get('title', 'N/A')}")
        console.print(f"  Status: {problem_details.get('status', 'N/A')}")
        if 'detail' in problem_details:
            console.print(f"  Detail: {problem_details['detail']}")


def print_success(message: str):
    """Print success message."""
    console.print(f"[bold green]✓[/bold green] {message}")


def print_json(data: Any):
    """Print data as formatted JSON."""
    console.print(JSON(json.dumps(data, indent=2)))


def print_task(task: Dict[str, Any], show_subtasks: bool = True):
    """Print task details in a formatted panel."""
    state_color = STATE_COLORS.get(task['state'], 'white')

    content = f"""[bold]ID:[/bold] {task['id']}
[bold]Name:[/bold] {task['name']}
[bold]State:[/bold] [{state_color}]{task['state']}[/{state_color}]
[bold]Created by:[/bold] {task['created_by']}
[bold]Created at:[/bold] {task['created_at']}
[bold]Updated at:[/bold] {task['updated_at']}"""

    if task.get('updated_by'):
        content += f"\n[bold]Updated by:[/bold] {task['updated_by']}"
    if task.get('start_date'):
        content += f"\n[bold]Started:[/bold] {task['start_date']}"
    if task.get('end_date'):
        content += f"\n[bold]Ended:[/bold] {task['end_date']}"

    console.print(Panel(content, title=f"Task: {task['name']}", border_style=state_color))

    if show_subtasks and task.get('subtasks'):
        print_subtasks_table(task['subtasks'])


def print_subtask(subtask: Dict[str, Any]):
    """Print subtask details in a formatted panel."""
    state_color = STATE_COLORS.get(subtask['state'], 'white')

    content = f"""[bold]ID:[/bold] {subtask['id']}
[bold]Name:[/bold] {subtask['name']}
[bold]State:[/bold] [{state_color}]{subtask['state']}[/{state_color}]
[bold]Created at:[/bold] {subtask['created_at']}
[bold]Updated at:[/bold] {subtask['updated_at']}"""

    if subtask.get('start_date'):
        content += f"\n[bold]Started:[/bold] {subtask['start_date']}"
    if subtask.get('end_date'):
        content += f"\n[bold]Ended:[/bold] {subtask['end_date']}"

    console.print(Panel(content, title=f"Subtask: {subtask['name']}", border_style=state_color))


def print_subtasks_table(subtasks: List[Dict[str, Any]]):
    """Print subtasks in a table format."""
    if not subtasks:
        console.print("[dim]No subtasks[/dim]")
        return

    table = Table(title="Subtasks", show_header=True, header_style="bold magenta")
    table.add_column("Name", style="cyan", no_wrap=False)
    table.add_column("State", style="white")
    table.add_column("ID", style="dim")

    for subtask in subtasks:
        state_color = STATE_COLORS.get(subtask['state'], 'white')
        table.add_row(
            subtask['name'],
            f"[{state_color}]{subtask['state']}[/{state_color}]",
            subtask['id']
        )

    console.print(table)


def print_tasks_table(tasks: List[Dict[str, Any]], pagination: Dict[str, Any] = None):
    """Print tasks in a table format."""
    if not tasks:
        console.print("[yellow]No tasks found[/yellow]")
        return

    table = Table(show_header=True, header_style="bold magenta")
    table.add_column("Name", style="cyan", no_wrap=False)
    table.add_column("State", style="white")
    table.add_column("Subtasks", justify="right")
    table.add_column("Created", style="dim")
    table.add_column("ID", style="dim")

    for task in tasks:
        state_color = STATE_COLORS.get(task['state'], 'white')
        subtask_count = len(task.get('subtasks', []))

        table.add_row(
            task['name'],
            f"[{state_color}]{task['state']}[/{state_color}]",
            str(subtask_count),
            task['created_at'][:10],  # Show only date
            task['id'][:8]  # Show only first 8 chars of UUID
        )

    console.print(table)

    if pagination:
        console.print(f"\n[dim]Page {pagination['page']}/{pagination['total_pages']} "
                     f"| Total: {pagination['total']} tasks[/dim]")


def print_health_status(health_data: Dict[str, Any]):
    """Print health check status."""
    status = health_data.get('status', 'unknown')
    db_status = health_data.get('database', 'unknown')

    status_color = 'green' if status == 'healthy' else 'red'
    db_color = 'green' if db_status == 'ok' else 'red'

    content = f"""[bold]Service Status:[/bold] [{status_color}]{status.upper()}[/{status_color}]
[bold]Database:[/bold] [{db_color}]{db_status.upper()}[/{db_color}]
[bold]Timestamp:[/bold] {health_data.get('timestamp', 'N/A')}"""

    console.print(Panel(content, title="Health Check", border_style=status_color))
