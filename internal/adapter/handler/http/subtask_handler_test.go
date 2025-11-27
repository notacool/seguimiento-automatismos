package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	subtaskUsecase "github.com/grupoapi/proces-log/internal/usecase/subtask"
)

// MockUpdateSubtaskUseCase es un mock del UpdateSubtaskUseCase
type MockUpdateSubtaskUseCase struct {
	mock.Mock
}

func (m *MockUpdateSubtaskUseCase) Execute(ctx context.Context, input subtaskUsecase.UpdateSubtaskInput) (*subtaskUsecase.UpdateSubtaskOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subtaskUsecase.UpdateSubtaskOutput), args.Error(1)
}

// MockDeleteSubtaskUseCase es un mock del DeleteSubtaskUseCase
type MockDeleteSubtaskUseCase struct {
	mock.Mock
}

func (m *MockDeleteSubtaskUseCase) Execute(ctx context.Context, input subtaskUsecase.DeleteSubtaskInput) (*subtaskUsecase.DeleteSubtaskOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subtaskUsecase.DeleteSubtaskOutput), args.Error(1)
}

func setupSubtaskTestRouter(handler *SubtaskHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/Subtask/:uuid", handler.Update)
	router.DELETE("/Subtask/:uuid", handler.Delete)
	return router
}

func TestSubtaskHandler_Update_Success(t *testing.T) {
	// Setup
	mockUpdate := new(MockUpdateSubtaskUseCase)
	mockDelete := new(MockDeleteSubtaskUseCase)

	handler := NewSubtaskHandler(mockUpdate, mockDelete)
	router := setupSubtaskTestRouter(handler)

	// Crear subtarea de prueba
	subtask, err := entity.NewSubtask("Test Subtask")
	require.NoError(t, err)
	subtaskID := subtask.ID
	subtask.State = entity.StateCompleted
	subtask.SetEndDate()

	// Configurar mock
	mockUpdate.On("Execute", mock.Anything, mock.MatchedBy(func(input subtaskUsecase.UpdateSubtaskInput) bool {
		return input.ID == subtaskID && input.State != nil && *input.State == entity.StateCompleted
	})).Return(&subtaskUsecase.UpdateSubtaskOutput{Subtask: subtask}, nil)

	// Request
	reqBody := UpdateSubtaskRequest{
		State:     stringPtr("COMPLETED"),
		UpdatedBy: "test-user",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/Subtask/"+subtaskID.String(), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response SubtaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, subtaskID.String(), response.ID)
	assert.Equal(t, entity.StateCompleted.String(), response.State)
	mockUpdate.AssertExpectations(t)
}

func TestSubtaskHandler_Update_InvalidUUID(t *testing.T) {
	// Setup
	mockUpdate := new(MockUpdateSubtaskUseCase)
	mockDelete := new(MockDeleteSubtaskUseCase)

	handler := NewSubtaskHandler(mockUpdate, mockDelete)
	router := setupSubtaskTestRouter(handler)

	// Los mocks no se usan en este test ya que falla antes de llamar al use case
	_ = mockUpdate
	_ = mockDelete

	// Request con UUID inválido
	reqBody := UpdateSubtaskRequest{
		State:     stringPtr("COMPLETED"),
		UpdatedBy: "test-user",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/Subtask/invalid-uuid", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/subtask-not-found", response.Type)
	mockUpdate.AssertNotCalled(t, "Execute")
}

func TestSubtaskHandler_Update_InvalidState(t *testing.T) {
	// Setup
	mockUpdate := new(MockUpdateSubtaskUseCase)
	mockDelete := new(MockDeleteSubtaskUseCase)

	handler := NewSubtaskHandler(mockUpdate, mockDelete)
	router := setupSubtaskTestRouter(handler)

	// Los mocks no se usan en este test ya que falla antes de llamar al use case
	_ = mockUpdate
	_ = mockDelete

	// Request con estado inválido
	reqBody := UpdateSubtaskRequest{
		State:     stringPtr("INVALID_STATE"),
		UpdatedBy: "test-user",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/Subtask/"+uuid.New().String(), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/invalid-state-transition", response.Type)
	mockUpdate.AssertNotCalled(t, "Execute")
}

func TestSubtaskHandler_Delete_Success(t *testing.T) {
	// Setup
	mockUpdate := new(MockUpdateSubtaskUseCase)
	mockDelete := new(MockDeleteSubtaskUseCase)

	handler := NewSubtaskHandler(mockUpdate, mockDelete)
	router := setupSubtaskTestRouter(handler)

	subtaskID := uuid.New()

	// Configurar mock
	mockDelete.On("Execute", mock.Anything, mock.MatchedBy(func(input subtaskUsecase.DeleteSubtaskInput) bool {
		return input.ID == subtaskID && input.DeletedBy == "test-user"
	})).Return(&subtaskUsecase.DeleteSubtaskOutput{Success: true}, nil)

	// Request
	reqBody := DeleteSubtaskRequest{
		DeletedBy: "test-user",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodDelete, "/Subtask/"+subtaskID.String(), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockDelete.AssertExpectations(t)
}

func TestSubtaskHandler_Delete_NotFound(t *testing.T) {
	// Setup
	mockUpdate := new(MockUpdateSubtaskUseCase)
	mockDelete := new(MockDeleteSubtaskUseCase)

	handler := NewSubtaskHandler(mockUpdate, mockDelete)
	router := setupSubtaskTestRouter(handler)

	subtaskID := uuid.New()

	// Configurar mock para retornar error
	mockDelete.On("Execute", mock.Anything, mock.Anything).Return(nil, entity.ErrSubtaskNotFound)

	// Request
	reqBody := DeleteSubtaskRequest{
		DeletedBy: "test-user",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodDelete, "/Subtask/"+subtaskID.String(), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/subtask-not-found", response.Type)
	mockDelete.AssertExpectations(t)
}

func TestSubtaskHandler_Delete_InvalidUUID(t *testing.T) {
	// Setup
	mockUpdate := new(MockUpdateSubtaskUseCase)
	mockDelete := new(MockDeleteSubtaskUseCase)

	handler := NewSubtaskHandler(mockUpdate, mockDelete)
	router := setupSubtaskTestRouter(handler)

	// Los mocks no se usan en este test ya que falla antes de llamar al use case
	_ = mockUpdate
	_ = mockDelete

	// Request con UUID inválido
	reqBody := DeleteSubtaskRequest{
		DeletedBy: "test-user",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodDelete, "/Subtask/invalid-uuid", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/subtask-not-found", response.Type)
	mockDelete.AssertNotCalled(t, "Execute")
}
