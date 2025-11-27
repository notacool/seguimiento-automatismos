package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/grupoapi/proces-log/internal/domain/repository"
)

// SubtaskRepository implementa el repositorio de subtareas usando PostgreSQL
type SubtaskRepository struct {
	pool *pgxpool.Pool
}

// NewSubtaskRepository crea una nueva instancia del repositorio de subtareas
func NewSubtaskRepository(pool *pgxpool.Pool) repository.SubtaskRepository {
	return &SubtaskRepository{pool: pool}
}

// Create crea una nueva subtarea
func (r *SubtaskRepository) Create(ctx context.Context, taskID uuid.UUID, subtask *entity.Subtask) error {
	query := `
		INSERT INTO subtasks (id, task_id, name, state, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		subtask.ID,
		taskID,
		subtask.Name,
		subtask.State.String(),
		subtask.StartDate,
		subtask.EndDate,
		subtask.CreatedAt,
		subtask.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create subtask: %w", err)
	}

	return nil
}

// FindByID busca una subtarea por su ID
func (r *SubtaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Subtask, error) {
	query := `
		SELECT id, name, state, start_date, end_date, created_at, updated_at, deleted_at
		FROM subtasks
		WHERE id = $1 AND deleted_at IS NULL
	`

	var subtask entity.Subtask
	var state string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&subtask.ID,
		&subtask.Name,
		&state,
		&subtask.StartDate,
		&subtask.EndDate,
		&subtask.CreatedAt,
		&subtask.UpdatedAt,
		&subtask.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrSubtaskNotFound
		}
		return nil, fmt.Errorf("failed to find subtask by ID: %w", err)
	}

	subtask.State = entity.State(state)

	return &subtask, nil
}

// FindParentTaskID busca el UUID de la tarea padre de una subtarea
func (r *SubtaskRepository) FindParentTaskID(ctx context.Context, subtaskID uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT task_id
		FROM subtasks
		WHERE id = $1 AND deleted_at IS NULL
	`

	var taskID uuid.UUID
	err := r.pool.QueryRow(ctx, query, subtaskID).Scan(&taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, entity.ErrSubtaskNotFound
		}
		return uuid.Nil, fmt.Errorf("failed to find parent task ID: %w", err)
	}

	return taskID, nil
}

// Update actualiza una subtarea existente
func (r *SubtaskRepository) Update(ctx context.Context, subtask *entity.Subtask) error {
	query := `
		UPDATE subtasks
		SET name = $2, state = $3, start_date = $4, end_date = $5, updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		subtask.ID,
		subtask.Name,
		subtask.State.String(),
		subtask.StartDate,
		subtask.EndDate,
		subtask.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update subtask: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrSubtaskNotFound
	}

	return nil
}

// Delete marca una subtarea como eliminada (soft delete)
func (r *SubtaskRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy string) error {
	query := `
		UPDATE subtasks
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subtask: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrSubtaskNotFound
	}

	return nil
}

// FindByTaskID retorna todas las subtareas de una tarea específica
func (r *SubtaskRepository) FindByTaskID(ctx context.Context, taskID uuid.UUID, includeDeleted bool) ([]*entity.Subtask, error) {
	query := `
		SELECT id, name, state, start_date, end_date, created_at, updated_at, deleted_at
		FROM subtasks
		WHERE task_id = $1
	`

	if !includeDeleted {
		query += " AND deleted_at IS NULL"
	}

	query += " ORDER BY created_at ASC"

	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subtasks: %w", err)
	}
	defer rows.Close()

	var subtasks []*entity.Subtask
	for rows.Next() {
		var subtask entity.Subtask
		var state string

		err := rows.Scan(
			&subtask.ID,
			&subtask.Name,
			&state,
			&subtask.StartDate,
			&subtask.EndDate,
			&subtask.CreatedAt,
			&subtask.UpdatedAt,
			&subtask.DeletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan subtask: %w", err)
		}

		subtask.State = entity.State(state)
		subtasks = append(subtasks, &subtask)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subtasks: %w", err)
	}

	return subtasks, nil
}

// DeleteByTaskID marca todas las subtareas de una tarea como eliminadas
func (r *SubtaskRepository) DeleteByTaskID(ctx context.Context, taskID uuid.UUID, deletedBy string) error {
	query := `
		UPDATE subtasks
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE task_id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete subtasks by task ID: %w", err)
	}

	return nil
}
