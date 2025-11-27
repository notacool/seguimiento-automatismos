package entity

import "errors"

// Errores de dominio
var (
	// ErrInvalidName indica que el nombre no cumple con las reglas de validación
	ErrInvalidName = errors.New("name must be alphanumeric with spaces, hyphens, or underscores, maximum 256 characters")

	// ErrInvalidStateTransition indica que la transición de estado no es válida
	ErrInvalidStateTransition = errors.New("invalid state transition")

	// ErrInconsistentParentChildState indica que el estado de la subtarea es inconsistente con el padre
	ErrInconsistentParentChildState = errors.New("subtask state inconsistent with parent task state")

	// ErrTaskNotFound indica que la tarea no existe o fue eliminada
	ErrTaskNotFound = errors.New("task not found or deleted")

	// ErrSubtaskNotFound indica que la subtarea no existe o fue eliminada
	ErrSubtaskNotFound = errors.New("subtask not found or deleted")

	// ErrMissingRequiredFields indica que faltan campos requeridos
	ErrMissingRequiredFields = errors.New("missing required fields")

	// ErrDatabaseError indica un error al interactuar con la base de datos
	ErrDatabaseError = errors.New("database error")

	// ErrDatabaseUnavailable indica que la base de datos no está disponible
	ErrDatabaseUnavailable = errors.New("database unavailable")
)
