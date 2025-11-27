package subtask

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/grupoapi/proces-log/internal/domain/repository"
)

// DeleteSubtaskInput representa los datos de entrada para eliminar una subtarea
type DeleteSubtaskInput struct {
	ID        uuid.UUID
	DeletedBy string
}

// DeleteSubtaskOutput representa el resultado de eliminar una subtarea
type DeleteSubtaskOutput struct {
	Success bool
}

// DeleteSubtaskUseCase maneja la eliminación (soft delete) de subtareas individuales
type DeleteSubtaskUseCase struct {
	subtaskRepo repository.SubtaskRepository
}

// NewDeleteSubtaskUseCase crea una nueva instancia del caso de uso
func NewDeleteSubtaskUseCase(subtaskRepo repository.SubtaskRepository) *DeleteSubtaskUseCase {
	return &DeleteSubtaskUseCase{
		subtaskRepo: subtaskRepo,
	}
}

// Execute ejecuta el caso de uso de eliminación de subtarea
func (uc *DeleteSubtaskUseCase) Execute(ctx context.Context, input DeleteSubtaskInput) (*DeleteSubtaskOutput, error) {
	// Validar input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Eliminar subtarea (soft delete)
	if err := uc.subtaskRepo.Delete(ctx, input.ID, input.DeletedBy); err != nil {
		return nil, fmt.Errorf("failed to delete subtask: %w", err)
	}

	return &DeleteSubtaskOutput{Success: true}, nil
}

// validateInput valida los datos de entrada
func (uc *DeleteSubtaskUseCase) validateInput(input DeleteSubtaskInput) error {
	if input.ID == uuid.Nil {
		return fmt.Errorf("%w: id is required", entity.ErrMissingRequiredFields)
	}
	if input.DeletedBy == "" {
		return fmt.Errorf("%w: deleted_by is required", entity.ErrMissingRequiredFields)
	}
	return nil
}
