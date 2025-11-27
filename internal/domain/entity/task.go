package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Task representa una tarea de automatización
type Task struct {
	ID        uuid.UUID
	Name      string
	State     State
	Subtasks  []*Subtask
	CreatedBy string
	UpdatedBy string
	StartDate *time.Time
	EndDate   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewTask crea una nueva tarea con validaciones
func NewTask(name, createdBy string) (*Task, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if createdBy == "" {
		return nil, fmt.Errorf("%w: created_by is required", ErrMissingRequiredFields)
	}
	if len(createdBy) > 256 {
		return nil, fmt.Errorf("%w: created_by exceeds 256 characters", ErrInvalidName)
	}

	now := time.Now()
	return &Task{
		ID:        uuid.New(),
		Name:      name,
		State:     StatePending,
		Subtasks:  make([]*Subtask, 0),
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// AddSubtask añade una subtarea a la tarea
func (t *Task) AddSubtask(subtask *Subtask) {
	t.Subtasks = append(t.Subtasks, subtask)
	t.UpdatedAt = time.Now()
}

// IsDeleted verifica si la tarea está eliminada
func (t *Task) IsDeleted() bool {
	return t.DeletedAt != nil
}

// SetStartDate asigna la fecha de inicio si no está ya asignada
func (t *Task) SetStartDate() {
	if t.StartDate == nil {
		now := time.Now()
		t.StartDate = &now
	}
}

// SetEndDate asigna la fecha de finalización si no está ya asignada
func (t *Task) SetEndDate() {
	if t.EndDate == nil {
		now := time.Now()
		t.EndDate = &now
	}
}

// Delete marca la tarea como eliminada (soft delete)
func (t *Task) Delete() {
	if t.DeletedAt == nil {
		now := time.Now()
		t.DeletedAt = &now
		t.UpdatedAt = now
	}
}

// PropagateStateToSubtasks propaga el estado a todas las subtareas
// Se usa cuando la tarea llega a un estado final
func (t *Task) PropagateStateToSubtasks() {
	if !t.State.IsFinal() {
		return
	}

	for _, subtask := range t.Subtasks {
		if subtask.IsDeleted() {
			continue
		}
		subtask.State = t.State
		subtask.UpdatedAt = time.Now()

		// Asignar fechas si la subtarea no las tiene
		if t.State == StateInProgress {
			subtask.SetStartDate()
		}
		if t.State.IsFinal() {
			subtask.SetEndDate()
		}
	}
}

// UpdateState actualiza el estado de la tarea y gestiona fechas
func (t *Task) UpdateState(newState State, updatedBy string) error {
	if !newState.IsValid() {
		return fmt.Errorf("%w: invalid state %s", ErrInvalidStateTransition, newState)
	}

	if updatedBy == "" {
		return fmt.Errorf("%w: updated_by is required", ErrMissingRequiredFields)
	}

	// Asignar fechas según el estado
	if newState == StateInProgress {
		t.SetStartDate()
	}
	if newState.IsFinal() {
		t.SetEndDate()
	}

	t.State = newState
	t.UpdatedBy = updatedBy
	t.UpdatedAt = time.Now()

	// Si el estado es final, propagar a subtareas
	if newState.IsFinal() {
		t.PropagateStateToSubtasks()
	}

	return nil
}
