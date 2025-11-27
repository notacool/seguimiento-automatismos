package task

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/grupoapi/proces-log/internal/domain/repository"
)

// DeleteTaskInput representa los datos de entrada para eliminar una tarea
type DeleteTaskInput struct {
	ID        uuid.UUID
	DeletedBy string
}

// DeleteTaskOutput representa el resultado de eliminar una tarea
type DeleteTaskOutput struct {
	Success bool
}

// DeleteTaskUseCase maneja la eliminación (soft delete) de tareas
type DeleteTaskUseCase struct {
	taskRepo repository.TaskRepository
}

// NewDeleteTaskUseCase crea una nueva instancia del caso de uso
func NewDeleteTaskUseCase(taskRepo repository.TaskRepository) *DeleteTaskUseCase {
	return &DeleteTaskUseCase{
		taskRepo: taskRepo,
	}
}

// Execute ejecuta el caso de uso de eliminación de tarea
func (uc *DeleteTaskUseCase) Execute(ctx context.Context, input DeleteTaskInput) (*DeleteTaskOutput, error) {
	// Validar input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Eliminar tarea (soft delete)
	if err := uc.taskRepo.Delete(ctx, input.ID, input.DeletedBy); err != nil {
		return nil, fmt.Errorf("failed to delete task: %w", err)
	}

	return &DeleteTaskOutput{Success: true}, nil
}

// validateInput valida los datos de entrada
func (uc *DeleteTaskUseCase) validateInput(input DeleteTaskInput) error {
	if input.ID == uuid.Nil {
		return fmt.Errorf("%w: id is required", entity.ErrMissingRequiredFields)
	}
	if input.DeletedBy == "" {
		return fmt.Errorf("%w: deleted_by is required", entity.ErrMissingRequiredFields)
	}
	return nil
}
