//go:build !container
// +build !container

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

// TestE2E_TaskCompleteLifecycle_Local ejecuta tests E2E usando PostgreSQL local
// en lugar de contenedores. Requiere PostgreSQL instalado localmente.
//
// Para ejecutar estos tests:
//   - Asegúrate de tener PostgreSQL instalado y corriendo
//   - Opcionalmente configura variables de entorno:
//     TEST_DB_HOST=localhost
//     TEST_DB_PORT=5432
//     TEST_DB_USER=postgres
//     TEST_DB_PASSWORD=postgres
//   - Ejecuta: go test -v -tags=!container ./test/e2e/...
func TestE2E_TaskCompleteLifecycle_Local(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Setup PostgreSQL local
	pg := integration.SetupPostgresLocal(ctx, t)
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
			"name":       "Tarea de Prueba E2E Local",
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
		assert.Equal(t, "Tarea de Prueba E2E Local", response["name"])
		assert.Equal(t, entity.StatePending.String(), response["state"])
	})

	t.Run("Get Task by ID", func(t *testing.T) {
		// Primero crear una tarea
		createBody := map[string]interface{}{
			"name":       "Tarea para Get Local",
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
		assert.Equal(t, "Tarea para Get Local", getResponse["name"])
	})

	t.Run("Update Task State", func(t *testing.T) {
		// Crear tarea
		createBody := map[string]interface{}{
			"name":       "Tarea para Update Local",
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
				"name":       "Tarea List Local " + string(rune('A'+i)),
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
