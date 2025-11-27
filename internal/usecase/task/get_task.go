package task

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/grupoapi/proces-log/internal/domain/repository"
)

// GetTaskInput representa los datos de entrada para obtener una tarea
type GetTaskInput struct {
	ID uuid.UUID
}

// GetTaskOutput representa el resultado de obtener una tarea
type GetTaskOutput struct {
	Task *entity.Task
}

// GetTaskUseCase maneja la obtención de una tarea por ID
type GetTaskUseCase struct {
	taskRepo repository.TaskRepository
}

// NewGetTaskUseCase crea una nueva instancia del caso de uso
func NewGetTaskUseCase(taskRepo repository.TaskRepository) *GetTaskUseCase {
	return &GetTaskUseCase{
		taskRepo: taskRepo,
	}
}

// Execute ejecuta el caso de uso de obtención de tarea
func (uc *GetTaskUseCase) Execute(ctx context.Context, input GetTaskInput) (*GetTaskOutput, error) {
	// Validar input
	if input.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: id is required", entity.ErrMissingRequiredFields)
	}

	// Buscar tarea
	task, err := uc.taskRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	return &GetTaskOutput{Task: task}, nil
}
