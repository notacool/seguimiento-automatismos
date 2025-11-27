package entity

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// Subtask representa una subtarea dentro de una tarea de automatización
type Subtask struct {
	ID        uuid.UUID
	Name      string
	State     State
	StartDate *time.Time
	EndDate   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9 _-]+$`)

// NewSubtask crea una nueva subtarea con validaciones
func NewSubtask(name string) (*Subtask, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Subtask{
		ID:        uuid.New(),
		Name:      name,
		State:     StatePending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// validateName valida que el nombre cumpla con las reglas
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidName)
	}
	if len(name) > 256 {
		return fmt.Errorf("%w: name exceeds 256 characters", ErrInvalidName)
	}
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("%w: name contains invalid characters", ErrInvalidName)
	}
	return nil
}

// IsDeleted verifica si la subtarea está eliminada
func (s *Subtask) IsDeleted() bool {
	return s.DeletedAt != nil
}

// SetStartDate asigna la fecha de inicio si no está ya asignada
func (s *Subtask) SetStartDate() {
	if s.StartDate == nil {
		now := time.Now()
		s.StartDate = &now
	}
}

// SetEndDate asigna la fecha de finalización si no está ya asignada
func (s *Subtask) SetEndDate() {
	if s.EndDate == nil {
		now := time.Now()
		s.EndDate = &now
	}
}

// Delete marca la subtarea como eliminada (soft delete)
func (s *Subtask) Delete() {
	if s.DeletedAt == nil {
		now := time.Now()
		s.DeletedAt = &now
		s.UpdatedAt = now
	}
}
