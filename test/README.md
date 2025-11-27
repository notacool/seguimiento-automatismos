# Tests

Este directorio contiene los tests de integración y End-to-End (E2E) del proyecto.

## Estructura

```
test/
├── e2e/              # Tests End-to-End (flujo completo de la API)
├── fixtures/         # Datos de prueba compartidos
├── helpers/          # Funciones auxiliares para tests
└── integration/      # Tests de integración (DB, servicios externos)
```

## Tipos de Tests

### Tests Unitarios

Ubicados junto al código fuente (`*_test.go`):

- `internal/domain/entity/*_test.go` - Tests de entidades
- `internal/domain/service/*_test.go` - Tests de servicios de dominio
- `internal/usecase/**/*_test.go` - Tests de casos de uso con mocks
- `internal/adapter/handler/http/*_test.go` - Tests de handlers HTTP

### Tests de Integración

Ubicados en `test/integration/`:

- Tests de repositorios con PostgreSQL real (testcontainers)
- Tests de conexión a base de datos
- Tests de migraciones

### Tests E2E

Ubicados en `test/e2e/`:

- Tests de flujos completos de la API
- Tests de endpoints REST
- Tests de casos de uso reales

## Ejecutar Tests

```bash
# Todos los tests
make test-all

# Solo tests unitarios
make test-unit

# Solo tests de integración
make test-integration

# Solo tests E2E
make test-e2e

# Con cobertura
make test-coverage
```

## Convenciones

- Usar `testify/assert` y `testify/require` para assertions
- Usar `testify/mock` para mocks en tests unitarios
- Usar `testcontainers-go` para tests de integración con PostgreSQL
- Nombrar tests descriptivamente: `TestFunction_Scenario_ExpectedBehavior`
- Limpiar recursos en `defer` o usando `t.Cleanup()`
- Usar table-driven tests para múltiples casos

## Fixtures

Los fixtures contienen datos de prueba reutilizables:

- `fixtures/tasks.json` - Tareas de ejemplo
- `fixtures/subtasks.json` - Subtareas de ejemplo
- `fixtures/migrations/` - Scripts SQL para setup de tests
