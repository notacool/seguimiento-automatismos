# Desarrollo

Información sobre lenguajes, frameworks, arquitectura, patrones y flujo de trabajo utilizado en el proyecto.

## Tecnologías y Frameworks

- **Go 1.23+**: Lenguaje principal del backend
- **Gin**: Framework web HTTP para Go
- **PostgreSQL 16**: Base de datos relacional
- **pgx/v5**: Driver de PostgreSQL con connection pooling
- **golang-migrate**: Herramienta para migraciones de base de datos
- **golangci-lint**: Linter integral para Go
- **gofumpt**: Formateador estricto de código Go
- **Python 3.9+**: Para CLI de consultas (Click framework)

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

## Principios de Diseño

- **API-First**: Especificación OpenAPI antes de implementación
- **TDD**: Test-Driven Development en dominio y casos de uso
- **SOLID**: Principios de diseño orientado a objetos
- **KISS**: Keep It Simple, Stupid
- **YAGNI**: You Aren't Gonna Need It
- **DRY**: Don't Repeat Yourself

## Flujo de Trabajo Git

Este proyecto sigue un flujo de trabajo basado en **Git Flow** con las siguientes ramas:

### Ramas Principales

- **`main`**: Rama de producción. Solo contiene código estable y versionado.
- **`develop`**: Rama de desarrollo. Contiene el código en desarrollo y nuevas features.

### Flujo de Trabajo

1. **Features y mejoras**: Crear rama desde `develop`
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/nombre-feature
   ```

2. **Bugfixes críticos en producción**: Crear rama desde `main`
   ```bash
   git checkout main
   git pull origin main
   git checkout -b bugfix/descripcion-bug
   ```

3. **Pull Requests**:
   - Features y mejoras → PR contra `develop` (usar "squash and merge")
   - Bugfixes críticos → PR contra `main` (merge normal)
   - No mergear a `main` hasta que el Milestone esté finalizado

4. **Merge de `develop` a `main`**:
   - Solo cuando se complete un Milestone
   - Merge normal (sin squash)
   - Generar tag de versión (ej: `v1.1.0`)

### Convenciones de Nombres

- **Ramas**: `tipo/descripcion-corta`
  - Ejemplos: `feature/login-social`, `bugfix/error-login`, `refactor/cleanup-handlers`
- **Commits**: Mensajes descriptivos en presente
  - Ejemplos: `Agrega validación de usuario`, `Corrige error en login`
- **Tags**: Formato `vX.Y.Z` (Semantic Versioning)
  - Ejemplos: `v1.0.0`, `v1.1.0`, `v2.0.0`

## Pre-commit Hooks

Antes de cada commit, se ejecutan automáticamente las siguientes validaciones:

1. **Formateo de código** con `gofumpt`
2. **Ordenamiento de imports** con `goimports`
3. **Linter** con `golangci-lint`
4. **Tests unitarios**

### Instalación de Pre-commit Hooks

#### Linux/macOS

```bash
# Hacer el script ejecutable
chmod +x scripts/pre-commit.sh

# Instalar como hook de Git
cp scripts/pre-commit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

#### Windows (PowerShell)

```powershell
# Instalar como hook de Git
Copy-Item scripts/pre-commit.ps1 .git/hooks/pre-commit.ps1
```

O ejecutar manualmente antes de cada commit:

```bash
# Linux/macOS
./scripts/pre-commit.sh

# Windows (PowerShell)
.\scripts\pre-commit.ps1
```

### Ejecución Manual

Si prefieres ejecutar las validaciones manualmente antes de commitear:

```bash
# Formatear código
make fmt

# Verificar formato
make fmt-check

# Ejecutar linter
make lint

# Ejecutar tests unitarios
make test-unit
```

## Desarrollo desde Cero

### 1. Configurar el Entorno

```bash
# Clonar el repositorio
git clone <repository-url>
cd seguimiento-automatismos

# Instalar dependencias de Go
make deps

# Configurar variables de entorno
cp .env.example .env
# Editar .env con tus configuraciones
```

### 2. Levantar Servicios

```bash
# Levantar PostgreSQL con Docker/Podman
make docker-up

# Aplicar migraciones
make migrate-up
```

### 3. Desarrollo Local

```bash
# Ejecutar aplicación
make run

# Ejecutar tests
make test

# Ver cobertura
make test-coverage
```

### 4. Flujo de Desarrollo

1. Crear rama desde `develop`:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/mi-feature
   ```

2. Desarrollar y hacer commits:
   ```bash
   # Los pre-commit hooks se ejecutarán automáticamente
   git add .
   git commit -m "Agrega nueva funcionalidad"
   ```

3. Crear Pull Request:
   - Crear PR en modo draft
   - Vincular con Issue correspondiente
   - Nomenclatura: `[FEATURE] Descripción breve`

4. Merge a `develop`:
   - Usar "squash and merge" para features
   - Solo cuando el PR esté aprobado y los tests pasen

5. Release a `main`:
   - Cuando se complete un Milestone
   - Merge de `develop` a `main` (merge normal)
   - Crear tag de versión: `git tag -a v1.1.0 -m "Release v1.1.0"`
   - Push del tag: `git push origin v1.1.0`
   - El workflow automático actualizará las versiones y creará el release

## Versionado

El proyecto usa **Semantic Versioning** (MAJOR.MINOR.PATCH):

- **MAJOR**: Cambios incompatibles en la API
- **MINOR**: Nuevas funcionalidades compatibles hacia atrás
- **PATCH**: Correcciones de bugs compatibles hacia atrás

### Generación Automática de Versiones

Al crear un tag con formato `vX.Y.Z` o `X.Y.Z`, el workflow de GitHub Actions:

1. Extrae la versión del tag
2. Actualiza automáticamente:
   - `api/openapi/spec.yaml` → `info.version`
   - `api/openapi-generator-config.json` → `packageVersion`
   - `scripts/cli/main.py` → `version` en `@click.version_option`
3. Crea un GitHub Release con notas desde `CHANGELOG.md`

Ver [`CHANGELOG.md`](../CHANGELOG.md) para el historial completo de versiones.

## Testing

Ver [`README_TESTING.md`](../README_TESTING.md) para información detallada sobre testing.

### Tipos de Tests

- **Unitarios**: Tests de dominio y casos de uso (`make test-unit`)
- **Integración**: Tests con PostgreSQL real (`make test-integration`)
- **E2E**: Tests end-to-end con contenedores (`make test-e2e`)

## Linting y Formateo

### Herramientas

- **gofumpt**: Formateador estricto
- **goimports**: Gestor de imports
- **golangci-lint**: Linter integral con 25+ linters

### Comandos

```bash
# Formatear código
make fmt

# Verificar formato sin modificar
make fmt-check

# Ejecutar linter
make lint
```

## Migraciones de Base de Datos

```bash
# Aplicar migraciones
make migrate-up

# Revertir última migración
make migrate-down

# Crear nueva migración
make migrate-create NAME=nombre_migracion
```

Las migraciones se encuentran en `internal/adapter/repository/postgres/migrations/`.

## Generación de Código

### Generar Servidor Go desde OpenAPI

```bash
make generate-server
```

### Generar Cliente Python desde OpenAPI

```bash
make generate-client
```

## Recursos Adicionales

- [Guía de Desarrollo Completa](../Guia_desarrollo.md)
- [Documentación de Testing](../README_TESTING.md)
- [Documentación de API-First](../docs/API-FIRST.md)
- [Arquitectura del Proyecto](../docs/ARCHITECTURE_REVIEW.md)
