package task

import (
	"context"
	"fmt"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/grupoapi/proces-log/internal/domain/repository"
)

// CreateTaskInput representa los datos de entrada para crear una tarea
type CreateTaskInput struct {
	Name         string
	CreatedBy    string
	SubtaskNames []string // Nombres de las subtareas (opcional)
}

// CreateTaskOutput representa el resultado de crear una tarea
type CreateTaskOutput struct {
	Task *entity.Task
}

// CreateTaskUseCase maneja la creación de nuevas tareas
type CreateTaskUseCase struct {
	taskRepo repository.TaskRepository
}

// NewCreateTaskUseCase crea una nueva instancia del caso de uso
func NewCreateTaskUseCase(taskRepo repository.TaskRepository) *CreateTaskUseCase {
	return &CreateTaskUseCase{
		taskRepo: taskRepo,
	}
}

// Execute ejecuta el caso de uso de creación de tarea
func (uc *CreateTaskUseCase) Execute(ctx context.Context, input CreateTaskInput) (*CreateTaskOutput, error) {
	// Validar input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Crear tarea usando constructor del dominio
	task, err := entity.NewTask(input.Name, input.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create task entity: %w", err)
	}

	// Crear subtareas si se proporcionaron nombres
	if len(input.SubtaskNames) > 0 {
		for _, subtaskName := range input.SubtaskNames {
			subtask, err := entity.NewSubtask(subtaskName)
			if err != nil {
				return nil, fmt.Errorf("failed to create subtask entity: %w", err)
			}
			task.AddSubtask(subtask)
		}
	}

	// Persistir en repositorio
	if err := uc.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to persist task: %w", err)
	}

	return &CreateTaskOutput{Task: task}, nil
}

// validateInput valida los datos de entrada
func (uc *CreateTaskUseCase) validateInput(input CreateTaskInput) error {
	if input.Name == "" {
		return fmt.Errorf("%w: name is required", entity.ErrMissingRequiredFields)
	}
	if input.CreatedBy == "" {
		return fmt.Errorf("%w: created_by is required", entity.ErrMissingRequiredFields)
	}
	return nil
}
