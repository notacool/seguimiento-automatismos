package task

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/grupoapi/proces-log/internal/domain/repository"
	"github.com/grupoapi/proces-log/internal/domain/service"
)

// UpdateTaskInput representa los datos de entrada para actualizar una tarea
type UpdateTaskInput struct {
	ID        uuid.UUID
	Name      *string       // Opcional: nuevo nombre
	State     *entity.State // Opcional: nuevo estado
	UpdatedBy string
}

// UpdateTaskOutput representa el resultado de actualizar una tarea
type UpdateTaskOutput struct {
	Task *entity.Task
}

// UpdateTaskUseCase maneja la actualización de tareas existentes
type UpdateTaskUseCase struct {
	taskRepo     repository.TaskRepository
	stateMachine *service.StateMachine
}

// NewUpdateTaskUseCase crea una nueva instancia del caso de uso
func NewUpdateTaskUseCase(taskRepo repository.TaskRepository, stateMachine *service.StateMachine) *UpdateTaskUseCase {
	return &UpdateTaskUseCase{
		taskRepo:     taskRepo,
		stateMachine: stateMachine,
	}
}

// Execute ejecuta el caso de uso de actualización de tarea
func (uc *UpdateTaskUseCase) Execute(ctx context.Context, input UpdateTaskInput) (*UpdateTaskOutput, error) {
	// Validar input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Buscar tarea existente
	task, err := uc.taskRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	// Actualizar nombre si se proporciona
	if input.Name != nil {
		if err := entity.ValidateName(*input.Name); err != nil {
			return nil, err
		}
		task.Name = *input.Name
		task.UpdatedBy = input.UpdatedBy
	}

	// Actualizar estado si se proporciona
	if input.State != nil {
		// Validar transición de estado
		if err := uc.stateMachine.ValidateTaskStateTransition(task, *input.State); err != nil {
			return nil, err
		}

		// Actualizar estado usando el método del dominio que gestiona fechas
		if err := task.UpdateState(*input.State, input.UpdatedBy); err != nil {
			return nil, fmt.Errorf("failed to update task state: %w", err)
		}
	}

	// Persistir cambios
	if err := uc.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to persist task updates: %w", err)
	}

	return &UpdateTaskOutput{Task: task}, nil
}

// validateInput valida los datos de entrada
func (uc *UpdateTaskUseCase) validateInput(input UpdateTaskInput) error {
	if input.ID == uuid.Nil {
		return fmt.Errorf("%w: id is required", entity.ErrMissingRequiredFields)
	}
	if input.UpdatedBy == "" {
		return fmt.Errorf("%w: updated_by is required", entity.ErrMissingRequiredFields)
	}
	// Al menos uno de los campos debe estar presente
	if input.Name == nil && input.State == nil {
		return fmt.Errorf("%w: at least one field (name or state) must be provided", entity.ErrMissingRequiredFields)
	}
	return nil
}
