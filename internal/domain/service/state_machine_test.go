package service

import (
	"testing"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateMachine_CanTransition(t *testing.T) {
	sm := NewStateMachine()

	tests := []struct {
		name     string
		from     entity.State
		to       entity.State
		expected bool
	}{
		// Transiciones válidas desde PENDING
		{
			name:     "PENDING to IN_PROGRESS is valid",
			from:     entity.StatePending,
			to:       entity.StateInProgress,
			expected: true,
		},
		{
			name:     "PENDING to CANCELLED is valid",
			from:     entity.StatePending,
			to:       entity.StateCancelled,
			expected: true,
		},
		// Transiciones inválidas desde PENDING
		{
			name:     "PENDING to COMPLETED is invalid",
			from:     entity.StatePending,
			to:       entity.StateCompleted,
			expected: false,
		},
		{
			name:     "PENDING to FAILED is invalid",
			from:     entity.StatePending,
			to:       entity.StateFailed,
			expected: false,
		},
		{
			name:     "PENDING to PENDING is invalid (same state)",
			from:     entity.StatePending,
			to:       entity.StatePending,
			expected: false,
		},
		// Transiciones válidas desde IN_PROGRESS
		{
			name:     "IN_PROGRESS to COMPLETED is valid",
			from:     entity.StateInProgress,
			to:       entity.StateCompleted,
			expected: true,
		},
		{
			name:     "IN_PROGRESS to FAILED is valid",
			from:     entity.StateInProgress,
			to:       entity.StateFailed,
			expected: true,
		},
		// Transiciones inválidas desde IN_PROGRESS
		{
			name:     "IN_PROGRESS to PENDING is invalid",
			from:     entity.StateInProgress,
			to:       entity.StatePending,
			expected: false,
		},
		{
			name:     "IN_PROGRESS to CANCELLED is invalid",
			from:     entity.StateInProgress,
			to:       entity.StateCancelled,
			expected: false,
		},
		{
			name:     "IN_PROGRESS to IN_PROGRESS is invalid (same state)",
			from:     entity.StateInProgress,
			to:       entity.StateInProgress,
			expected: false,
		},
		// Estados finales no permiten transiciones
		{
			name:     "COMPLETED to PENDING is invalid (final state)",
			from:     entity.StateCompleted,
			to:       entity.StatePending,
			expected: false,
		},
		{
			name:     "COMPLETED to IN_PROGRESS is invalid (final state)",
			from:     entity.StateCompleted,
			to:       entity.StateInProgress,
			expected: false,
		},
		{
			name:     "COMPLETED to FAILED is invalid (final state)",
			from:     entity.StateCompleted,
			to:       entity.StateFailed,
			expected: false,
		},
		{
			name:     "COMPLETED to CANCELLED is invalid (final state)",
			from:     entity.StateCompleted,
			to:       entity.StateCancelled,
			expected: false,
		},
		{
			name:     "FAILED to PENDING is invalid (final state)",
			from:     entity.StateFailed,
			to:       entity.StatePending,
			expected: false,
		},
		{
			name:     "FAILED to IN_PROGRESS is invalid (final state)",
			from:     entity.StateFailed,
			to:       entity.StateInProgress,
			expected: false,
		},
		{
			name:     "FAILED to COMPLETED is invalid (final state)",
			from:     entity.StateFailed,
			to:       entity.StateCompleted,
			expected: false,
		},
		{
			name:     "CANCELLED to PENDING is invalid (final state)",
			from:     entity.StateCancelled,
			to:       entity.StatePending,
			expected: false,
		},
		{
			name:     "CANCELLED to IN_PROGRESS is invalid (final state)",
			from:     entity.StateCancelled,
			to:       entity.StateInProgress,
			expected: false,
		},
		{
			name:     "CANCELLED to COMPLETED is invalid (final state)",
			from:     entity.StateCancelled,
			to:       entity.StateCompleted,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.CanTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStateMachine_ValidateTransition(t *testing.T) {
	sm := NewStateMachine()

	tests := []struct {
		name        string
		from        entity.State
		to          entity.State
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid transition PENDING to IN_PROGRESS",
			from:    entity.StatePending,
			to:      entity.StateInProgress,
			wantErr: false,
		},
		{
			name:    "valid transition IN_PROGRESS to COMPLETED",
			from:    entity.StateInProgress,
			to:      entity.StateCompleted,
			wantErr: false,
		},
		{
			name:        "invalid transition PENDING to COMPLETED",
			from:        entity.StatePending,
			to:          entity.StateCompleted,
			wantErr:     true,
			errContains: "invalid state transition from PENDING to COMPLETED",
		},
		{
			name:        "invalid transition from final state COMPLETED",
			from:        entity.StateCompleted,
			to:          entity.StatePending,
			wantErr:     true,
			errContains: "cannot transition from final state COMPLETED",
		},
		{
			name:        "invalid transition from final state FAILED",
			from:        entity.StateFailed,
			to:          entity.StateInProgress,
			wantErr:     true,
			errContains: "cannot transition from final state FAILED",
		},
		{
			name:        "same state transition",
			from:        entity.StatePending,
			to:          entity.StatePending,
			wantErr:     true,
			errContains: "invalid state transition from PENDING to PENDING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.ValidateTransition(tt.from, tt.to)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStateMachine_ValidateTaskStateTransition(t *testing.T) {
	sm := NewStateMachine()

	t.Run("valid task state transition", func(t *testing.T) {
		task, _ := entity.NewTask("Test Task", "Team A")
		task.State = entity.StatePending

		err := sm.ValidateTaskStateTransition(task, entity.StateInProgress)
		require.NoError(t, err)
	})

	t.Run("invalid task state transition", func(t *testing.T) {
		task, _ := entity.NewTask("Test Task", "Team A")
		task.State = entity.StatePending

		err := sm.ValidateTaskStateTransition(task, entity.StateCompleted)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid state transition")
	})

	t.Run("transition from final state", func(t *testing.T) {
		task, _ := entity.NewTask("Test Task", "Team A")
		task.State = entity.StateCompleted

		err := sm.ValidateTaskStateTransition(task, entity.StatePending)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot transition from final state")
	})
}

func TestStateMachine_ValidateSubtaskStateTransition(t *testing.T) {
	sm := NewStateMachine()

	t.Run("valid subtask state transition when parent allows", func(t *testing.T) {
		task, _ := entity.NewTask("Test Task", "Team A")
		task.State = entity.StateInProgress

		subtask, _ := entity.NewSubtask("Subtask 1")
		subtask.State = entity.StatePending
		task.AddSubtask(subtask)

		err := sm.ValidateSubtaskStateTransition(task, subtask, entity.StateInProgress)
		require.NoError(t, err)
	})

	t.Run("invalid: subtask IN_PROGRESS when parent PENDING", func(t *testing.T) {
		task, _ := entity.NewTask("Test Task", "Team A")
		task.State = entity.StatePending

		subtask, _ := entity.NewSubtask("Subtask 1")
		subtask.State = entity.StatePending
		task.AddSubtask(subtask)

		err := sm.ValidateSubtaskStateTransition(task, subtask, entity.StateInProgress)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "subtask cannot have state IN_PROGRESS when parent is PENDING")
	})

	t.Run("invalid: subtask active when parent in final state", func(t *testing.T) {
		task, _ := entity.NewTask("Test Task", "Team A")
		task.State = entity.StateCompleted

		subtask, _ := entity.NewSubtask("Subtask 1")
		subtask.State = entity.StatePending
		task.AddSubtask(subtask)

		err := sm.ValidateSubtaskStateTransition(task, subtask, entity.StateInProgress)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "subtask cannot transition when parent task is in final state COMPLETED")
	})

	t.Run("valid: subtask inherits parent final state", func(t *testing.T) {
		task, _ := entity.NewTask("Test Task", "Team A")
		task.State = entity.StateCompleted

		subtask, _ := entity.NewSubtask("Subtask 1")
		subtask.State = entity.StateInProgress
		task.AddSubtask(subtask)

		err := sm.ValidateSubtaskStateTransition(task, subtask, entity.StateCompleted)
		require.NoError(t, err)
	})

	t.Run("invalid: subtask COMPLETED when parent IN_PROGRESS", func(t *testing.T) {
		task, _ := entity.NewTask("Test Task", "Team A")
		task.State = entity.StateInProgress

		subtask, _ := entity.NewSubtask("Subtask 1")
		subtask.State = entity.StatePending
		task.AddSubtask(subtask)

		err := sm.ValidateSubtaskStateTransition(task, subtask, entity.StateCompleted)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "subtask cannot reach final state COMPLETED when parent is IN_PROGRESS")
	})

	t.Run("invalid: basic transition validation for subtask", func(t *testing.T) {
		task, _ := entity.NewTask("Test Task", "Team A")
		task.State = entity.StateInProgress

		subtask, _ := entity.NewSubtask("Subtask 1")
		subtask.State = entity.StateCompleted
		task.AddSubtask(subtask)

		err := sm.ValidateSubtaskStateTransition(task, subtask, entity.StatePending)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot transition from final state")
	})
}

func TestStateMachine_GetAllowedTransitions(t *testing.T) {
	sm := NewStateMachine()

	tests := []struct {
		name     string
		state    entity.State
		expected []entity.State
	}{
		{
			name:     "allowed transitions from PENDING",
			state:    entity.StatePending,
			expected: []entity.State{entity.StateInProgress, entity.StateCancelled},
		},
		{
			name:     "allowed transitions from IN_PROGRESS",
			state:    entity.StateInProgress,
			expected: []entity.State{entity.StateCompleted, entity.StateFailed},
		},
		{
			name:     "no transitions from COMPLETED",
			state:    entity.StateCompleted,
			expected: []entity.State{},
		},
		{
			name:     "no transitions from FAILED",
			state:    entity.StateFailed,
			expected: []entity.State{},
		},
		{
			name:     "no transitions from CANCELLED",
			state:    entity.StateCancelled,
			expected: []entity.State{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.GetAllowedTransitions(tt.state)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}
