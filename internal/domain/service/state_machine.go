package service

import (
	"fmt"

	"github.com/grupoapi/proces-log/internal/domain/entity"
)

// StateMachine gestiona las transiciones de estado permitidas
type StateMachine struct {
	transitions map[entity.State][]entity.State
}

// NewStateMachine crea una nueva instancia de StateMachine con las reglas de transición
func NewStateMachine() *StateMachine {
	return &StateMachine{
		transitions: map[entity.State][]entity.State{
			entity.StatePending:    {entity.StateInProgress, entity.StateCancelled},
			entity.StateInProgress: {entity.StateCompleted, entity.StateFailed},
			entity.StateCompleted:  {}, // Estado final, sin transiciones
			entity.StateFailed:     {}, // Estado final, sin transiciones
			entity.StateCancelled:  {}, // Estado final, sin transiciones
		},
	}
}

// CanTransition verifica si una transición de estado es válida
func (sm *StateMachine) CanTransition(from, to entity.State) bool {
	// No se permite transición al mismo estado
	if from == to {
		return false
	}

	// Estados finales no permiten transiciones
	if from.IsFinal() {
		return false
	}

	// Verificar si la transición está en las permitidas
	allowedStates, exists := sm.transitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == to {
			return true
		}
	}

	return false
}

// ValidateTransition valida una transición y retorna error descriptivo si no es válida
func (sm *StateMachine) ValidateTransition(from, to entity.State) error {
	// Verificar si es estado final
	if from.IsFinal() {
		return fmt.Errorf("%w: cannot transition from final state %s", entity.ErrInvalidStateTransition, from)
	}

	// Verificar si la transición es válida
	if !sm.CanTransition(from, to) {
		return fmt.Errorf("%w: invalid state transition from %s to %s", entity.ErrInvalidStateTransition, from, to)
	}

	return nil
}

// ValidateTaskStateTransition valida una transición de estado para una tarea
func (sm *StateMachine) ValidateTaskStateTransition(task *entity.Task, newState entity.State) error {
	return sm.ValidateTransition(task.State, newState)
}

// ValidateSubtaskStateTransition valida una transición de estado para una subtarea
// considerando el estado de la tarea padre
func (sm *StateMachine) ValidateSubtaskStateTransition(task *entity.Task, subtask *entity.Subtask, newState entity.State) error {
	// Si la tarea padre está en estado final, la subtarea solo puede heredar ese estado
	if task.State.IsFinal() {
		if newState != task.State {
			return fmt.Errorf("%w: subtask cannot transition when parent task is in final state %s", 
				entity.ErrInconsistentParentChildState, task.State)
		}
		// Validar transición básica solo después de verificar consistencia con padre
		return sm.ValidateTransition(subtask.State, newState)
	}

	// Una subtarea no puede estar en IN_PROGRESS si el padre está PENDING
	if task.State == entity.StatePending && newState == entity.StateInProgress {
		return fmt.Errorf("%w: subtask cannot have state IN_PROGRESS when parent is PENDING", 
			entity.ErrInconsistentParentChildState)
	}

	// Una subtarea no puede alcanzar un estado final si el padre no está en estado final
	if newState.IsFinal() && !task.State.IsFinal() {
		return fmt.Errorf("%w: subtask cannot reach final state %s when parent is %s", 
			entity.ErrInconsistentParentChildState, newState, task.State)
	}

	// Validar la transición básica al final
	return sm.ValidateTransition(subtask.State, newState)
}

// GetAllowedTransitions retorna los estados a los que se puede transicionar desde un estado dado
func (sm *StateMachine) GetAllowedTransitions(state entity.State) []entity.State {
	allowedStates, exists := sm.transitions[state]
	if !exists {
		return []entity.State{}
	}

	// Retornar copia para evitar modificaciones externas
	result := make([]entity.State, len(allowedStates))
	copy(result, allowedStates)
	return result
}
