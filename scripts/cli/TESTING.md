# Testing del CLI

Guía completa para ejecutar y escribir tests del CLI de Automatizaciones.

## Configuración Inicial

### Instalar Dependencias de Testing

```bash
# Con venv activado
source venv/bin/activate
pip install -r requirements-dev.txt
```

O usando make:

```bash
make cli-install  # Ya instala las dependencias de desarrollo
```

## Ejecutar Tests

### Todos los Tests

```bash
# Con venv activado
pytest

# O usando make (sin activar venv)
make cli-test
```

### Tests con Cobertura

```bash
# Con venv activado
pytest --cov=. --cov-report=html --cov-report=term

# O usando make
make cli-test-coverage
```

El reporte HTML se genera en `htmlcov/index.html`

### Tests Específicos

```bash
# Test de un archivo específico
pytest tests/test_api_client.py

# Test de una clase específica
pytest tests/test_api_client.py::TestAPIClient

# Test de una función específica
pytest tests/test_api_client.py::TestAPIClient::test_health_check_success

# Tests que coincidan con un patrón
pytest -k "health"
```

### Opciones Útiles

```bash
# Verbose output
pytest -v

# Mostrar output de print()
pytest -s

# Detener en el primer error
pytest -x

# Ejecutar los últimos tests que fallaron
pytest --lf

# Ejecutar tests en paralelo (requiere pytest-xdist)
pip install pytest-xdist
pytest -n auto
```

## Estructura de Tests

```
tests/
├── __init__.py           # Paquete de tests
├── conftest.py          # Fixtures compartidas
├── test_api_client.py   # Tests del cliente HTTP
├── test_commands.py     # Tests de comandos CLI
├── test_config.py       # Tests de configuración
└── test_utils.py        # Tests de utilidades
```

## Fixtures Disponibles

### Fixtures de API

- `api_base_url`: URL base para tests
- `api_client`: Instancia del cliente API
- `mock_health_response`: Respuesta simulada de health check
- `mock_task_response`: Respuesta simulada de tarea
- `mock_task_list_response`: Respuesta simulada de listado
- `mock_error_response`: Respuesta simulada de error RFC 7807

### Uso de Fixtures

```python
def test_example(api_client, mock_task_response):
    # api_client ya está configurado
    # mock_task_response contiene datos de prueba
    pass
```

## Escribir Tests

### Test de API Client

```python
import responses
from api_client import APIClient

@responses.activate
def test_get_task(api_client):
    # Mock de la respuesta HTTP
    responses.add(
        responses.GET,
        "http://localhost:8080/Automatizacion/task-id",
        json={"id": "task-id", "name": "Test"},
        status=200
    )

    # Ejecutar función
    result = api_client.get_task("task-id")

    # Verificar resultado
    assert result["name"] == "Test"
```

### Test de Comando CLI

```python
from click.testing import CliRunner
from main import cli

def test_health_command():
    runner = CliRunner()
    result = runner.invoke(cli, ['health'])

    assert result.exit_code == 0
    assert "HEALTHY" in result.output
```

### Test de Utilidad

```python
def test_print_function(capsys):
    from utils import print_success

    print_success("Test message")

    captured = capsys.readouterr()
    assert "Test message" in captured.out
```

## Cobertura de Tests

### Ver Reporte de Cobertura

```bash
# Generar reporte HTML
make cli-test-coverage

# Abrir en navegador (Linux)
xdg-open scripts/cli/htmlcov/index.html

# Abrir en navegador (Mac)
open scripts/cli/htmlcov/index.html
```

### Objetivo de Cobertura

Objetivo: **>80% de cobertura**

Áreas críticas que deben tener 100% de cobertura:
- API client (`api_client.py`)
- Comandos CLI (`commands/*.py`)

Áreas que pueden tener menor cobertura:
- Funciones de formateo (`utils.py`)
- Configuración (`config.py`)

## Mocking y Responses

### Mocking HTTP Requests

Usamos `responses` para simular respuestas HTTP:

```python
import responses

@responses.activate
def test_api_call():
    responses.add(
        responses.GET,
        "http://api.example.com/endpoint",
        json={"key": "value"},
        status=200
    )

    # Tu código que hace la petición HTTP
```

### Mocking con pytest-mock

```python
def test_with_mock(mocker):
    # Mock de una función
    mock_func = mocker.patch('module.function')
    mock_func.return_value = "mocked value"

    # Código que usa la función
```

## CI/CD

Los tests se pueden integrar en CI/CD:

```yaml
# Ejemplo para GitHub Actions
- name: Run CLI Tests
  run: |
    cd scripts/cli
    source venv/bin/activate
    pip install -r requirements-dev.txt
    pytest --cov=. --cov-report=xml

- name: Upload Coverage
  uses: codecov/codecov-action@v3
  with:
    files: scripts/cli/coverage.xml
```

## Troubleshooting

### Error: "Module not found"

Asegúrate de:
1. Estar en el directorio `scripts/cli`
2. Tener el venv activado
3. Haber instalado `requirements-dev.txt`

```bash
cd scripts/cli
source venv/bin/activate
pip install -r requirements-dev.txt
```

### Error: "No module named pytest"

```bash
pip install pytest
```

### Tests fallan por timeout

Aumenta el timeout en los tests:

```python
def test_slow_operation():
    client = APIClient(timeout=60)  # 60 segundos
```

### Dependencias circulares

Asegúrate de que las importaciones en los tests sean absolutas:

```python
# Correcto
from api_client import APIClient

# Evitar
from ..api_client import APIClient
```

## Mejores Prácticas

1. **Nombrar tests descriptivamente**
   ```python
   def test_health_check_returns_200_when_service_is_healthy():
       pass
   ```

2. **Un assert por test (idealmente)**
   ```python
   def test_task_has_correct_name():
       assert task["name"] == "Expected Name"
   ```

3. **Usar fixtures para datos comunes**
   ```python
   @pytest.fixture
   def sample_task():
       return {"id": "123", "name": "Test"}
   ```

4. **Documentar tests complejos**
   ```python
   def test_complex_scenario():
       """
       Given: Una tarea en estado PENDING
       When: Se actualiza a IN_PROGRESS
       Then: La fecha de inicio se asigna automáticamente
       """
       pass
   ```

5. **Parametrizar tests similares**
   ```python
   @pytest.mark.parametrize("state,expected", [
       ("PENDING", "yellow"),
       ("COMPLETED", "green"),
       ("FAILED", "red"),
   ])
   def test_state_colors(state, expected):
       assert STATE_COLORS[state] == expected
   ```

## Comandos Make Disponibles

```bash
make cli-test              # Ejecutar todos los tests
make cli-test-coverage     # Tests con reporte de cobertura
make cli-test-watch        # Tests en modo watch (auto-reload)
```

## Referencias

- [Pytest Documentation](https://docs.pytest.org/)
- [Click Testing](https://click.palletsprojects.com/en/8.1.x/testing/)
- [Responses Library](https://github.com/getsentry/responses)
- [pytest-cov](https://pytest-cov.readthedocs.io/)
