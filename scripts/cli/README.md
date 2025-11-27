# CLI de Consulta para Automatizaciones

CLI de Python para consultar el estado de automatizaciones y subtareas desde la API REST.

## Instalación

### Opción 1: Instalación Automática (Recomendada)

**Linux/Mac:**
```bash
cd scripts/cli
./setup.sh
```

**Windows:**
```cmd
cd scripts\cli
setup.bat
```

### Opción 2: Usando Makefile

```bash
# Crear venv e instalar dependencias
make cli-install

# O paso por paso:
make cli-setup     # Crear entorno virtual
make cli-deps      # Instalar dependencias (requiere venv activado)
```

### Opción 3: Instalación Manual

```bash
cd scripts/cli

# Crear entorno virtual
python3 -m venv venv

# Activar entorno virtual
source venv/bin/activate  # Linux/Mac
# o
venv\Scripts\activate.bat  # Windows

# Actualizar pip
pip install --upgrade pip

# Instalar dependencias
pip install -r requirements.txt
```

## Configuración

El CLI utiliza variables de entorno para la configuración. Crea un archivo `.env` en la raíz del proyecto o configura las siguientes variables:

```bash
API_BASE_URL=http://localhost:8080
API_TIMEOUT=30
```

Por defecto, el CLI se conecta a `http://localhost:8080`.

## Uso

### Activar el entorno virtual

Antes de usar el CLI, activa el entorno virtual:

```bash
# Linux/Mac
source scripts/cli/venv/bin/activate

# Windows
scripts\cli\venv\Scripts\activate.bat
```

### Ejecutar el CLI

```bash
# Desde el directorio scripts/cli (con venv activado)
python main.py --help

# O usando make desde la raíz del proyecto
make cli-run CMD="--help"
make cli-run CMD="health"
make cli-run CMD="task list"
```

### Desactivar el entorno virtual

Cuando termines de usar el CLI:

```bash
deactivate
```

### Comandos disponibles

#### Health Check

Verificar el estado del servicio y la conexión a base de datos:

```bash
python main.py health
python main.py health --url http://production-api:8080
```

#### Listar Tareas

Listar todas las tareas con paginación:

```bash
python main.py task list
```

Filtrar por estado:

```bash
python main.py task list --state IN_PROGRESS
python main.py task list --state COMPLETED
python main.py task list --state FAILED
```

Filtrar por nombre (búsqueda parcial, case-insensitive):

```bash
python main.py task list --name "Facturacion"
python main.py task list --name "Backup"
```

Paginación:

```bash
python main.py task list --page 2 --limit 50
```

Combinar filtros:

```bash
python main.py task list --state IN_PROGRESS --name "ETL" --page 1 --limit 10
```

Salida en JSON:

```bash
python main.py task list --json
```

#### Obtener Tarea por ID

Ver detalles completos de una tarea específica con sus subtareas:

```bash
python main.py task get 550e8400-e29b-41d4-a716-446655440000
python main.py task get 550e8400-e29b-41d4-a716-446655440000 --json
```

#### Ver Subtareas

Mostrar solo las subtareas de una tarea:

```bash
python main.py task subtasks 550e8400-e29b-41d4-a716-446655440000
```

### Opciones Globales

- `--help`: Mostrar ayuda para cualquier comando
- `--version`: Mostrar versión del CLI
- `--url`: Sobrescribir la URL base de la API

## Generar Ejecutables

Los ejecutables permiten distribuir el CLI sin necesidad de Python instalado.

### Windows

```bash
# Con venv activado
cd scripts/cli
source venv/bin/activate  # o venv\Scripts\activate.bat en Windows
pyinstaller --onefile --name automatizacion-cli-windows.exe main.py

# O usando make
make cli-build-windows
```

El ejecutable se genera en `scripts/cli/dist/automatizacion-cli-windows.exe`

### Linux

```bash
# Con venv activado
cd scripts/cli
source venv/bin/activate
pyinstaller --onefile --name automatizacion-cli-linux main.py

# O usando make
make cli-build-linux
```

El ejecutable se genera en `scripts/cli/dist/automatizacion-cli-linux`

**Nota:** PyInstaller debe estar instalado en el venv (incluido en requirements.txt)

## Ejemplos de Uso

### Monitorear tareas en progreso

```bash
python main.py task list --state IN_PROGRESS
```

### Buscar tareas de un equipo específico

```bash
python main.py task list --name "DevOps"
```

### Ver detalles completos de una tarea

```bash
# Primero listar para obtener el ID
python main.py task list --name "Backup"

# Luego obtener detalles
python main.py task get <task-id>
```

### Verificar estado del servicio antes de consultar

```bash
python main.py health && python main.py task list
```

## Salida

El CLI utiliza `rich` para formatear la salida:

- **Tablas**: Para listados de tareas y subtareas
- **Paneles**: Para detalles individuales de tareas
- **Colores**:
  - Verde: COMPLETED
  - Azul: IN_PROGRESS
  - Amarillo: PENDING
  - Rojo: FAILED
  - Magenta: CANCELLED

## Estados Válidos

- `PENDING`: Pendiente de iniciar
- `IN_PROGRESS`: En ejecución
- `COMPLETED`: Completada exitosamente (estado final)
- `FAILED`: Fallida (estado final)
- `CANCELLED`: Cancelada (estado final)

## Testing

El CLI incluye una suite completa de tests unitarios.

### Ejecutar Tests

```bash
# Todos los tests
make cli-test

# Tests con cobertura
make cli-test-coverage
```

### Cobertura Actual

- ✅ API Client: Tests completos con mocking HTTP
- ✅ Comandos CLI: Tests con Click testing utilities
- ✅ Utilidades: Tests de funciones de formateo
- ✅ Configuración: Tests de variables de entorno

Ver [TESTING.md](TESTING.md) para documentación completa de testing.

## Estructura del Proyecto

```
scripts/cli/
├── main.py                  # Punto de entrada del CLI
├── requirements.txt         # Dependencias producción
├── requirements-dev.txt     # Dependencias desarrollo/testing
├── config.py               # Configuración (URLs, timeouts)
├── api_client.py           # Cliente HTTP para la API
├── utils.py                # Utilidades de formateo
├── commands/               # Comandos del CLI
│   ├── __init__.py
│   ├── health.py           # Comando health check
│   └── task.py             # Comandos de consulta de tareas
├── tests/                  # Suite de tests
│   ├── __init__.py
│   ├── conftest.py         # Fixtures compartidas
│   ├── test_api_client.py  # Tests del cliente API
│   ├── test_commands.py    # Tests de comandos
│   ├── test_config.py      # Tests de configuración
│   └── test_utils.py       # Tests de utilidades
├── pytest.ini              # Configuración pytest
├── .coveragerc             # Configuración cobertura
├── setup.sh                # Script instalación Linux/Mac
├── setup.bat               # Script instalación Windows
├── README.md               # Esta documentación
├── QUICKSTART.md           # Guía rápida
└── TESTING.md              # Documentación de tests
```

## Manejo de Errores

El CLI maneja errores según RFC 7807 y muestra información detallada:

```bash
$ python main.py task get invalid-uuid
Error: API returned error (HTTP 404)

Details:
  Type: https://api.grupoapi.com/problems/task-not-found
  Title: Task Not Found
  Status: 404
  Detail: Task with ID 'invalid-uuid' not found or has been deleted
```

## Soporte

Para problemas o sugerencias, consulta la documentación principal del proyecto en `/docs`.
