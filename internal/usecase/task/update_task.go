package task

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/grupoapi/proces-log/internal/domain/repository"
	"github.com/grupoapi/proces-log/internal/domain/service"
)

// UpdateSubtaskItemInput representa una subtarea en el request de actualización
type UpdateSubtaskItemInput struct {
	ID    *uuid.UUID    // Si tiene ID, es una subtarea existente a actualizar
	Name  *string       // Si no tiene ID pero tiene Name, es una nueva subtarea a crear
	State *entity.State // Nuevo estado (opcional)
}

// UpdateTaskInput representa los datos de entrada para actualizar una tarea
type UpdateTaskInput struct {
	ID        uuid.UUID
	Name      *string       // Opcional: nuevo nombre
	State     *entity.State // Opcional: nuevo estado
	UpdatedBy string
	Subtasks  []UpdateSubtaskItemInput // Opcional: lista de subtareas a actualizar/añadir/eliminar
}

// UpdateTaskOutput representa el resultado de actualizar una tarea
type UpdateTaskOutput struct {
	Task *entity.Task
}

// UpdateTaskUseCase maneja la actualización de tareas existentes
type UpdateTaskUseCase struct {
	taskRepo     repository.TaskRepository
	subtaskRepo  repository.SubtaskRepository
	stateMachine *service.StateMachine
}

// NewUpdateTaskUseCase crea una nueva instancia del caso de uso
func NewUpdateTaskUseCase(
	taskRepo repository.TaskRepository,
	subtaskRepo repository.SubtaskRepository,
	stateMachine *service.StateMachine,
) *UpdateTaskUseCase {
	return &UpdateTaskUseCase{
		taskRepo:     taskRepo,
		subtaskRepo:  subtaskRepo,
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

	// Manejar subtareas si se proporcionan
	if len(input.Subtasks) > 0 {
		if err := uc.handleSubtasks(ctx, task, input.Subtasks); err != nil {
			return nil, fmt.Errorf("failed to handle subtasks: %w", err)
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
	if input.Name == nil && input.State == nil && len(input.Subtasks) == 0 {
		return fmt.Errorf("%w: at least one field (name, state, or subtasks) must be provided", entity.ErrMissingRequiredFields)
	}
	return nil
}

// handleSubtasks procesa las subtareas del request de actualización
func (uc *UpdateTaskUseCase) handleSubtasks(
	ctx context.Context,
	task *entity.Task,
	subtaskInputs []UpdateSubtaskItemInput,
) error {
	// Crear un mapa de subtareas existentes por ID para acceso rápido
	existingSubtasksMap := make(map[uuid.UUID]*entity.Subtask)
	for _, st := range task.Subtasks {
		if !st.IsDeleted() {
			existingSubtasksMap[st.ID] = st
		}
	}

	// Procesar cada subtarea del input
	processedIDs := make(map[uuid.UUID]bool)
	for _, stInput := range subtaskInputs {
		if stInput.ID != nil {
			// Actualizar subtarea existente
			subtask, exists := existingSubtasksMap[*stInput.ID]
			if !exists {
				// Intentar cargar desde el repositorio
				loadedSubtask, err := uc.subtaskRepo.FindByID(ctx, *stInput.ID)
				if err != nil {
					return fmt.Errorf("subtask with ID %s not found: %w", *stInput.ID, err)
				}
				subtask = loadedSubtask
			}

			// Actualizar nombre si se proporciona
			if stInput.Name != nil {
				if err := entity.ValidateName(*stInput.Name); err != nil {
					return err
				}
				subtask.Name = *stInput.Name
			}

			// Actualizar estado si se proporciona
			if stInput.State != nil {
				// Validar transición de estado
				if err := uc.stateMachine.ValidateSubtaskStateTransition(task, subtask, *stInput.State); err != nil {
					return err
				}

				// Actualizar estado y fechas
				if *stInput.State == entity.StateInProgress {
					subtask.SetStartDate()
				}
				if stInput.State.IsFinal() {
					subtask.SetEndDate()
				}
				subtask.State = *stInput.State
			}

			subtask.UpdatedAt = task.UpdatedAt
			processedIDs[subtask.ID] = true

			// Actualizar en la lista de subtareas de la tarea
			found := false
			for i, st := range task.Subtasks {
				if st.ID == subtask.ID {
					task.Subtasks[i] = subtask
					found = true
					break
				}
			}
			if !found {
				task.Subtasks = append(task.Subtasks, subtask)
			}
		} else if stInput.Name != nil {
			// Crear nueva subtarea
			newSubtask, err := entity.NewSubtask(*stInput.Name)
			if err != nil {
				return fmt.Errorf("failed to create subtask: %w", err)
			}

			// Establecer estado si se proporciona
			if stInput.State != nil {
				// Validar transición de estado (desde PENDING)
				if err := uc.stateMachine.ValidateSubtaskStateTransition(task, newSubtask, *stInput.State); err != nil {
					return err
				}
				newSubtask.State = *stInput.State
				if *stInput.State == entity.StateInProgress {
					newSubtask.SetStartDate()
				}
				if stInput.State.IsFinal() {
					newSubtask.SetEndDate()
				}
			}

			task.AddSubtask(newSubtask)
		}
	}

	// Eliminar subtareas que no están en la lista (soft delete)
	for _, existingSubtask := range task.Subtasks {
		if !existingSubtask.IsDeleted() && !processedIDs[existingSubtask.ID] {
			// Marcar como eliminada
			existingSubtask.Delete()
		}
	}

	return nil
}
