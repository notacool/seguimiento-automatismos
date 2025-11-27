"""Task query commands."""
import click
from api_client import APIClient, APIError
from utils import (
    print_task, print_tasks_table, print_json,
    print_error, STATES
)


@click.group()
def task():
    """Query tasks (automatizaciones)."""
    pass


@task.command()
@click.argument('task_id')
@click.option('--url', help='API base URL (overrides default)')
@click.option('--json', 'output_json', is_flag=True, help='Output as JSON')
def get(task_id, url, output_json):
    """Get task by ID.

    Example:
        automatizacion-cli task get 550e8400-e29b-41d4-a716-446655440000
    """
    try:
        client = APIClient(base_url=url) if url else APIClient()
        task_data = client.get_task(task_id)

        if output_json:
            print_json(task_data)
        else:
            print_task(task_data, show_subtasks=True)

    except APIError as e:
        print_error(f"API returned error (HTTP {e.status_code})", e.problem_details)
        raise click.Abort()
    except Exception as e:
        print_error(f"Failed to get task: {str(e)}")
        raise click.Abort()


@task.command()
@click.option('--state', type=click.Choice(STATES), help='Filter by state')
@click.option('--name', help='Filter by name (partial match, case-insensitive)')
@click.option('--page', type=int, default=1, help='Page number (default: 1)')
@click.option('--limit', type=int, default=20, help='Results per page (default: 20)')
@click.option('--url', help='API base URL (overrides default)')
@click.option('--json', 'output_json', is_flag=True, help='Output as JSON')
def list(state, name, page, limit, url, output_json):
    """List tasks with optional filters.

    Examples:
        automatizacion-cli task list
        automatizacion-cli task list --state IN_PROGRESS
        automatizacion-cli task list --name "Facturacion" --page 2
    """
    try:
        client = APIClient(base_url=url) if url else APIClient()
        response = client.list_tasks(
            state=state,
            name=name,
            page=page,
            limit=limit
        )

        if output_json:
            print_json(response)
        else:
            print_tasks_table(response['tasks'], response['pagination'])

    except APIError as e:
        print_error(f"API returned error (HTTP {e.status_code})", e.problem_details)
        raise click.Abort()
    except Exception as e:
        print_error(f"Failed to list tasks: {str(e)}")
        raise click.Abort()


@task.command()
@click.argument('task_id')
@click.option('--url', help='API base URL (overrides default)')
def subtasks(task_id, url):
    """Show subtasks for a specific task.

    Example:
        automatizacion-cli task subtasks 550e8400-e29b-41d4-a716-446655440000
    """
    try:
        client = APIClient(base_url=url) if url else APIClient()
        task_data = client.get_task(task_id)

        from utils import print_subtasks_table
        click.echo(f"\nSubtasks for: {task_data['name']}")
        print_subtasks_table(task_data.get('subtasks', []))

    except APIError as e:
        print_error(f"API returned error (HTTP {e.status_code})", e.problem_details)
        raise click.Abort()
    except Exception as e:
        print_error(f"Failed to get task subtasks: {str(e)}")
        raise click.Abort()
