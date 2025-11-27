package task

import (
	"context"
	"fmt"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/grupoapi/proces-log/internal/domain/repository"
)

// ListTasksInput representa los datos de entrada para listar tareas
type ListTasksInput struct {
	State          *entity.State // Filtro opcional por estado
	NameContains   *string       // Filtro opcional por nombre (búsqueda parcial)
	Page           int           // Número de página (1-indexed)
	Limit          int           // Cantidad de resultados por página
	IncludeDeleted bool          // Incluir tareas eliminadas
}

// ListTasksOutput representa el resultado de listar tareas
type ListTasksOutput struct {
	Tasks      []*entity.Task
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

// ListTasksUseCase maneja el listado paginado de tareas con filtros
type ListTasksUseCase struct {
	taskRepo repository.TaskRepository
}

// NewListTasksUseCase crea una nueva instancia del caso de uso
func NewListTasksUseCase(taskRepo repository.TaskRepository) *ListTasksUseCase {
	return &ListTasksUseCase{
		taskRepo: taskRepo,
	}
}

// Execute ejecuta el caso de uso de listado de tareas
func (uc *ListTasksUseCase) Execute(ctx context.Context, input ListTasksInput) (*ListTasksOutput, error) {
	// Validar y normalizar input
	if err := uc.validateInput(&input); err != nil {
		return nil, err
	}

	// Construir filtros para el repositorio
	filters := repository.TaskFilters{
		State:          input.State,
		Name:           input.NameContains,
		Page:           input.Page,
		Limit:          input.Limit,
		Offset:         (input.Page - 1) * input.Limit,
		IncludeDeleted: input.IncludeDeleted,
	}

	// Obtener tareas del repositorio
	result, err := uc.taskRepo.FindAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return &ListTasksOutput{
		Tasks:      result.Tasks,
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
	}, nil
}

// validateInput valida y normaliza los datos de entrada
func (uc *ListTasksUseCase) validateInput(input *ListTasksInput) error {
	// Validar y normalizar paginación
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 {
		input.Limit = 20 // Default
	}
	if input.Limit > 100 {
		input.Limit = 100 // Máximo permitido
	}

	// Validar estado si se proporciona
	if input.State != nil && !input.State.IsValid() {
		return fmt.Errorf("%w: invalid state filter", entity.ErrInvalidStateTransition)
	}

	return nil
}
