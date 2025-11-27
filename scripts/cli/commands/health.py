"""Health check command."""
import click
from api_client import APIClient, APIError
from utils import print_health_status, print_error


@click.command()
@click.option('--url', help='API base URL (overrides default)')
def health(url):
    """Check API health status."""
    try:
        client = APIClient(base_url=url) if url else APIClient()
        health_data = client.health_check()
        print_health_status(health_data)
    except APIError as e:
        print_error(f"API returned error (HTTP {e.status_code})", e.problem_details)
        raise click.Abort()
    except Exception as e:
        print_error(f"Failed to connect to API: {str(e)}")
        raise click.Abort()
