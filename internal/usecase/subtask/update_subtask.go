package subtask

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/grupoapi/proces-log/internal/domain/repository"
	"github.com/grupoapi/proces-log/internal/domain/service"
)

// UpdateSubtaskInput representa los datos de entrada para actualizar una subtarea
type UpdateSubtaskInput struct {
	ID        uuid.UUID
	Name      *string       // Opcional: nuevo nombre
	State     *entity.State // Opcional: nuevo estado
	UpdatedBy string
}

// UpdateSubtaskOutput representa el resultado de actualizar una subtarea
type UpdateSubtaskOutput struct {
	Subtask *entity.Subtask
}

// UpdateSubtaskUseCase maneja la actualización de subtareas individuales
type UpdateSubtaskUseCase struct {
	subtaskRepo  repository.SubtaskRepository
	taskRepo     repository.TaskRepository
	stateMachine *service.StateMachine
}

// NewUpdateSubtaskUseCase crea una nueva instancia del caso de uso
func NewUpdateSubtaskUseCase(
	subtaskRepo repository.SubtaskRepository,
	taskRepo repository.TaskRepository,
	stateMachine *service.StateMachine,
) *UpdateSubtaskUseCase {
	return &UpdateSubtaskUseCase{
		subtaskRepo:  subtaskRepo,
		taskRepo:     taskRepo,
		stateMachine: stateMachine,
	}
}

// Execute ejecuta el caso de uso de actualización de subtarea
func (uc *UpdateSubtaskUseCase) Execute(ctx context.Context, input UpdateSubtaskInput) (*UpdateSubtaskOutput, error) {
	// Validar input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Buscar subtarea existente
	subtask, err := uc.subtaskRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find subtask: %w", err)
	}

	// Actualizar nombre si se proporciona
	if input.Name != nil {
		if err := entity.ValidateName(*input.Name); err != nil {
			return nil, err
		}
		subtask.Name = *input.Name
		subtask.UpdatedAt = time.Now()
	}

	// Actualizar estado si se proporciona
	if input.State != nil {
		// Necesitamos la tarea padre para validar la transición
		// Primero encontramos qué tarea contiene esta subtarea
		task, err := uc.findParentTask(ctx, subtask.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to find parent task: %w", err)
		}

		// Validar transición de estado considerando la tarea padre
		if err := uc.stateMachine.ValidateSubtaskStateTransition(task, subtask, *input.State); err != nil {
			return nil, err
		}

		// Actualizar estado y fechas según el nuevo estado
		if *input.State == entity.StateInProgress {
			subtask.SetStartDate()
		}
		if input.State.IsFinal() {
			subtask.SetEndDate()
		}

		subtask.State = *input.State
		subtask.UpdatedAt = time.Now()
	}

	// Persistir cambios
	if err := uc.subtaskRepo.Update(ctx, subtask); err != nil {
		return nil, fmt.Errorf("failed to persist subtask updates: %w", err)
	}

	return &UpdateSubtaskOutput{Subtask: subtask}, nil
}

// validateInput valida los datos de entrada
func (uc *UpdateSubtaskUseCase) validateInput(input UpdateSubtaskInput) error {
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

// findParentTask encuentra la tarea padre de una subtarea
// Utiliza la foreign key task_id en la tabla subtasks para una búsqueda O(1)
func (uc *UpdateSubtaskUseCase) findParentTask(ctx context.Context, subtaskID uuid.UUID) (*entity.Task, error) {
	// Obtener el task_id directamente de la subtarea (O(1))
	taskID, err := uc.subtaskRepo.FindParentTaskID(ctx, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find parent task ID: %w", err)
	}

	// Cargar la tarea completa
	task, err := uc.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to load parent task: %w", err)
	}

	return task, nil
}
