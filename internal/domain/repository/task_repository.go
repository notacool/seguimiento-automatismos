package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
)

// TaskFilters representa los filtros para listar tareas
type TaskFilters struct {
	State          *entity.State // Filtrar por estado (opcional)
	NameContains   string        // Búsqueda parcial en nombre (case-insensitive)
	Page           int           // Número de página (1-indexed)
	Limit          int           // Cantidad de resultados por página
	IncludeDeleted bool          // Incluir tareas eliminadas (soft-deleted)
}

// TaskListResult representa el resultado paginado de tareas
type TaskListResult struct {
	Tasks      []*entity.Task
	Total      int // Total de resultados (sin paginación)
	Page       int
	Limit      int
	TotalPages int
}

// TaskRepository define el contrato para la persistencia de tareas
type TaskRepository interface {
	// Create crea una nueva tarea con sus subtareas en una transacción
	Create(ctx context.Context, task *entity.Task) error

	// Update actualiza una tarea existente y sus subtareas en una transacción
	// Puede añadir, modificar o eliminar subtareas
	Update(ctx context.Context, task *entity.Task) error

	// FindByID busca una tarea por su UUID
	// Retorna entity.ErrTaskNotFound si no existe o está eliminada
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Task, error)

	// FindAll retorna una lista paginada de tareas según los filtros
	// Ordena siempre por created_at DESC
	FindAll(ctx context.Context, filters TaskFilters) (*TaskListResult, error)

	// Delete marca una tarea como eliminada (soft delete)
	Delete(ctx context.Context, id uuid.UUID, deletedBy string) error

	// HardDelete elimina permanentemente tareas soft-deleted hace más de 30 días
	// Usado por el job de limpieza automática
	HardDelete(ctx context.Context) (int, error)
}
