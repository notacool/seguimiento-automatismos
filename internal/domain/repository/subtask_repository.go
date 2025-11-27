package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
)

// SubtaskRepository define el contrato para la persistencia de subtareas
type SubtaskRepository interface {
	// Create crea una nueva subtarea
	Create(ctx context.Context, taskID uuid.UUID, subtask *entity.Subtask) error

	// Update actualiza una subtarea existente
	Update(ctx context.Context, subtask *entity.Subtask) error

	// FindByID busca una subtarea por su UUID
	// Retorna entity.ErrSubtaskNotFound si no existe o está eliminada
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Subtask, error)

	// FindParentTaskID busca el UUID de la tarea padre de una subtarea
	// Retorna entity.ErrSubtaskNotFound si la subtarea no existe o está eliminada
	FindParentTaskID(ctx context.Context, subtaskID uuid.UUID) (uuid.UUID, error)

	// FindByTaskID retorna todas las subtareas de una tarea (incluyendo eliminadas si se especifica)
	FindByTaskID(ctx context.Context, taskID uuid.UUID, includeDeleted bool) ([]*entity.Subtask, error)

	// Delete marca una subtarea como eliminada (soft delete)
	Delete(ctx context.Context, id uuid.UUID, deletedBy string) error

	// DeleteByTaskID marca todas las subtareas de una tarea como eliminadas
	DeleteByTaskID(ctx context.Context, taskID uuid.UUID, deletedBy string) error
}
