package http

import (
	"time"

	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
)

// CreateTaskRequest representa el request para crear una tarea
type CreateTaskRequest struct {
	Name      string                 `json:"name" binding:"required"`
	State     *string                `json:"state,omitempty"`
	CreatedBy string                 `json:"created_by" binding:"required"`
	Subtasks  []CreateSubtaskRequest `json:"subtasks,omitempty"`
}

// CreateSubtaskRequest representa una subtarea en el request de creación
type CreateSubtaskRequest struct {
	Name  string  `json:"name" binding:"required"`
	State *string `json:"state,omitempty"`
}

// UpdateTaskRequest representa el request para actualizar una tarea
type UpdateTaskRequest struct {
	ID        string                     `json:"id" binding:"required"`
	Name      *string                    `json:"name,omitempty"`
	State     *string                    `json:"state,omitempty"`
	UpdatedBy string                     `json:"updated_by" binding:"required"`
	Subtasks  []UpdateSubtaskItemRequest `json:"subtasks,omitempty"`
}

// UpdateSubtaskItemRequest representa una subtarea en el request de actualización
// Si tiene ID, es una subtarea existente a actualizar
// Si no tiene ID pero tiene Name, es una nueva subtarea a crear
type UpdateSubtaskItemRequest struct {
	ID    *string `json:"id,omitempty"`
	Name  *string `json:"name,omitempty"`
	State *string `json:"state,omitempty"`
}

// TaskResponse representa la respuesta de una tarea
type TaskResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	State     string            `json:"state"`
	Subtasks  []SubtaskResponse `json:"subtasks"`
	CreatedBy string            `json:"created_by"`
	UpdatedBy *string           `json:"updated_by,omitempty"`
	StartDate *time.Time        `json:"start_date,omitempty"`
	EndDate   *time.Time        `json:"end_date,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty"`
}

// SubtaskResponse representa la respuesta de una subtarea
type SubtaskResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	State     string     `json:"state"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// TaskListResponse representa la respuesta del listado de tareas
type TaskListResponse struct {
	Tasks      []TaskResponse     `json:"tasks"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginationResponse representa la información de paginación
type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ToTaskResponse convierte una entidad Task a TaskResponse
func ToTaskResponse(task *entity.Task) TaskResponse {
	subtasks := make([]SubtaskResponse, 0, len(task.Subtasks))
	for _, st := range task.Subtasks {
		if !st.IsDeleted() {
			subtasks = append(subtasks, ToSubtaskResponse(st))
		}
	}

	var updatedBy *string
	if task.UpdatedBy != "" {
		updatedBy = &task.UpdatedBy
	}

	return TaskResponse{
		ID:        task.ID.String(),
		Name:      task.Name,
		State:     task.State.String(),
		Subtasks:  subtasks,
		CreatedBy: task.CreatedBy,
		UpdatedBy: updatedBy,
		StartDate: task.StartDate,
		EndDate:   task.EndDate,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
		DeletedAt: task.DeletedAt,
	}
}

// ToSubtaskResponse convierte una entidad Subtask a SubtaskResponse
func ToSubtaskResponse(subtask *entity.Subtask) SubtaskResponse {
	return SubtaskResponse{
		ID:        subtask.ID.String(),
		Name:      subtask.Name,
		State:     subtask.State.String(),
		StartDate: subtask.StartDate,
		EndDate:   subtask.EndDate,
		CreatedAt: subtask.CreatedAt,
		UpdatedAt: subtask.UpdatedAt,
		DeletedAt: subtask.DeletedAt,
	}
}

// ParseState convierte un string a entity.State
func ParseState(stateStr string) (entity.State, error) {
	state := entity.State(stateStr)
	if !state.IsValid() {
		return entity.StatePending, entity.ErrInvalidStateTransition
	}
	return state, nil
}

// ParseUUID convierte un string a uuid.UUID
func ParseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}
