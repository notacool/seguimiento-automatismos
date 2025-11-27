package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpHandler "github.com/grupoapi/proces-log/internal/adapter/handler/http"
	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/grupoapi/proces-log/test/integration"
)

func TestE2E_TaskCompleteLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Setup PostgreSQL container
	pg := integration.SetupPostgresContainer(ctx, t)
	defer pg.Teardown(ctx, t)

	// Create tables
	pg.CreateTasksTable(ctx, t)
	pg.CreateSubtasksTable(ctx, t)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup health handler
	healthHandler := httpHandler.NewHealthHandler(pg.Pool)
	router.GET("/health", healthHandler.Check)

	// TODO: Cuando se implementen los handlers de Task, agregar aquí
	// taskHandler := httpHandler.NewTaskHandler(taskUseCase)
	// router.POST("/Automatizacion", taskHandler.Create)
	// router.PUT("/Automatizacion", taskHandler.Update)
	// router.GET("/Automatizacion/:id", taskHandler.Get)
	// router.GET("/AutomatizacionListado", taskHandler.List)

	t.Run("Health Check Works", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "ok", response["status"])
		assert.Equal(t, "connected", response["database"])
	})

	// TODO: Descomentar cuando se implementen los handlers
	/*
		t.Run("Create Task Successfully", func(t *testing.T) {
			requestBody := map[string]interface{}{
				"nombre": "Tarea de Prueba E2E",
				"creado_por": "test-user",
			}

			bodyBytes, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPost, "/Automatizacion", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.NotEmpty(t, response["id"])
			assert.Equal(t, "Tarea de Prueba E2E", response["nombre"])
			assert.Equal(t, entity.StatePending.String(), response["estado"])
		})
	*/
}

func TestE2E_TaskStateTransitions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Setup PostgreSQL container
	pg := integration.SetupPostgresContainer(ctx, t)
	defer pg.Teardown(ctx, t)

	// Create tables
	pg.CreateTasksTable(ctx, t)
	pg.CreateSubtasksTable(ctx, t)

	// Insert test task directly
	taskID := uuid.New()
	insertSQL := `
		INSERT INTO tasks (id, name, state, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	_, err := pg.Pool.Exec(ctx, insertSQL, taskID, "Test Task", entity.StatePending.String(), "test-user", "test-user", now, now)
	require.NoError(t, err)

	// Verify task was inserted
	var count int
	err = pg.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM tasks WHERE id = $1", taskID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// TODO: Test state transitions when handlers are implemented
	t.Log("Task created successfully with ID:", taskID)
}

func TestE2E_TaskWithSubtasks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Setup PostgreSQL container
	pg := integration.SetupPostgresContainer(ctx, t)
	defer pg.Teardown(ctx, t)

	// Create tables
	pg.CreateTasksTable(ctx, t)
	pg.CreateSubtasksTable(ctx, t)

	// Insert test task
	taskID := uuid.New()
	now := time.Now()
	_, err := pg.Pool.Exec(ctx,
		`INSERT INTO tasks (id, name, state, created_by, updated_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		taskID, "Parent Task", entity.StatePending.String(), "test-user", "test-user", now, now)
	require.NoError(t, err)

	// Insert subtasks
	subtask1ID := uuid.New()
	_, err = pg.Pool.Exec(ctx,
		`INSERT INTO subtasks (id, task_id, name, state, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		subtask1ID, taskID, "Subtask 1", entity.StatePending.String(), "test-user", now, now)
	require.NoError(t, err)

	subtask2ID := uuid.New()
	_, err = pg.Pool.Exec(ctx,
		`INSERT INTO subtasks (id, task_id, name, state, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		subtask2ID, taskID, "Subtask 2", entity.StatePending.String(), "test-user", now, now)
	require.NoError(t, err)

	// Verify subtasks were inserted
	var subtaskCount int
	err = pg.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM subtasks WHERE task_id = $1", taskID).Scan(&subtaskCount)
	require.NoError(t, err)
	assert.Equal(t, 2, subtaskCount)

	// TODO: Test subtask operations when handlers are implemented
	t.Log("Task with subtasks created successfully")
}

func TestE2E_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Setup PostgreSQL container
	pg := integration.SetupPostgresContainer(ctx, t)
	defer pg.Teardown(ctx, t)

	// Create tables
	pg.CreateTasksTable(ctx, t)
	pg.CreateSubtasksTable(ctx, t)

	// Insert multiple tasks for pagination testing
	now := time.Now()
	for i := 0; i < 25; i++ {
		taskID := uuid.New()
		_, err := pg.Pool.Exec(ctx,
			`INSERT INTO tasks (id, name, state, created_by, updated_by, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			taskID, "Task "+string(rune(i)), entity.StatePending.String(), "test-user", "test-user", now, now)
		require.NoError(t, err)
	}

	// Verify tasks were inserted
	var taskCount int
	err := pg.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	require.NoError(t, err)
	assert.Equal(t, 25, taskCount)

	// TODO: Test pagination when list handler is implemented
	t.Log("Created 25 tasks for pagination testing")
}
