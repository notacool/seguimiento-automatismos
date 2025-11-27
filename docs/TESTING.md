# Guía de Testing

Esta guía describe la estrategia de testing del proyecto API-LOG-AUTOS.

## Pirámide de Testing

Seguimos la pirámide de testing con tres niveles:

```
        /\
       /  \  E2E Tests (Pocos, lentos, completos)
      /____\
     /      \  Integration Tests (Algunos, medios)
    /________\
   /          \  Unit Tests (Muchos, rápidos, aislados)
  /____________\
```

### 1. Tests Unitarios (70%)

**Ubicación:** Junto al código (`*_test.go`)

**Características:**

- Rápidos (< 10ms por test)
- Aislados (sin dependencias externas)
- Usan mocks para dependencias
- Alta cobertura de código

**Qué se testea:**

- Entidades de dominio (`internal/domain/entity/*_test.go`)
- Servicios de dominio (`internal/domain/service/*_test.go`)
- Casos de uso con repositorios mockeados
- Handlers HTTP con routers mockeados
- Validaciones y reglas de negocio

**Ejemplo:**

```go
func TestTask_UpdateState_ValidTransition(t *testing.T) {
    task, _ := entity.NewTask("Test", "user", nil)
    
    err := task.UpdateState(entity.StateInProgress)
    
    assert.NoError(t, err)
    assert.Equal(t, entity.StateInProgress, task.State)
}
```

### 2. Tests de Integración (20%)

**Ubicación:** `test/integration/`

**Características:**

- Medianamente rápidos (< 1s por test)
- Usan dependencias reales (PostgreSQL en contenedor)
- Verifican integración entre componentes
- Requieren Docker

**Qué se testea:**

- Repositorios con PostgreSQL real
- Migraciones de base de datos
- Transacciones y constraints
- Queries complejas

**Ejemplo:**

```go
func TestTaskRepository_Create(t *testing.T) {
    ctx := context.Background()
    pg := integration.SetupPostgresContainer(ctx, t)
    defer pg.Teardown(ctx, t)
    
    repo := postgres.NewTaskRepository(pg.Pool)
    task, _ := entity.NewTask("Test", "user", nil)
    
    err := repo.Create(ctx, task)
    
    require.NoError(t, err)
}
```

### 3. Tests E2E (10%)

**Ubicación:** `test/e2e/`

**Características:**

- Lentos (varios segundos por test)
- Testean flujos completos
- API real con base de datos real
- Simulan usuarios reales

**Qué se testea:**

- Flujos completos de usuario
- Endpoints REST end-to-end
- Casos de negocio complejos
- Integraciones entre módulos

**Ejemplo:**

```go
func TestE2E_TaskCompleteLifecycle(t *testing.T) {
    // 1. Crear tarea
    // 2. Añadir subtareas
    // 3. Actualizar estados
    // 4. Listar tareas
    // 5. Eliminar tarea
}
```

## Ejecutar Tests

### Tests Unitarios (por defecto)

```bash
make test
# o
make test-unit
# o
go test -short ./internal/...
```

### Tests de Integración

```bash
make test-integration
# o
go test -tags=integration ./test/integration/...
```

**Requisitos:** Docker Desktop corriendo

### Tests E2E

```bash
make test-e2e
# o
go test ./test/e2e/...
```

**Requisitos:** Docker Desktop corriendo

### Todos los Tests

```bash
make test-all
# o
go test ./...
```

### Cobertura

```bash
make test-coverage
# Abre navegador con reporte HTML

# O generar archivo HTML
make test-coverage-html
```

**Meta de cobertura:** ≥ 80%

## Convenciones de Naming

### Nombres de Tests

Usar formato: `TestFunction_Scenario_ExpectedBehavior`

**Ejemplos:**

```go
func TestTask_UpdateState_ValidTransition_Success(t *testing.T)
func TestTask_UpdateState_InvalidTransition_ReturnsError(t *testing.T)
func TestTaskRepository_Create_DuplicateID_ReturnsError(t *testing.T)
```

### Table-Driven Tests

Para múltiples casos similares:

```go
func TestState_IsValid(t *testing.T) {
    tests := []struct {
        name  string
        state entity.State
        want  bool
    }{
        {"PENDING is valid", entity.StatePending, true},
        {"INVALID is invalid", entity.State("INVALID"), false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.state.IsValid()
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Herramientas

### testify

**Instalación:**

```bash
go get github.com/stretchr/testify
```

**Uso:**

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
)

// assert - continúa test si falla
assert.Equal(t, expected, actual)

// require - detiene test si falla
require.NoError(t, err)

// mock - crear mocks
mockRepo := new(MockTaskRepository)
mockRepo.On("Create", ctx, task).Return(nil)
```

### testcontainers-go

**Uso:**

```go
pg := integration.SetupPostgresContainer(ctx, t)
defer pg.Teardown(ctx, t)

// Usar pg.Pool para queries
```

## Fixtures y Helpers

### Builders

Usar builders para crear entidades de test:

```go
task := helpers.NewTaskBuilder().
    WithName("Test Task").
    WithState(entity.StateInProgress).
    Build()
```

### Fixtures JSON

Datos compartidos en `test/fixtures/`:

- `tasks.json` - Tareas de ejemplo
- `subtasks.json` - Subtareas de ejemplo

### Test Helpers

Funciones auxiliares en `test/helpers/`:

- `MakeRequest()` - Hacer petición HTTP
- `AssertJSONResponse()` - Validar respuesta JSON
- `SetupTestGin()` - Configurar Gin para tests

## Best Practices

### 1. Tests Independientes

❌ **Mal:**

```go
var sharedTask *entity.Task // Estado compartido

func TestA(t *testing.T) {
    sharedTask = ...
}

func TestB(t *testing.T) {
    // Depende de TestA
}
```

✅ **Bien:**

```go
func TestA(t *testing.T) {
    task := createTestTask() // Independiente
}

func TestB(t *testing.T) {
    task := createTestTask() // Independiente
}
```

### 2. Cleanup Apropiado

✅ **Usar defer o t.Cleanup():**

```go
func TestWithResources(t *testing.T) {
    pg := setupPostgres(t)
    defer pg.Teardown(ctx, t)
    
    // O usar t.Cleanup()
    t.Cleanup(func() {
        pg.Teardown(ctx, t)
    })
}
```

### 3. Nombres Descriptivos

❌ **Mal:**

```go
func TestTask1(t *testing.T)
func TestTask2(t *testing.T)
```

✅ **Bien:**

```go
func TestTask_Create_ValidInput_Success(t *testing.T)
func TestTask_Create_InvalidName_ReturnsError(t *testing.T)
```

### 4. Assertions Claras

❌ **Mal:**

```go
if err != nil {
    t.Error("error")
}
```

✅ **Bien:**

```go
require.NoError(t, err, "should create task successfully")
assert.Equal(t, expected, actual, "task state should be PENDING")
```

### 5. Mock Solo lo Necesario

❌ **Mal:**

```go
// Mockear todo
mockDB := new(MockDB)
mockLogger := new(MockLogger)
mockCache := new(MockCache)
```

✅ **Bien:**

```go
// Solo mockear dependencias del unit bajo test
mockRepo := new(MockTaskRepository)
```

## Debugging Tests

### Ejecutar un test específico

```bash
go test -v -run TestTask_UpdateState ./internal/domain/entity/
```

### Ver logs detallados

```bash
go test -v -race ./... 2>&1 | tee test.log
```

### Usar testcontainers logs

```go
pg := SetupPostgresContainer(ctx, t)
logs, _ := pg.Container.Logs(ctx)
// Leer logs del contenedor
```

## CI/CD Integration

GitHub Actions ejecuta automáticamente:

```yaml
- name: Run Unit Tests
  run: make test-unit

- name: Run Integration Tests
  run: make test-integration

- name: Run E2E Tests
  run: make test-e2e

- name: Upload Coverage
  uses: codecov/codecov-action@v3
```

## Métricas de Calidad

### Cobertura Actual

- **Entidades:** 98.6%
- **Servicios:** 93.9%
- **Meta General:** ≥ 80%

### Velocidad

- **Unit tests:** < 2s total
- **Integration tests:** < 30s total
- **E2E tests:** < 2min total

## Referencias

- [Testing in Go](https://go.dev/doc/tutorial/add-a-test)
- [testify Documentation](https://github.com/stretchr/testify)
- [testcontainers-go](https://golang.testcontainers.org/)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
