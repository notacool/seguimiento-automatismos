package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	taskUsecase "github.com/grupoapi/proces-log/internal/usecase/task"
)

// CreateTaskUseCaseInterface define la interfaz para crear tareas
type CreateTaskUseCaseInterface interface {
	Execute(ctx context.Context, input taskUsecase.CreateTaskInput) (*taskUsecase.CreateTaskOutput, error)
}

// GetTaskUseCaseInterface define la interfaz para obtener tareas
type GetTaskUseCaseInterface interface {
	Execute(ctx context.Context, input taskUsecase.GetTaskInput) (*taskUsecase.GetTaskOutput, error)
}

// ListTasksUseCaseInterface define la interfaz para listar tareas
type ListTasksUseCaseInterface interface {
	Execute(ctx context.Context, input taskUsecase.ListTasksInput) (*taskUsecase.ListTasksOutput, error)
}

// UpdateTaskUseCaseInterface define la interfaz para actualizar tareas
type UpdateTaskUseCaseInterface interface {
	Execute(ctx context.Context, input taskUsecase.UpdateTaskInput) (*taskUsecase.UpdateTaskOutput, error)
}

// TaskHandler maneja las peticiones HTTP relacionadas con tareas
type TaskHandler struct {
	createUseCase CreateTaskUseCaseInterface
	getUseCase    GetTaskUseCaseInterface
	listUseCase   ListTasksUseCaseInterface
	updateUseCase UpdateTaskUseCaseInterface
}

// NewTaskHandler crea una nueva instancia de TaskHandler
func NewTaskHandler(
	createUseCase CreateTaskUseCaseInterface,
	getUseCase GetTaskUseCaseInterface,
	listUseCase ListTasksUseCaseInterface,
	updateUseCase UpdateTaskUseCaseInterface,
) *TaskHandler {
	return &TaskHandler{
		createUseCase: createUseCase,
		getUseCase:    getUseCase,
		listUseCase:   listUseCase,
		updateUseCase: updateUseCase,
	}
}

// Create maneja POST /Automatizacion
func (h *TaskHandler) Create(c *gin.Context) {
	req, ok := h.bindAndValidateCreateRequest(c)
	if !ok {
		return
	}

	initialState, ok := parseStateOrError(c, req.State)
	if !ok {
		return
	}

	input, ok := h.parseCreateInput(c, req)
	if !ok {
		return
	}

	output, err := h.createUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		MapErrorToProblemDetails(c, err)
		return
	}

	if !h.applyInitialState(c, output, initialState, req.CreatedBy) {
		return
	}

	if !h.applyInitialSubtaskStates(c, output, req.Subtasks, req.CreatedBy) {
		return
	}

	c.JSON(http.StatusCreated, ToTaskResponse(output.Task))
}

// bindAndValidateCreateRequest realiza el binding del request y valida el nombre de la tarea
func (h *TaskHandler) bindAndValidateCreateRequest(c *gin.Context) (*CreateTaskRequest, bool) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MapErrorToProblemDetails(c, entity.ErrMissingRequiredFields)
		return nil, false
	}

	if err := entity.ValidateName(req.Name); err != nil {
		MapErrorToProblemDetails(c, err)
		return nil, false
	}

	return &req, true
}

// parseCreateInput parsea el request y crea el input para el use case, validando subtareas
func (h *TaskHandler) parseCreateInput(c *gin.Context, req *CreateTaskRequest) (taskUsecase.CreateTaskInput, bool) {
	input := taskUsecase.CreateTaskInput{
		Name:         req.Name,
		CreatedBy:    req.CreatedBy,
		SubtaskNames: make([]string, 0, len(req.Subtasks)),
	}

	for _, stReq := range req.Subtasks {
		if err := entity.ValidateName(stReq.Name); err != nil {
			MapErrorToProblemDetails(c, err)
			return taskUsecase.CreateTaskInput{}, false
		}
		input.SubtaskNames = append(input.SubtaskNames, stReq.Name)
	}

	return input, true
}

// applyInitialState aplica el estado inicial de la tarea si se proporcionó uno diferente a PENDING
func (h *TaskHandler) applyInitialState(c *gin.Context, output *taskUsecase.CreateTaskOutput, initialState *entity.State, updatedBy string) bool {
	if initialState == nil || *initialState == entity.StatePending {
		return true
	}

	updateInput := taskUsecase.UpdateTaskInput{
		ID:        output.Task.ID,
		State:     initialState,
		UpdatedBy: updatedBy,
	}

	updateOutput, err := h.updateUseCase.Execute(c.Request.Context(), updateInput)
	if err != nil {
		MapErrorToProblemDetails(c, err)
		return false
	}

	output.Task = updateOutput.Task
	return true
}

// applyInitialSubtaskStates aplica los estados iniciales de las subtareas si se proporcionaron
func (h *TaskHandler) applyInitialSubtaskStates(c *gin.Context, output *taskUsecase.CreateTaskOutput, reqSubtasks []CreateSubtaskRequest, updatedBy string) bool {
	if len(reqSubtasks) == 0 || len(output.Task.Subtasks) != len(reqSubtasks) {
		return true
	}

	subtaskUpdates := make([]taskUsecase.UpdateSubtaskItemInput, 0, len(reqSubtasks))
	for i, stReq := range reqSubtasks {
		if stReq.State == nil || *stReq.State == "PENDING" {
			continue
		}

		parsedState, ok := parseStateOrError(c, stReq.State)
		if !ok {
			return false
		}

		subtaskID := output.Task.Subtasks[i].ID
		subtaskUpdates = append(subtaskUpdates, taskUsecase.UpdateSubtaskItemInput{
			ID:    &subtaskID,
			State: parsedState,
		})
	}

	if len(subtaskUpdates) == 0 {
		return true
	}

	updateInput := taskUsecase.UpdateTaskInput{
		ID:        output.Task.ID,
		UpdatedBy: updatedBy,
		Subtasks:  subtaskUpdates,
	}

	updateOutput, err := h.updateUseCase.Execute(c.Request.Context(), updateInput)
	if err != nil {
		MapErrorToProblemDetails(c, err)
		return false
	}

	output.Task = updateOutput.Task
	return true
}

// Update maneja PUT /Automatizacion
func (h *TaskHandler) Update(c *gin.Context) {
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MapErrorToProblemDetails(c, entity.ErrMissingRequiredFields)
		return
	}

	// Parsear UUID
	taskID, ok := parseUUIDOrError(c, req.ID, entity.ErrTaskNotFound)
	if !ok {
		return
	}

	// Validar nombre si se proporciona
	if req.Name != nil {
		if err := entity.ValidateName(*req.Name); err != nil {
			MapErrorToProblemDetails(c, err)
			return
		}
	}

	// Parsear estado si se proporciona
	state, ok := parseStateOrError(c, req.State)
	if !ok {
		return
	}

	// Convertir subtareas del request
	subtaskInputs := make([]taskUsecase.UpdateSubtaskItemInput, 0, len(req.Subtasks))
		for _, stReq := range req.Subtasks {
		var stID *uuid.UUID
		if stReq.ID != nil {
			parsedID, ok := parseUUIDOrError(c, *stReq.ID, entity.ErrSubtaskNotFound)
			if !ok {
				return
			}
			stID = &parsedID
		}

		// Validar que tenga ID o Name
		if stID == nil && stReq.Name == nil {
			MapErrorToProblemDetails(c, entity.ErrMissingRequiredFields)
			return
		}

		// Validar nombre si se proporciona
		if stReq.Name != nil {
			if err := entity.ValidateName(*stReq.Name); err != nil {
				MapErrorToProblemDetails(c, err)
				return
			}
		}

		stState, ok := parseStateOrError(c, stReq.State)
		if !ok {
			return
		}

		subtaskInputs = append(subtaskInputs, taskUsecase.UpdateSubtaskItemInput{
			ID:    stID,
			Name:  stReq.Name,
			State: stState,
		})
	}

	// Crear input para el use case
	input := taskUsecase.UpdateTaskInput{
		ID:        taskID,
		Name:      req.Name,
		State:     state,
		UpdatedBy: req.UpdatedBy,
		Subtasks:  subtaskInputs,
	}

	// Ejecutar use case
	output, err := h.updateUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		MapErrorToProblemDetails(c, err)
		return
	}

	c.JSON(http.StatusOK, ToTaskResponse(output.Task))
}

// Get maneja GET /Automatizacion/{uuid}
func (h *TaskHandler) Get(c *gin.Context) {
	uuidStr := c.Param("uuid")
	taskID, ok := parseUUIDOrError(c, uuidStr, entity.ErrTaskNotFound)
	if !ok {
		return
	}

	input := taskUsecase.GetTaskInput{
		ID: taskID,
	}

	output, err := h.getUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		MapErrorToProblemDetails(c, err)
		return
	}

	c.JSON(http.StatusOK, ToTaskResponse(output.Task))
}

// List maneja GET /AutomatizacionListado
func (h *TaskHandler) List(c *gin.Context) {
	// Parsear query parameters
	var state *entity.State
	if stateStr := c.Query("state"); stateStr != "" {
		statePtr := &stateStr
		parsedState, ok := parseStateOrError(c, statePtr)
		if !ok {
			return
		}
		state = parsedState
	}

	var name *string
	if nameStr := c.Query("name"); nameStr != "" {
		name = &nameStr
	}

	// Parsear paginación
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage < 1 {
			MapErrorToProblemDetails(c, entity.ErrMissingRequiredFields)
			return
		}
		page = parsedPage
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			MapErrorToProblemDetails(c, entity.ErrMissingRequiredFields)
			return
		}
		limit = parsedLimit
	}

	input := taskUsecase.ListTasksInput{
		State:          state,
		NameContains:   name,
		Page:           page,
		Limit:          limit,
		IncludeDeleted: false,
	}

	output, err := h.listUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		MapErrorToProblemDetails(c, err)
		return
	}

	// Convertir a response
	tasks := make([]TaskResponse, 0, len(output.Tasks))
	for _, task := range output.Tasks {
		tasks = append(tasks, ToTaskResponse(task))
	}

	response := TaskListResponse{
		Tasks: tasks,
		Pagination: PaginationResponse{
			Page:       output.Page,
			Limit:      output.Limit,
			Total:      output.Total,
			TotalPages: output.TotalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}
