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
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MapErrorToProblemDetails(c, entity.ErrMissingRequiredFields)
		return
	}

	// Validar nombre
	if err := entity.ValidateName(req.Name); err != nil {
		MapErrorToProblemDetails(c, err)
		return
	}

	// Parsear estado si se proporciona
	var initialState *entity.State
	if req.State != nil {
		state, err := ParseState(*req.State)
		if err != nil {
			MapErrorToProblemDetails(c, err)
			return
		}
		initialState = &state
	}

	// Crear input para el use case
	input := taskUsecase.CreateTaskInput{
		Name:         req.Name,
		CreatedBy:    req.CreatedBy,
		SubtaskNames: make([]string, 0, len(req.Subtasks)),
	}

	// Validar y procesar subtareas
	for _, stReq := range req.Subtasks {
		if err := entity.ValidateName(stReq.Name); err != nil {
			MapErrorToProblemDetails(c, err)
			return
		}
		input.SubtaskNames = append(input.SubtaskNames, stReq.Name)
	}

	// Ejecutar use case
	output, err := h.createUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate) // Log error for debugging
		MapErrorToProblemDetails(c, err)
		return
	}

	// Actualizar estado de la tarea si se proporcionó uno diferente a PENDING
	if initialState != nil && *initialState != entity.StatePending {
		updateInput := taskUsecase.UpdateTaskInput{
			ID:        output.Task.ID,
			State:     initialState,
			UpdatedBy: req.CreatedBy,
		}
		updateOutput, err := h.updateUseCase.Execute(c.Request.Context(), updateInput)
		if err != nil {
			MapErrorToProblemDetails(c, err)
			return
		}
		output.Task = updateOutput.Task
	}

	// Actualizar estados de subtareas si se proporcionaron
	if len(req.Subtasks) > 0 && len(output.Task.Subtasks) == len(req.Subtasks) {
		subtaskUpdates := make([]taskUsecase.UpdateSubtaskItemInput, 0, len(req.Subtasks))
		for i, stReq := range req.Subtasks {
			if stReq.State != nil && *stReq.State != "PENDING" {
				parsedState, err := ParseState(*stReq.State)
				if err != nil {
					MapErrorToProblemDetails(c, err)
					return
				}
				subtaskID := output.Task.Subtasks[i].ID
				subtaskUpdates = append(subtaskUpdates, taskUsecase.UpdateSubtaskItemInput{
					ID:    &subtaskID,
					State: &parsedState,
				})
			}
		}

		if len(subtaskUpdates) > 0 {
			updateInput := taskUsecase.UpdateTaskInput{
				ID:        output.Task.ID,
				UpdatedBy: req.CreatedBy,
				Subtasks:  subtaskUpdates,
			}
			updateOutput, err := h.updateUseCase.Execute(c.Request.Context(), updateInput)
			if err != nil {
				MapErrorToProblemDetails(c, err)
				return
			}
			output.Task = updateOutput.Task
		}
	}

	c.JSON(http.StatusCreated, ToTaskResponse(output.Task))
}

// Update maneja PUT /Automatizacion
func (h *TaskHandler) Update(c *gin.Context) {
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MapErrorToProblemDetails(c, entity.ErrMissingRequiredFields)
		return
	}

	// Parsear UUID
	taskID, err := ParseUUID(req.ID)
	if err != nil {
		MapErrorToProblemDetails(c, entity.ErrTaskNotFound)
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
	var state *entity.State
	if req.State != nil {
		parsedState, err := ParseState(*req.State)
		if err != nil {
			MapErrorToProblemDetails(c, err)
			return
		}
		state = &parsedState
	}

	// Convertir subtareas del request
	subtaskInputs := make([]taskUsecase.UpdateSubtaskItemInput, 0, len(req.Subtasks))
	for _, stReq := range req.Subtasks {
		var stID *uuid.UUID
		if stReq.ID != nil {
			parsedID, err := ParseUUID(*stReq.ID)
			if err != nil {
				MapErrorToProblemDetails(c, entity.ErrSubtaskNotFound)
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

		var stState *entity.State
		if stReq.State != nil {
			parsedState, err := ParseState(*stReq.State)
			if err != nil {
				MapErrorToProblemDetails(c, err)
				return
			}
			stState = &parsedState
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
	taskID, err := ParseUUID(uuidStr)
	if err != nil {
		MapErrorToProblemDetails(c, entity.ErrTaskNotFound)
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
		parsedState, err := ParseState(stateStr)
		if err != nil {
			MapErrorToProblemDetails(c, err)
			return
		}
		state = &parsedState
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
