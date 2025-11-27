package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grupoapi/proces-log/internal/domain/entity"
)

func setupErrorTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		// Este endpoint se usará para probar diferentes errores
		errorType := c.Query("error")
		var err error

		switch errorType {
		case "invalid-name":
			err = entity.ErrInvalidName
		case "invalid-state-transition":
			err = entity.ErrInvalidStateTransition
		case "inconsistent-state":
			err = entity.ErrInconsistentParentChildState
		case "missing-fields":
			err = entity.ErrMissingRequiredFields
		case "task-not-found":
			err = entity.ErrTaskNotFound
		case "subtask-not-found":
			err = entity.ErrSubtaskNotFound
		case "database-unavailable":
			err = entity.ErrDatabaseUnavailable
		case "database-error":
			err = entity.ErrDatabaseError
		default:
			err = entity.ErrTaskNotFound
		}

		MapErrorToProblemDetails(c, err)
	})
	return router
}

func TestErrorMapper_InvalidName(t *testing.T) {
	router := setupErrorTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/test?error=invalid-name", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/invalid-name", response.Type)
	assert.Equal(t, "Invalid Task Name", response.Title)
	assert.Equal(t, http.StatusBadRequest, response.Status)
}

func TestErrorMapper_InvalidStateTransition(t *testing.T) {
	router := setupErrorTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/test?error=invalid-state-transition", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/invalid-state-transition", response.Type)
	assert.Equal(t, "Invalid State Transition", response.Title)
}

func TestErrorMapper_InconsistentParentChildState(t *testing.T) {
	router := setupErrorTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/test?error=inconsistent-state", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/inconsistent-parent-child-state", response.Type)
	assert.Equal(t, "Inconsistent Parent-Child State", response.Title)
}

func TestErrorMapper_MissingRequiredFields(t *testing.T) {
	router := setupErrorTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/test?error=missing-fields", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/missing-required-fields", response.Type)
	assert.Equal(t, "Missing Required Fields", response.Title)
}

func TestErrorMapper_TaskNotFound(t *testing.T) {
	router := setupErrorTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/test?error=task-not-found", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/task-not-found", response.Type)
	assert.Equal(t, "Task Not Found", response.Title)
	assert.Equal(t, http.StatusNotFound, response.Status)
}

func TestErrorMapper_SubtaskNotFound(t *testing.T) {
	router := setupErrorTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/test?error=subtask-not-found", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/subtask-not-found", response.Type)
	assert.Equal(t, "Subtask Not Found", response.Title)
}

func TestErrorMapper_DatabaseUnavailable(t *testing.T) {
	router := setupErrorTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/test?error=database-unavailable", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/database-unavailable", response.Type)
	assert.Equal(t, "Database Unavailable", response.Title)
}

func TestErrorMapper_DatabaseError(t *testing.T) {
	router := setupErrorTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/test?error=database-error", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/database-error", response.Type)
	assert.Equal(t, "Database Error", response.Title)
}

func TestErrorMapper_GenericError(t *testing.T) {
	router := setupErrorTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/test?error=unknown", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code) // Default error mapping
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Type)
	assert.NotEmpty(t, response.Title)
}
