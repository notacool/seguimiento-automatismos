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

	// Si la subtarea pasó a estado final, verificar si todas las subtareas están completas
	// para completar automáticamente la tarea padre
	if input.State != nil && input.State.IsFinal() {
		// Obtener el ID de la tarea padre
		parentTaskID, err := uc.subtaskRepo.FindParentTaskID(ctx, subtask.ID)
		if err == nil {
			// Recargar la tarea padre para tener subtareas actualizadas
			parentTask, err := uc.taskRepo.FindByID(ctx, parentTaskID)
			if err == nil {
				if err := uc.checkAndCompleteParentTask(ctx, parentTask, input.UpdatedBy); err != nil {
					// Log el error pero no fallar la operación principal
					// La subtarea ya fue actualizada exitosamente
					fmt.Printf("Warning: failed to auto-complete parent task: %v\n", err)
				}
			}
		}
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

// checkAndCompleteParentTask verifica si todas las subtareas están completas
// y completa automáticamente la tarea padre si es necesario
func (uc *UpdateSubtaskUseCase) checkAndCompleteParentTask(ctx context.Context, task *entity.Task, updatedBy string) error {
	// Solo proceder si la tarea está en IN_PROGRESS
	if task.State != entity.StateInProgress {
		return nil // No hay nada que hacer
	}

	// Obtener todas las subtareas de la tarea
	subtasks, err := uc.subtaskRepo.FindByTaskID(ctx, task.ID, false) // false = no incluir eliminadas
	if err != nil {
		return fmt.Errorf("failed to find subtasks: %w", err)
	}

	// Si no hay subtareas, no hay nada que hacer
	if len(subtasks) == 0 {
		return nil
	}

	// Verificar si todas las subtareas están en estado final exitoso (COMPLETED)
	allCompleted := true
	for _, subtask := range subtasks {
		if subtask.State != entity.StateCompleted {
			allCompleted = false
			break
		}
	}

	// Si todas las subtareas están completas, completar la tarea padre
	if allCompleted {
		// Validar que la transición sea válida
		if err := uc.stateMachine.ValidateTaskStateTransition(task, entity.StateCompleted); err != nil {
			return fmt.Errorf("invalid transition to complete parent task: %w", err)
		}

		// Actualizar estado y fechas
		task.State = entity.StateCompleted
		task.SetEndDate()
		task.UpdatedAt = time.Now()

		// Persistir los cambios
		if err := uc.taskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("failed to update parent task: %w", err)
		}
	}

	return nil
}
