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
	taskUsecase "github.com/grupoapi/proces-log/internal/usecase/task"
)

// MockCreateTaskUseCase es un mock del CreateTaskUseCase
type MockCreateTaskUseCase struct {
	mock.Mock
}

func (m *MockCreateTaskUseCase) Execute(ctx context.Context, input taskUsecase.CreateTaskInput) (*taskUsecase.CreateTaskOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*taskUsecase.CreateTaskOutput), args.Error(1)
}

// MockGetTaskUseCase es un mock del GetTaskUseCase
type MockGetTaskUseCase struct {
	mock.Mock
}

func (m *MockGetTaskUseCase) Execute(ctx context.Context, input taskUsecase.GetTaskInput) (*taskUsecase.GetTaskOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*taskUsecase.GetTaskOutput), args.Error(1)
}

// MockListTasksUseCase es un mock del ListTasksUseCase
type MockListTasksUseCase struct {
	mock.Mock
}

func (m *MockListTasksUseCase) Execute(ctx context.Context, input taskUsecase.ListTasksInput) (*taskUsecase.ListTasksOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*taskUsecase.ListTasksOutput), args.Error(1)
}

// MockUpdateTaskUseCase es un mock del UpdateTaskUseCase
type MockUpdateTaskUseCase struct {
	mock.Mock
}

func (m *MockUpdateTaskUseCase) Execute(ctx context.Context, input taskUsecase.UpdateTaskInput) (*taskUsecase.UpdateTaskOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*taskUsecase.UpdateTaskOutput), args.Error(1)
}

func setupTestRouter(handler *TaskHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/Automatizacion", handler.Create)
	router.PUT("/Automatizacion", handler.Update)
	router.GET("/Automatizacion/:uuid", handler.Get)
	router.GET("/AutomatizacionListado", handler.List)
	return router
}

func TestTaskHandler_Create_Success(t *testing.T) {
	// Setup
	mockCreate := new(MockCreateTaskUseCase)
	mockGet := new(MockGetTaskUseCase)
	mockList := new(MockListTasksUseCase)
	mockUpdate := new(MockUpdateTaskUseCase)

	handler := NewTaskHandler(mockCreate, mockGet, mockList, mockUpdate)
	router := setupTestRouter(handler)

	// Crear tarea de prueba
	task, err := entity.NewTask("Test Task", "test-user")
	require.NoError(t, err)

	// Configurar mock
	mockCreate.On("Execute", mock.Anything, mock.MatchedBy(func(input taskUsecase.CreateTaskInput) bool {
		return input.Name == "Test Task" && input.CreatedBy == "test-user"
	})).Return(&taskUsecase.CreateTaskOutput{Task: task}, nil)

	// Request
	reqBody := CreateTaskRequest{
		Name:      "Test Task",
		CreatedBy: "test-user",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/Automatizacion", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	var response TaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Test Task", response.Name)
	assert.Equal(t, entity.StatePending.String(), response.State)
	mockCreate.AssertExpectations(t)
}

func TestTaskHandler_Create_InvalidName(t *testing.T) {
	// Setup
	mockCreate := new(MockCreateTaskUseCase)
	mockGet := new(MockGetTaskUseCase)
	mockList := new(MockListTasksUseCase)
	mockUpdate := new(MockUpdateTaskUseCase)

	handler := NewTaskHandler(mockCreate, mockGet, mockList, mockUpdate)
	router := setupTestRouter(handler)

	// Los mocks mockGet, mockList y mockUpdate no se usan en este test
	_ = mockGet
	_ = mockList
	_ = mockUpdate

	// Request con nombre inválido
	reqBody := CreateTaskRequest{
		Name:      "@@Invalid@@",
		CreatedBy: "test-user",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/Automatizacion", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/invalid-name", response.Type)
	mockCreate.AssertNotCalled(t, "Execute")
}

func TestTaskHandler_Create_WithSubtasks(t *testing.T) {
	// Setup
	mockCreate := new(MockCreateTaskUseCase)
	mockGet := new(MockGetTaskUseCase)
	mockList := new(MockListTasksUseCase)
	mockUpdate := new(MockUpdateTaskUseCase)

	handler := NewTaskHandler(mockCreate, mockGet, mockList, mockUpdate)
	router := setupTestRouter(handler)

	// Los mocks mockGet, mockList y mockUpdate no se usan en este test
	_ = mockGet
	_ = mockList
	_ = mockUpdate

	// Crear tarea con subtareas
	task, err := entity.NewTask("Test Task", "test-user")
	require.NoError(t, err)
	subtask1, _ := entity.NewSubtask("Subtask 1")
	subtask2, _ := entity.NewSubtask("Subtask 2")
	task.AddSubtask(subtask1)
	task.AddSubtask(subtask2)

	// Configurar mock
	mockCreate.On("Execute", mock.Anything, mock.MatchedBy(func(input taskUsecase.CreateTaskInput) bool {
		return len(input.SubtaskNames) == 2
	})).Return(&taskUsecase.CreateTaskOutput{Task: task}, nil)

	// Request
	reqBody := CreateTaskRequest{
		Name:      "Test Task",
		CreatedBy: "test-user",
		Subtasks: []CreateSubtaskRequest{
			{Name: "Subtask 1"},
			{Name: "Subtask 2"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/Automatizacion", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	var response TaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 2, len(response.Subtasks))
	mockCreate.AssertExpectations(t)
}

func TestTaskHandler_Get_Success(t *testing.T) {
	// Setup
	mockCreate := new(MockCreateTaskUseCase)
	mockGet := new(MockGetTaskUseCase)
	mockList := new(MockListTasksUseCase)
	mockUpdate := new(MockUpdateTaskUseCase)

	handler := NewTaskHandler(mockCreate, mockGet, mockList, mockUpdate)
	router := setupTestRouter(handler)

	// Los mocks mockCreate, mockList y mockUpdate no se usan en este test
	_ = mockCreate
	_ = mockList
	_ = mockUpdate

	// Crear tarea de prueba
	task, err := entity.NewTask("Test Task", "test-user")
	require.NoError(t, err)
	taskID := task.ID

	// Configurar mock
	mockGet.On("Execute", mock.Anything, mock.MatchedBy(func(input taskUsecase.GetTaskInput) bool {
		return input.ID == taskID
	})).Return(&taskUsecase.GetTaskOutput{Task: task}, nil)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/Automatizacion/"+taskID.String(), nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response TaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, taskID.String(), response.ID)
	assert.Equal(t, "Test Task", response.Name)
	mockGet.AssertExpectations(t)
}

func TestTaskHandler_Get_InvalidUUID(t *testing.T) {
	// Setup
	mockCreate := new(MockCreateTaskUseCase)
	mockGet := new(MockGetTaskUseCase)
	mockList := new(MockListTasksUseCase)
	mockUpdate := new(MockUpdateTaskUseCase)

	handler := NewTaskHandler(mockCreate, mockGet, mockList, mockUpdate)
	router := setupTestRouter(handler)

	// Los mocks no se usan en este test ya que falla antes de llamar al use case
	_ = mockCreate
	_ = mockGet
	_ = mockList
	_ = mockUpdate

	// Request con UUID inválido
	req := httptest.NewRequest(http.MethodGet, "/Automatizacion/invalid-uuid", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/task-not-found", response.Type)
	mockGet.AssertNotCalled(t, "Execute")
}

func TestTaskHandler_List_Success(t *testing.T) {
	// Setup
	mockCreate := new(MockCreateTaskUseCase)
	mockGet := new(MockGetTaskUseCase)
	mockList := new(MockListTasksUseCase)
	mockUpdate := new(MockUpdateTaskUseCase)

	handler := NewTaskHandler(mockCreate, mockGet, mockList, mockUpdate)
	router := setupTestRouter(handler)

	// Los mocks mockCreate, mockGet y mockUpdate no se usan en este test
	_ = mockCreate
	_ = mockGet
	_ = mockUpdate

	// Crear tareas de prueba
	task1, _ := entity.NewTask("Task 1", "user1")
	task2, _ := entity.NewTask("Task 2", "user2")

	// Configurar mock
	mockList.On("Execute", mock.Anything, mock.MatchedBy(func(input taskUsecase.ListTasksInput) bool {
		return input.Page == 1 && input.Limit == 20
	})).Return(&taskUsecase.ListTasksOutput{
		Tasks:      []*entity.Task{task1, task2},
		Total:      2,
		Page:       1,
		Limit:      20,
		TotalPages: 1,
	}, nil)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/AutomatizacionListado?page=1&limit=20", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response TaskListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 2, len(response.Tasks))
	assert.Equal(t, 2, response.Pagination.Total)
	mockList.AssertExpectations(t)
}

func TestTaskHandler_List_WithFilters(t *testing.T) {
	// Setup
	mockCreate := new(MockCreateTaskUseCase)
	mockGet := new(MockGetTaskUseCase)
	mockList := new(MockListTasksUseCase)
	mockUpdate := new(MockUpdateTaskUseCase)

	handler := NewTaskHandler(mockCreate, mockGet, mockList, mockUpdate)
	router := setupTestRouter(handler)

	// Los mocks mockCreate, mockGet y mockUpdate no se usan en este test
	_ = mockCreate
	_ = mockGet
	_ = mockUpdate

	// Crear tarea de prueba
	task, _ := entity.NewTask("Test Task", "user")
	task.UpdateState(entity.StateInProgress, "user")

	// Configurar mock
	mockList.On("Execute", mock.Anything, mock.MatchedBy(func(input taskUsecase.ListTasksInput) bool {
		return input.State != nil && *input.State == entity.StateInProgress
	})).Return(&taskUsecase.ListTasksOutput{
		Tasks:      []*entity.Task{task},
		Total:      1,
		Page:       1,
		Limit:      20,
		TotalPages: 1,
	}, nil)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/AutomatizacionListado?state=IN_PROGRESS", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response TaskListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 1, len(response.Tasks))
	assert.Equal(t, entity.StateInProgress.String(), response.Tasks[0].State)
	mockList.AssertExpectations(t)
}

func TestTaskHandler_Update_Success(t *testing.T) {
	// Setup
	mockCreate := new(MockCreateTaskUseCase)
	mockGet := new(MockGetTaskUseCase)
	mockList := new(MockListTasksUseCase)
	mockUpdate := new(MockUpdateTaskUseCase)

	handler := NewTaskHandler(mockCreate, mockGet, mockList, mockUpdate)
	router := setupTestRouter(handler)

	// Los mocks mockCreate, mockGet y mockList no se usan en este test
	_ = mockCreate
	_ = mockGet
	_ = mockList

	// Crear tarea de prueba
	task, err := entity.NewTask("Test Task", "test-user")
	require.NoError(t, err)
	taskID := task.ID
	task.UpdateState(entity.StateInProgress, "test-user")

	// Configurar mock
	mockUpdate.On("Execute", mock.Anything, mock.MatchedBy(func(input taskUsecase.UpdateTaskInput) bool {
		return input.ID == taskID && input.State != nil && *input.State == entity.StateInProgress
	})).Return(&taskUsecase.UpdateTaskOutput{Task: task}, nil)

	// Request
	reqBody := UpdateTaskRequest{
		ID:        taskID.String(),
		State:     stringPtr("IN_PROGRESS"),
		UpdatedBy: "test-user",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/Automatizacion", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response TaskResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, entity.StateInProgress.String(), response.State)
	mockUpdate.AssertExpectations(t)
}

func TestTaskHandler_Update_NotFound(t *testing.T) {
	// Setup
	mockCreate := new(MockCreateTaskUseCase)
	mockGet := new(MockGetTaskUseCase)
	mockList := new(MockListTasksUseCase)
	mockUpdate := new(MockUpdateTaskUseCase)

	handler := NewTaskHandler(mockCreate, mockGet, mockList, mockUpdate)
	router := setupTestRouter(handler)

	// Los mocks mockCreate, mockGet y mockList no se usan en este test
	_ = mockCreate
	_ = mockGet
	_ = mockList

	// Configurar mock para retornar error
	mockUpdate.On("Execute", mock.Anything, mock.Anything).Return(nil, entity.ErrTaskNotFound)

	// Request
	reqBody := UpdateTaskRequest{
		ID:        uuid.New().String(),
		State:     stringPtr("IN_PROGRESS"),
		UpdatedBy: "test-user",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/Automatizacion", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var response ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "https://api.grupoapi.com/problems/task-not-found", response.Type)
	mockUpdate.AssertExpectations(t)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
