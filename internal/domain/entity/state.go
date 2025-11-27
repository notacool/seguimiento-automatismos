package entity

// State representa el estado de una tarea o subtarea
type State string

const (
	// StatePending indica que la tarea está pendiente de iniciar
	StatePending State = "PENDING"

	// StateInProgress indica que la tarea está en ejecución
	StateInProgress State = "IN_PROGRESS"

	// StateCompleted indica que la tarea se completó exitosamente (estado final)
	StateCompleted State = "COMPLETED"

	// StateFailed indica que la tarea falló (estado final)
	StateFailed State = "FAILED"

	// StateCancelled indica que la tarea fue cancelada (estado final)
	StateCancelled State = "CANCELLED"
)

// IsValid verifica si el estado es válido
func (s State) IsValid() bool {
	switch s {
	case StatePending, StateInProgress, StateCompleted, StateFailed, StateCancelled:
		return true
	default:
		return false
	}
}

// IsFinal verifica si el estado es final (no permite transiciones)
func (s State) IsFinal() bool {
	return s == StateCompleted || s == StateFailed || s == StateCancelled
}

// String retorna la representación en string del estado
func (s State) String() string {
	return string(s)
}
