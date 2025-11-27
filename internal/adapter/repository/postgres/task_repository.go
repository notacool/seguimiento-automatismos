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

// TaskRepository implementa el repositorio de tareas usando PostgreSQL
type TaskRepository struct {
	pool *pgxpool.Pool
}

// NewTaskRepository crea una nueva instancia del repositorio de tareas
func NewTaskRepository(pool *pgxpool.Pool) repository.TaskRepository {
	return &TaskRepository{pool: pool}
}

// Create crea una nueva tarea en la base de datos
func (r *TaskRepository) Create(ctx context.Context, task *entity.Task) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Insert task
	queryTask := `
		INSERT INTO tasks (id, name, state, created_by, updated_by, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = tx.Exec(ctx, queryTask,
		task.ID,
		task.Name,
		task.State.String(),
		task.CreatedBy,
		task.UpdatedBy,
		task.StartDate,
		task.EndDate,
		task.CreatedAt,
		task.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Insert subtasks
	if len(task.Subtasks) > 0 {
		querySubtask := `
			INSERT INTO subtasks (id, task_id, name, state, start_date, end_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`

		for _, subtask := range task.Subtasks {
			_, err = tx.Exec(ctx, querySubtask,
				subtask.ID,
				task.ID,
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
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Update actualiza una tarea existente en la base de datos
func (r *TaskRepository) Update(ctx context.Context, task *entity.Task) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Update task
	queryTask := `
		UPDATE tasks
		SET name = $2, state = $3, updated_by = $4, start_date = $5, end_date = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := tx.Exec(ctx, queryTask,
		task.ID,
		task.Name,
		task.State.String(),
		task.UpdatedBy,
		task.StartDate,
		task.EndDate,
		task.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrTaskNotFound
	}

	// Update/insert subtasks
	if len(task.Subtasks) > 0 {
		for _, subtask := range task.Subtasks {
			// Check if subtask exists
			var exists bool
			err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM subtasks WHERE id = $1)", subtask.ID).Scan(&exists)
			if err != nil {
				return fmt.Errorf("failed to check subtask existence: %w", err)
			}

			if exists {
				// Update existing subtask
				_, err = tx.Exec(ctx, `
					UPDATE subtasks
					SET name = $2, state = $3, start_date = $4, end_date = $5, updated_at = $6
					WHERE id = $1
				`, subtask.ID, subtask.Name, subtask.State.String(), subtask.StartDate, subtask.EndDate, subtask.UpdatedAt)
			} else {
				// Insert new subtask
				_, err = tx.Exec(ctx, `
					INSERT INTO subtasks (id, task_id, name, state, start_date, end_date, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				`, subtask.ID, task.ID, subtask.Name, subtask.State.String(), subtask.StartDate, subtask.EndDate, subtask.CreatedAt, subtask.UpdatedAt)
			}

			if err != nil {
				return fmt.Errorf("failed to update/insert subtask: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID busca una tarea por su ID
func (r *TaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Task, error) {
	query := `
		SELECT id, name, state, created_by, updated_by, start_date, end_date, created_at, updated_at, deleted_at
		FROM tasks
		WHERE id = $1 AND deleted_at IS NULL
	`

	var task entity.Task
	var state string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Name,
		&state,
		&task.CreatedBy,
		&task.UpdatedBy,
		&task.StartDate,
		&task.EndDate,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to find task by ID: %w", err)
	}

	// Parse state
	task.State = entity.State(state)

	// Load subtasks
	subtasks, err := r.loadSubtasks(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load subtasks: %w", err)
	}
	task.Subtasks = subtasks

	return &task, nil
}

// FindAll retorna todas las tareas con paginación y filtros opcionales
func (r *TaskRepository) FindAll(ctx context.Context, filters repository.TaskFilters) (*repository.TaskListResult, error) {
	// Build query with filters
	query, countQuery, args := r.buildFindAllQuery(filters)

	// Get total count
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Get tasks
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*entity.Task
	for rows.Next() {
		var task entity.Task
		var state string

		err := rows.Scan(
			&task.ID,
			&task.Name,
			&state,
			&task.CreatedBy,
			&task.UpdatedBy,
			&task.StartDate,
			&task.EndDate,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DeletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		task.State = entity.State(state)

		// Load subtasks for each task
		subtasks, err := r.loadSubtasks(ctx, task.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load subtasks for task %s: %w", task.ID, err)
		}
		task.Subtasks = subtasks

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	// Calculate total pages
	totalPages := int(total) / filters.Limit
	if int(total)%filters.Limit > 0 {
		totalPages++
	}

	return &repository.TaskListResult{
		Tasks:      tasks,
		Total:      int(total),
		Page:       filters.Page,
		Limit:      filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// Delete marca una tarea como eliminada (soft delete)
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy string) error {
	query := `
		UPDATE tasks
		SET deleted_at = NOW(), updated_by = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, id, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrTaskNotFound
	}

	return nil
}

// HardDelete elimina permanentemente tareas soft-deleted hace más de 30 días
func (r *TaskRepository) HardDelete(ctx context.Context) (int, error) {
	query := `
		DELETE FROM tasks
		WHERE deleted_at IS NOT NULL
		  AND deleted_at < NOW() - INTERVAL '30 days'
	`

	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to hard delete tasks: %w", err)
	}

	return int(result.RowsAffected()), nil
}

// buildFindAllQuery construye la query de búsqueda con filtros
func (r *TaskRepository) buildFindAllQuery(filters repository.TaskFilters) (string, string, []interface{}) {
	baseQuery := `
		SELECT id, name, state, created_by, updated_by, start_date, end_date, created_at, updated_at, deleted_at
		FROM tasks
		WHERE deleted_at IS NULL
	`

	countQuery := `
		SELECT COUNT(*)
		FROM tasks
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argIndex := 1

	// Add state filter
	if filters.State != nil {
		filter := fmt.Sprintf(" AND state = $%d", argIndex)
		baseQuery += filter
		countQuery += filter
		args = append(args, filters.State.String())
		argIndex++
	}

	// Add name filter (case-insensitive partial match)
	if filters.Name != nil {
		filter := fmt.Sprintf(" AND LOWER(name) LIKE LOWER($%d)", argIndex)
		baseQuery += filter
		countQuery += filter
		args = append(args, "%"+*filters.Name+"%")
		argIndex++
	}

	// Add ordering
	baseQuery += " ORDER BY created_at DESC"

	// Add pagination
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filters.Limit, filters.Offset)

	return baseQuery, countQuery, args
}

// loadSubtasks carga las subtareas de una tarea
func (r *TaskRepository) loadSubtasks(ctx context.Context, taskID uuid.UUID) ([]*entity.Subtask, error) {
	query := `
		SELECT id, name, state, start_date, end_date, created_at, updated_at, deleted_at
		FROM subtasks
		WHERE task_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

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
