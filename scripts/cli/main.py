#!/usr/bin/env python3
"""CLI for querying Automatizaciones API."""
import click
from commands.health import health
from commands.task import task


@click.group()
@click.version_option(version='1.0.0', prog_name='automatizacion-cli')
def cli():
    """CLI de consulta para la API de Automatizaciones.

    Este CLI permite consultar el estado de tareas y subtareas
    de automatizaciones gestionadas por múltiples equipos.

    Use --help en cada comando para más información.
    """
    pass


# Register commands
cli.add_command(health)
cli.add_command(task)


if __name__ == '__main__':
    cli()
