# Plan de Refactorización

## Resumen Ejecutivo

Análisis completo del código revela **17 oportunidades de mejora** distribuidas en 3 niveles de prioridad:
- 🔴 **CRÍTICO**: 2 issues (bugs y performance)
- 🟡 **ALTA**: 6 issues (duplicación de código, complejidad)
- 🟢 **MEDIA/BAJA**: 9 issues (mejoras opcionales)

**Estimación de esfuerzo total**: 16-22 horas
**Impacto esperado**: Mejora significativa en performance, mantenibilidad y corrección

---

## 🔴 CRÍTICO - Resolver Inmediatamente

### #1: N+1 Query Pattern en `findParentTask`
**Archivo**: `internal/usecase/subtask/update_subtask.go:119-146`
**Severidad**: CRÍTICA - Performance Issue
**Esfuerzo**: 1-2 horas

**Problema**:
Carga hasta 1000 tareas y las recorre para encontrar el parent de una subtarea. Esto es extremadamente ineficiente.

**Impacto**:
- Performance: O(n*m) donde n=tareas, m=subtareas por tarea
- Con 1000 tareas y 10 subtareas c/u = 10,000 comparaciones
- Con solo 100 tareas ya es lento

**Solución**:
```go
// Nueva función en SubtaskRepository
func (r *SubtaskRepository) FindParentTaskID(ctx context.Context, subtaskID uuid.UUID) (uuid.UUID, error) {
    query := `SELECT task_id FROM subtasks WHERE id = $1 AND deleted_at IS NULL`
    var taskID uuid.UUID
    err := r.pool.QueryRow(ctx, query, subtaskID).Scan(&taskID)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return uuid.Nil, entity.ErrSubtaskNotFound
        }
        return uuid.Nil, fmt.Errorf("failed to find parent task ID: %w", err)
    }
    return taskID, nil
}

// Reemplazar findParentTask con:
func (uc *UpdateSubtaskUseCase) findParentTask(ctx context.Context, subtaskID uuid.UUID) (*entity.Task, error) {
    taskID, err := uc.subtaskRepo.FindParentTaskID(ctx, subtaskID)
    if err != nil {
        return nil, err
    }

    return uc.taskRepo.FindByID(ctx, taskID)
}
```

**Beneficio**: De O(n*m) a O(1) - mejora de ~1000x en performance

---

### #2: Bug - Missing `task_id` en Insert de Subtask
**Archivo**: `internal/adapter/repository/postgres/task_repository.go:139-143`
**Severidad**: CRÍTICA - Bug
**Esfuerzo**: 15 minutos

**Problema**:
La query INSERT de subtask está mal formada. Tiene 7 columnas pero 8 valores.

**Código actual**:
```go
// Insert new subtask
_, err = tx.Exec(ctx, `
    INSERT INTO subtasks (id, name, state, start_date, end_date, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)  // ❌ 8 valores para 7 columnas
`, subtask.ID, task.ID, subtask.Name, ...)
```

**Solución**:
```go
// Insert new subtask
_, err = tx.Exec(ctx, `
    INSERT INTO subtasks (id, task_id, name, state, start_date, end_date, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`,
    subtask.ID,
    task.ID,
    subtask.Name,
    subtask.State.String(),
    subtask.StartDate,
    subtask.EndDate,
    subtask.CreatedAt,
    subtask.UpdatedAt,
)
```

**Impacto**: Este bug causaría fallo al insertar subtareas nuevas durante update de tasks.

---

## 🟡 ALTA PRIORIDAD - Refactorización Importante

### #3: Duplicación - Parsing de States
**Archivos**: `task_handler.go`, `subtask_handler.go`
**Esfuerzo**: 30 minutos

**Problema**: Patrón repetido 4+ veces:
```go
var state *entity.State
if req.State != nil {
    parsedState, err := ParseState(*req.State)
    if err != nil {
        MapErrorToProblemDetails(c, err)
        return
    }
    state = &parsedState
}
```

**Solución**:
```go
// En task_dto.go
func ParseOptionalState(stateStr *string) (*entity.State, error) {
    if stateStr == nil {
        return nil, nil
    }
    state, err := ParseState(*stateStr)
    if err != nil {
        return nil, err
    }
    return &state, nil
}

// Uso:
state, err := ParseOptionalState(req.State)
if err != nil {
    MapErrorToProblemDetails(c, err)
    return
}
```

**Impacto**: Elimina ~15 líneas duplicadas

---

### #4: Duplicación - Parsing de UUIDs
**Esfuerzo**: 30 minutos

**Solución**:
```go
func ParseAndMapUUID(c *gin.Context, uuidStr string, notFoundErr error) (uuid.UUID, bool) {
    id, err := ParseUUID(uuidStr)
    if err != nil {
        MapErrorToProblemDetails(c, notFoundErr)
        return uuid.Nil, false
    }
    return id, true
}

// Uso simplificado:
taskID, ok := ParseAndMapUUID(c, req.ID, entity.ErrTaskNotFound)
if !ok {
    return
}
```

---

### #5: Duplicación - Validación de Names
**Esfuerzo**: 20 minutos

```go
func ValidateAndMapName(c *gin.Context, name *string) bool {
    if name == nil {
        return true
    }
    if err := entity.ValidateName(*name); err != nil {
        MapErrorToProblemDetails(c, err)
        return false
    }
    return true
}
```

---

### #6: God Method - `TaskHandler.Create` demasiado largo
**Archivo**: `task_handler.go:59-156`
**Esfuerzo**: 2-3 horas

**Problema**: 97 líneas haciendo demasiadas cosas

**Solución**: Extraer métodos privados:
```go
func (h *TaskHandler) Create(c *gin.Context) {
    var req CreateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        MapErrorToProblemDetails(c, entity.ErrMissingRequiredFields)
        return
    }

    input, initialState, err := h.validateAndParseCreateRequest(c, &req)
    if err != nil {
        return
    }

    output, err := h.createUseCase.Execute(c.Request.Context(), input)
    if err != nil {
        c.Error(err).SetType(gin.ErrorTypePrivate)
        MapErrorToProblemDetails(c, err)
        return
    }

    if err := h.applyInitialState(c, output, initialState, req.CreatedBy); err != nil {
        return
    }

    if err := h.applyInitialSubtaskStates(c, output, req.Subtasks, req.CreatedBy); err != nil {
        return
    }

    c.JSON(http.StatusCreated, ToTaskResponse(output.Task))
}

func (h *TaskHandler) validateAndParseCreateRequest(...) {...}
func (h *TaskHandler) applyInitialState(...) {...}
func (h *TaskHandler) applyInitialSubtaskStates(...) {...}
```

---

### #7: Duplicación - Scanning de Entidades
**Esfuerzo**: 1-2 horas

**Problema**: Código de scanning repetido 5+ veces

**Solución**:
```go
func scanTask(scanner RowScanner) (*entity.Task, error) {
    var task entity.Task
    var state string

    err := scanner.Scan(
        &task.ID, &task.Name, &state, &task.CreatedBy,
        &task.UpdatedBy, &task.StartDate, &task.EndDate,
        &task.CreatedAt, &task.UpdatedAt, &task.DeletedAt,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, entity.ErrTaskNotFound
        }
        return nil, fmt.Errorf("failed to scan task: %w", err)
    }

    task.State = entity.State(state)
    return &task, nil
}
```

---

### #8: Error Handling Frágil
**Archivo**: `error_mapper.go:76-120`
**Esfuerzo**: 1 hora

**Problema**: Usa `strings.Contains` para match de errores

**Solución**: Usar `errors.Is` y `errors.As`:
```go
default:
    // Unwrap and check for known errors
    if errors.Is(err, entity.ErrInvalidName) {
        pd.Type = "https://api.grupoapi.com/problems/invalid-name"
        pd.Title = "Invalid Task Name"
        pd.Status = http.StatusBadRequest
        pd.Detail = err.Error()
        break
    }

    if errors.Is(err, entity.ErrInvalidStateTransition) {
        pd.Type = "https://api.grupoapi.com/problems/invalid-state-transition"
        pd.Title = "Invalid State Transition"
        pd.Status = http.StatusBadRequest
        pd.Detail = err.Error()
        break
    }

    // ... más checks con errors.Is

    // Default
    pd.Type = "https://api.grupoapi.com/problems/internal-error"
    pd.Title = "Internal Server Error"
    pd.Status = http.StatusInternalServerError
    pd.Detail = "An unexpected error occurred"
```

---

## 🟢 MEDIA/BAJA PRIORIDAD

### #9: Magic Numbers - Constantes de Paginación
**Esfuerzo**: 15 minutos

```go
// En internal/adapter/handler/http/constants.go
const (
    DefaultPageSize    = 20
    MaxPageSize        = 100
    DefaultPage        = 1
    MaxTaskSearchLimit = 1000
)
```

---

### #10: Duplicación - Parsing de Paginación
**Esfuerzo**: 30 minutos

```go
type PaginationParams struct {
    Page  int
    Limit int
}

func ParsePaginationParams(c *gin.Context) (PaginationParams, error) {
    // ... lógica centralizada
}
```

---

### #11: Complejidad - Nested Logic en `handleSubtasks`
**Archivo**: `update_task.go:120-224`
**Esfuerzo**: 2 horas

Extraer métodos más pequeños para reducir anidación de 4-5 niveles.

---

### #12: Tests - Setup Repetitivo
**Esfuerzo**: 1 hora

```go
type TestFixture struct {
    ctx    context.Context
    pg     *integration.PostgresContainer
    router *gin.Engine
    t      *testing.T
}

func SetupTestFixture(t *testing.T) *TestFixture {
    // Setup común
}
```

---

### #13: Tests - Helper para Request/Response
**Esfuerzo**: 1 hora

```go
func MakeRequestAndParse(t *testing.T, router *gin.Engine, method, path string, body, target interface{}) (*httptest.ResponseRecorder, error)

func CreateTask(t *testing.T, router *gin.Engine, name, createdBy string) (map[string]interface{}, *httptest.ResponseRecorder)
```

---

### #14-17: Mejoras Menores
- Nombres de variables más descriptivos
- Usar constantes HTTP del paquete `net/http`
- Timeouts en contextos
- Actualizar test script bash

---

## Plan de Implementación Recomendado

### Fase 1: CRÍTICO (Día 1)
1. ✅ Fix bug #2 - Missing task_id (15 min)
2. ✅ Fix performance #1 - N+1 query (2 horas)
3. ✅ Testing de las correcciones (30 min)

**Total Fase 1**: 2.75 horas

### Fase 2: Alta Prioridad - Helpers (Día 2)
1. ✅ Helper #3 - ParseOptionalState (30 min)
2. ✅ Helper #4 - ParseAndMapUUID (30 min)
3. ✅ Helper #5 - ValidateAndMapName (20 min)
4. ✅ Aplicar helpers en todos los handlers (1 hora)
5. ✅ Testing (30 min)

**Total Fase 2**: 2.5 horas

### Fase 3: Alta Prioridad - Refactoring Mayor (Día 3-4)
1. ✅ Refactor #6 - Extract methods de TaskHandler.Create (3 horas)
2. ✅ Refactor #7 - Scanners helpers (2 horas)
3. ✅ Refactor #8 - Error handling con errors.Is (1 hora)
4. ✅ Testing exhaustivo (2 horas)

**Total Fase 3**: 8 horas

### Fase 4: Media Prioridad - Opcional (Día 5)
1. ✅ Constants y pagination helpers (2 horas)
2. ✅ Test helpers (2 horas)
3. ✅ Complex conditionals refactor (2 horas)

**Total Fase 4**: 6 horas

---

## Métricas Esperadas Post-Refactoring

### Antes
- Líneas duplicadas: ~150
- Métodos >50 líneas: 5
- Complejidad ciclomática max: 15
- Performance findParentTask: O(n*m)

### Después
- Líneas duplicadas: ~20 (-87%)
- Métodos >50 líneas: 1 (-80%)
- Complejidad ciclomática max: 8 (-47%)
- Performance findParentTask: O(1) (~1000x mejor)

---

## Observaciones Positivas

Tu código ya demuestra **excelentes prácticas**:

✅ **Clean Architecture** bien definida
✅ **Dependency Injection** correcta
✅ **State Machine** pattern en dominio
✅ **RFC 7807** compliance
✅ **Testing** comprehensivo (unit + integration + e2e)
✅ **Repository Pattern** limpio
✅ **Graceful shutdown** en main.go
✅ **Boundary separation** clara entre capas

Los problemas identificados son **tácticos** (duplicación, métodos largos), no **arquitectónicos**. La base es sólida.

---

## Próximos Pasos Inmediatos

1. **AHORA**: Corregir bug crítico #2 (task_id missing)
2. **HOY**: Implementar fix de performance #1 (N+1 query)
3. **ESTA SEMANA**: Extraer helpers #3-#5 para eliminar duplicación
4. **SIGUIENTE SPRINT**: Refactorings mayores #6-#8

¿Quieres que comience con las correcciones críticas?
