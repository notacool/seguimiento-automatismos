package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

	// Setup router completo con todos los handlers
	router := httpHandler.SetupRouter(pg.Pool, gin.TestMode)

	t.Run("Health Check Works", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "ok", response["database"])
	})

	t.Run("Create Task Successfully", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"name":       "Tarea de Prueba E2E",
			"created_by": "test-user",
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
		assert.Equal(t, "Tarea de Prueba E2E", response["name"])
		assert.Equal(t, entity.StatePending.String(), response["state"])
	})

	t.Run("Get Task by ID", func(t *testing.T) {
		// Primero crear una tarea
		createBody := map[string]interface{}{
			"name":       "Tarea para Get",
			"created_by": "test-user",
		}
		createBytes, _ := json.Marshal(createBody)
		createReq := httptest.NewRequest(http.MethodPost, "/Automatizacion", bytes.NewBuffer(createBytes))
		createReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		router.ServeHTTP(createW, createReq)

		require.Equal(t, http.StatusCreated, createW.Code)
		var createdTask map[string]interface{}
		err := json.Unmarshal(createW.Body.Bytes(), &createdTask)
		require.NoError(t, err)
		taskID := createdTask["id"].(string)

		// Ahora obtener la tarea
		getReq := httptest.NewRequest(http.MethodGet, "/Automatizacion/"+taskID, nil)
		getW := httptest.NewRecorder()
		router.ServeHTTP(getW, getReq)

		assert.Equal(t, http.StatusOK, getW.Code)
		var getResponse map[string]interface{}
		err = json.Unmarshal(getW.Body.Bytes(), &getResponse)
		require.NoError(t, err)
		assert.Equal(t, taskID, getResponse["id"])
		assert.Equal(t, "Tarea para Get", getResponse["name"])
	})

	t.Run("Update Task State", func(t *testing.T) {
		// Crear tarea
		createBody := map[string]interface{}{
			"name":       "Tarea para Update",
			"created_by": "test-user",
		}
		createBytes, _ := json.Marshal(createBody)
		createReq := httptest.NewRequest(http.MethodPost, "/Automatizacion", bytes.NewBuffer(createBytes))
		createReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		router.ServeHTTP(createW, createReq)

		require.Equal(t, http.StatusCreated, createW.Code)
		var createdTask map[string]interface{}
		err := json.Unmarshal(createW.Body.Bytes(), &createdTask)
		require.NoError(t, err)
		taskID := createdTask["id"].(string)

		// Actualizar estado
		updateBody := map[string]interface{}{
			"id":         taskID,
			"state":      "IN_PROGRESS",
			"updated_by": "test-user",
		}
		updateBytes, _ := json.Marshal(updateBody)
		updateReq := httptest.NewRequest(http.MethodPut, "/Automatizacion", bytes.NewBuffer(updateBytes))
		updateReq.Header.Set("Content-Type", "application/json")
		updateW := httptest.NewRecorder()
		router.ServeHTTP(updateW, updateReq)

		assert.Equal(t, http.StatusOK, updateW.Code)
		var updateResponse map[string]interface{}
		err = json.Unmarshal(updateW.Body.Bytes(), &updateResponse)
		require.NoError(t, err)
		assert.Equal(t, "IN_PROGRESS", updateResponse["state"])
	})

	t.Run("List Tasks", func(t *testing.T) {
		// Crear algunas tareas
		for i := 0; i < 3; i++ {
			createBody := map[string]interface{}{
				"name":       "Tarea List " + string(rune('A'+i)),
				"created_by": "test-user",
			}
			createBytes, _ := json.Marshal(createBody)
			createReq := httptest.NewRequest(http.MethodPost, "/Automatizacion", bytes.NewBuffer(createBytes))
			createReq.Header.Set("Content-Type", "application/json")
			createW := httptest.NewRecorder()
			router.ServeHTTP(createW, createReq)
			require.Equal(t, http.StatusCreated, createW.Code)
		}

		// Listar tareas
		listReq := httptest.NewRequest(http.MethodGet, "/AutomatizacionListado?page=1&limit=10", nil)
		listW := httptest.NewRecorder()
		router.ServeHTTP(listW, listReq)

		assert.Equal(t, http.StatusOK, listW.Code)
		var listResponse map[string]interface{}
		err := json.Unmarshal(listW.Body.Bytes(), &listResponse)
		require.NoError(t, err)
		assert.NotNil(t, listResponse["tasks"])
		assert.NotNil(t, listResponse["pagination"])
	})
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

	// Setup router
	router := httpHandler.SetupRouter(pg.Pool, gin.TestMode)

	// Crear tarea vía API
	createBody := map[string]interface{}{
		"name":       "Test Task Transitions",
		"created_by": "test-user",
	}
	createBytes, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/Automatizacion", bytes.NewBuffer(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	require.Equal(t, http.StatusCreated, createW.Code)
	var createdTask map[string]interface{}
	err := json.Unmarshal(createW.Body.Bytes(), &createdTask)
	require.NoError(t, err)
	taskID := createdTask["id"].(string)
	assert.Equal(t, entity.StatePending.String(), createdTask["state"])

	// Transición PENDING -> IN_PROGRESS
	updateBody1 := map[string]interface{}{
		"id":         taskID,
		"state":      "IN_PROGRESS",
		"updated_by": "test-user",
	}
	updateBytes1, _ := json.Marshal(updateBody1)
	updateReq1 := httptest.NewRequest(http.MethodPut, "/Automatizacion", bytes.NewBuffer(updateBytes1))
	updateReq1.Header.Set("Content-Type", "application/json")
	updateW1 := httptest.NewRecorder()
	router.ServeHTTP(updateW1, updateReq1)

	assert.Equal(t, http.StatusOK, updateW1.Code)
	var updateResponse1 map[string]interface{}
	err = json.Unmarshal(updateW1.Body.Bytes(), &updateResponse1)
	require.NoError(t, err)
	assert.Equal(t, "IN_PROGRESS", updateResponse1["state"])

	// Transición IN_PROGRESS -> COMPLETED
	updateBody2 := map[string]interface{}{
		"id":         taskID,
		"state":      "COMPLETED",
		"updated_by": "test-user",
	}
	updateBytes2, _ := json.Marshal(updateBody2)
	updateReq2 := httptest.NewRequest(http.MethodPut, "/Automatizacion", bytes.NewBuffer(updateBytes2))
	updateReq2.Header.Set("Content-Type", "application/json")
	updateW2 := httptest.NewRecorder()
	router.ServeHTTP(updateW2, updateReq2)

	assert.Equal(t, http.StatusOK, updateW2.Code)
	var updateResponse2 map[string]interface{}
	err = json.Unmarshal(updateW2.Body.Bytes(), &updateResponse2)
	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", updateResponse2["state"])

	// Intentar transición inválida: COMPLETED -> PENDING (debe fallar)
	updateBody3 := map[string]interface{}{
		"id":         taskID,
		"state":      "PENDING",
		"updated_by": "test-user",
	}
	updateBytes3, _ := json.Marshal(updateBody3)
	updateReq3 := httptest.NewRequest(http.MethodPut, "/Automatizacion", bytes.NewBuffer(updateBytes3))
	updateReq3.Header.Set("Content-Type", "application/json")
	updateW3 := httptest.NewRecorder()
	router.ServeHTTP(updateW3, updateReq3)

	assert.Equal(t, http.StatusBadRequest, updateW3.Code)
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

	// Setup router
	router := httpHandler.SetupRouter(pg.Pool, gin.TestMode)

	// Crear tarea con subtareas vía API
	createBody := map[string]interface{}{
		"name":       "Parent Task",
		"created_by": "test-user",
		"subtasks": []map[string]interface{}{
			{"name": "Subtask 1"},
			{"name": "Subtask 2"},
		},
	}
	createBytes, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/Automatizacion", bytes.NewBuffer(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	require.Equal(t, http.StatusCreated, createW.Code)
	var createdTask map[string]interface{}
	err := json.Unmarshal(createW.Body.Bytes(), &createdTask)
	require.NoError(t, err)
	taskID := createdTask["id"].(string)

	// Verificar que la tarea tiene subtareas
	subtasks := createdTask["subtasks"].([]interface{})
	assert.Equal(t, 2, len(subtasks))

	// Obtener subtask ID para actualizar
	subtask1 := subtasks[0].(map[string]interface{})
	subtask1ID := subtask1["id"].(string)

	// Primero actualizar la tarea padre a IN_PROGRESS
	updateTaskBody1 := map[string]interface{}{
		"id":         taskID,
		"state":      "IN_PROGRESS",
		"updated_by": "test-user",
	}
	updateTaskBytes1, _ := json.Marshal(updateTaskBody1)
	updateTaskReq1 := httptest.NewRequest(http.MethodPut, "/Automatizacion", bytes.NewBuffer(updateTaskBytes1))
	updateTaskReq1.Header.Set("Content-Type", "application/json")
	updateTaskW1 := httptest.NewRecorder()
	router.ServeHTTP(updateTaskW1, updateTaskReq1)
	require.Equal(t, http.StatusOK, updateTaskW1.Code, "La tarea padre debe actualizarse a IN_PROGRESS")

	// Actualizar subtarea a IN_PROGRESS (las subtareas pueden estar en IN_PROGRESS cuando el padre está en IN_PROGRESS)
	updateSubtaskBody1 := map[string]interface{}{
		"state":      "IN_PROGRESS",
		"updated_by": "test-user",
	}
	updateSubtaskBytes1, _ := json.Marshal(updateSubtaskBody1)
	updateSubtaskReq1 := httptest.NewRequest(http.MethodPut, "/Subtask/"+subtask1ID, bytes.NewBuffer(updateSubtaskBytes1))
	updateSubtaskReq1.Header.Set("Content-Type", "application/json")
	updateSubtaskW1 := httptest.NewRecorder()
	router.ServeHTTP(updateSubtaskW1, updateSubtaskReq1)
	require.Equal(t, http.StatusOK, updateSubtaskW1.Code, "La subtarea debe actualizarse a IN_PROGRESS")

	// Ahora completar la tarea padre (esto propagará el estado a las subtareas)
	updateTaskBody2 := map[string]interface{}{
		"id":         taskID,
		"state":      "COMPLETED",
		"updated_by": "test-user",
	}
	updateSubtaskBytes, _ := json.Marshal(updateTaskBody2)
	updateSubtaskReq := httptest.NewRequest(http.MethodPut, "/Subtask/"+subtask1ID, bytes.NewBuffer(updateSubtaskBytes))
	updateSubtaskReq.Header.Set("Content-Type", "application/json")
	updateSubtaskW := httptest.NewRecorder()
	router.ServeHTTP(updateSubtaskW, updateSubtaskReq)

	assert.Equal(t, http.StatusOK, updateSubtaskW.Code)
	var updateSubtaskResponse map[string]interface{}
	err = json.Unmarshal(updateSubtaskW.Body.Bytes(), &updateSubtaskResponse)
	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", updateSubtaskResponse["state"])

	// Verificar que la tarea padre tiene la subtarea actualizada
	getReq := httptest.NewRequest(http.MethodGet, "/Automatizacion/"+taskID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusOK, getW.Code)
	var getResponse map[string]interface{}
	err = json.Unmarshal(getW.Body.Bytes(), &getResponse)
	require.NoError(t, err)
	updatedSubtasks := getResponse["subtasks"].([]interface{})
	found := false
	for _, st := range updatedSubtasks {
		subtask := st.(map[string]interface{})
		if subtask["id"] == subtask1ID {
			assert.Equal(t, "COMPLETED", subtask["state"])
			found = true
			break
		}
	}
	assert.True(t, found, "Subtask should be found in parent task")
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

	// Setup router
	router := httpHandler.SetupRouter(pg.Pool, gin.TestMode)

	// Crear múltiples tareas vía API
	for i := 0; i < 25; i++ {
		createBody := map[string]interface{}{
			"name":       "Task " + string(rune('A'+i)),
			"created_by": "test-user",
		}
		createBytes, _ := json.Marshal(createBody)
		createReq := httptest.NewRequest(http.MethodPost, "/Automatizacion", bytes.NewBuffer(createBytes))
		createReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		router.ServeHTTP(createW, createReq)
		require.Equal(t, http.StatusCreated, createW.Code)
	}

	// Test paginación - primera página
	listReq1 := httptest.NewRequest(http.MethodGet, "/AutomatizacionListado?page=1&limit=10", nil)
	listW1 := httptest.NewRecorder()
	router.ServeHTTP(listW1, listReq1)

	assert.Equal(t, http.StatusOK, listW1.Code)
	var listResponse1 map[string]interface{}
	err := json.Unmarshal(listW1.Body.Bytes(), &listResponse1)
	require.NoError(t, err)
	tasks1 := listResponse1["tasks"].([]interface{})
	pagination1 := listResponse1["pagination"].(map[string]interface{})
	assert.Equal(t, 10, len(tasks1))
	assert.Equal(t, float64(1), pagination1["page"])
	assert.Equal(t, float64(10), pagination1["limit"])
	assert.Equal(t, float64(25), pagination1["total"])

	// Test paginación - segunda página
	listReq2 := httptest.NewRequest(http.MethodGet, "/AutomatizacionListado?page=2&limit=10", nil)
	listW2 := httptest.NewRecorder()
	router.ServeHTTP(listW2, listReq2)

	assert.Equal(t, http.StatusOK, listW2.Code)
	var listResponse2 map[string]interface{}
	err = json.Unmarshal(listW2.Body.Bytes(), &listResponse2)
	require.NoError(t, err)
	tasks2 := listResponse2["tasks"].([]interface{})
	pagination2 := listResponse2["pagination"].(map[string]interface{})
	assert.Equal(t, 10, len(tasks2))
	assert.Equal(t, float64(2), pagination2["page"])

	// Test filtro por estado
	listReq3 := httptest.NewRequest(http.MethodGet, "/AutomatizacionListado?state=PENDING&page=1&limit=10", nil)
	listW3 := httptest.NewRecorder()
	router.ServeHTTP(listW3, listReq3)

	assert.Equal(t, http.StatusOK, listW3.Code)
	var listResponse3 map[string]interface{}
	err = json.Unmarshal(listW3.Body.Bytes(), &listResponse3)
	require.NoError(t, err)
	tasks3 := listResponse3["tasks"].([]interface{})
	for _, task := range tasks3 {
		taskMap := task.(map[string]interface{})
		assert.Equal(t, "PENDING", taskMap["state"])
	}
}
