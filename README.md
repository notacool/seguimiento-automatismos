# API-LOG-AUTOS

API REST en Go para gestión centralizada de estados de automatizaciones de múltiples equipos.

## Arquitectura

Este proyecto sigue **Clean Architecture (Hexagonal/Ports & Adapters)** con las siguientes capas:

```
grupoapi-proces-log/
├── cmd/api/                    # Entry point de la aplicación
├── internal/
│   ├── domain/                 # Capa de dominio (entidades, lógica de negocio)
│   │   ├── entity/            # Entidades: Task, Subtask
│   │   ├── service/           # Servicios de dominio: StateMachine
│   │   └── repository/        # Interfaces (ports)
│   ├── usecase/               # Casos de uso (application layer)
│   │   ├── task/             # Casos de uso de tareas
│   │   └── subtask/          # Casos de uso de subtareas
│   ├── adapter/               # Adaptadores (implementaciones)
│   │   ├── handler/          # HTTP handlers (Gin)
│   │   └── repository/       # Implementaciones de repositorio (PostgreSQL)
│   └── infrastructure/        # Configuración e infraestructura
│       ├── config/           # Configuración
│       └── database/         # Conexión a BD
├── api/openapi/               # Especificación OpenAPI 3.0
├── scripts/cli/               # CLI Python
├── deployments/docker/        # Docker & Docker Compose
└── test/                      # Tests de integración
```

## Tecnologías

- **Go 1.21+** con framework **Gin**
- **PostgreSQL 16** con extensión **pg_cron**
- **Docker & Docker Compose**
- **Python CLI** (Click) con binarios para Windows/Linux
- **OpenAPI 3.0** para especificación API-First
- **golang-migrate** para migraciones de base de datos

## Principios Aplicados

- **API-First**: Especificación OpenAPI antes de implementación
- **TDD**: Test-Driven Development en dominio y casos de uso
- **SOLID**: Principios de diseño orientado a objetos
- **KISS**: Keep It Simple, Stupid
- **YAGNI**: You Aren't Gonna Need It
- **DRY**: Don't Repeat Yourself

## Requisitos

- Go 1.21 o superior
- Docker & Docker Compose
- Make (para Linux/macOS) o usar comandos manualmente en Windows
- Python 3.9+ (para CLI)

## Configuración Rápida

1. **Clonar repositorio**

```bash
git clone <repository-url>
cd grupoapi-proces-log
```

2. **Configurar variables de entorno**

```bash
cp .env.example .env
# Editar .env con tus configuraciones
```

3. **Levantar servicios con Docker Compose**

```bash
make docker-up
# o manualmente: docker-compose -f deployments/docker/docker-compose.yml up -d
```

4. **Verificar salud del sistema**

```bash
curl http://localhost:8080/health
```

## Desarrollo

### Compilar y ejecutar localmente

```bash
# Descargar dependencias
make deps

# Compilar
make build

# Ejecutar tests
make test

# Ver cobertura
make test-coverage

# Ejecutar aplicación
make run
```

### Migraciones de base de datos

```bash
# Aplicar migraciones
make migrate-up

# Revertir migraciones
make migrate-down

# Crear nueva migración
make migrate-create NAME=create_users_table
```

### Linting y formato

```bash
# Formatear código
make fmt

# Verificar formato sin modificar
make fmt-check

# Ejecutar linter
make lint
```

**Herramientas instaladas:**

- `golangci-lint` v1.64+ - Linter integral con 25+ linters habilitados
- `gofumpt` v0.9+ - Formateador estricto
- `goimports` - Gestor de imports

Ver documentación completa en [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md)

## API Endpoints

### Health Check

- `GET /health` - Verificar estado del servicio y conexión a BD

### Automatizaciones (Tasks)

- `POST /Automatizacion` - Crear nueva tarea
- `PUT /Automatizacion` - Actualizar tarea (modificar, añadir/eliminar subtareas)
- `GET /Automatizacion/{uuid}` - Obtener tarea por ID
- `GET /AutomatizacionListado` - Listar tareas con filtros y paginación

### Subtareas

- `PUT /Subtask/{uuid}` - Actualizar subtarea individual
- `DELETE /Subtask/{uuid}` - Eliminar subtarea (soft delete)

Ver especificación completa en `api/openapi/spec.yaml`

## Estados de Tareas

- `PENDING` - Tarea pendiente de iniciar
- `IN_PROGRESS` - Tarea en ejecución
- `COMPLETED` - Tarea completada exitosamente
- `FAILED` - Tarea fallida
- `CANCELLED` - Tarea cancelada

### Reglas de Transición de Estado

- No se puede retroceder desde estados finales (COMPLETED, FAILED, CANCELLED)
- Las subtareas heredan automáticamente el estado final de la tarea padre
- Fecha de inicio se asigna al pasar a IN_PROGRESS
- Fecha de fin se asigna al llegar a estados finales

## CLI Python

### Instalación

```bash
cd scripts/cli
pip install -r requirements.txt
```

### Uso

```bash
# Crear tarea
python main.py create --name "Proceso Facturación" --created-by "Equipo Finanzas"

# Listar tareas
python main.py list --state PENDING --page 1 --limit 20

# Obtener tarea
python main.py get <uuid>

# Actualizar tarea
python main.py update <uuid> --state IN_PROGRESS --updated-by "Equipo Finanzas"

# Eliminar tarea (soft delete)
python main.py delete <uuid> --deleted-by "Equipo Finanzas"
```

### Generar binarios

```bash
make cli-build-windows  # Genera ejecutable para Windows
make cli-build-linux    # Genera ejecutable para Linux
```

## Testing

```bash
# Ejecutar todos los tests
make test

# Tests con cobertura
make test-coverage

# Tests de un paquete específico
go test -v ./internal/domain/entity/...
```

## Gestión de Errores (RFC 7807)

La API retorna errores siguiendo el estándar RFC 7807 (Problem Details):

```json
{
  "type": "https://api.example.com/problems/invalid-state-transition",
  "title": "Invalid State Transition",
  "status": 400,
  "detail": "Cannot transition from COMPLETED to PENDING"
}
```

Ver documento completo en `docs/RFC7807.md`

## Limpieza Automática

Las tareas eliminadas (soft delete) se borran permanentemente después de 30 días mediante un job automático de PostgreSQL (pg_cron).

## CI/CD

El proyecto incluye GitHub Actions para:

- Ejecutar tests automáticamente
- Compilar binarios
- Construir imágenes Docker

Ver `.github/workflows/ci.yml`

## Contribuir

1. Crear rama feature: `git checkout -b feature/nueva-funcionalidad`
2. Escribir tests primero (TDD)
3. Implementar funcionalidad
4. Ejecutar tests y linter: `make test lint`
5. Commit y push
6. Crear Pull Request

## Licencia

[Especificar licencia]
