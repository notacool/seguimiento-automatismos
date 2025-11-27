package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	tests := []struct {
		name        string
		taskName    string
		createdBy   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid task",
			taskName:  "Valid Task",
			createdBy: "Equipo DevOps",
			wantErr:   false,
		},
		{
			name:        "empty name",
			taskName:    "",
			createdBy:   "Equipo DevOps",
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "invalid name characters",
			taskName:    "Invalid@Name",
			createdBy:   "Equipo DevOps",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "name too long",
			taskName:    strings.Repeat("a", 257),
			createdBy:   "Equipo DevOps",
			wantErr:     true,
			errContains: "exceeds 256 characters",
		},
		{
			name:        "empty created_by",
			taskName:    "Valid Task",
			createdBy:   "",
			wantErr:     true,
			errContains: "created_by is required",
		},
		{
			name:        "created_by too long",
			taskName:    "Valid Task",
			createdBy:   strings.Repeat("a", 257),
			wantErr:     true,
			errContains: "exceeds 256 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewTask(tt.taskName, tt.createdBy)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.NotEqual(t, uuid.Nil, task.ID)
				assert.Equal(t, tt.taskName, task.Name)
				assert.Equal(t, tt.createdBy, task.CreatedBy)
				assert.Equal(t, tt.createdBy, task.UpdatedBy)
				assert.Equal(t, StatePending, task.State)
				assert.Empty(t, task.Subtasks)
				assert.False(t, task.CreatedAt.IsZero())
				assert.False(t, task.UpdatedAt.IsZero())
				assert.Nil(t, task.StartDate)
				assert.Nil(t, task.EndDate)
				assert.Nil(t, task.DeletedAt)
			}
		})
	}
}

func TestTask_AddSubtask(t *testing.T) {
	task, _ := NewTask("Test Task", "Team A")
	subtask1, _ := NewSubtask("Subtask 1")
	subtask2, _ := NewSubtask("Subtask 2")

	originalUpdatedAt := task.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	task.AddSubtask(subtask1)
	assert.Len(t, task.Subtasks, 1)
	assert.Equal(t, subtask1, task.Subtasks[0])
	assert.True(t, task.UpdatedAt.After(originalUpdatedAt))

	task.AddSubtask(subtask2)
	assert.Len(t, task.Subtasks, 2)
	assert.Equal(t, subtask2, task.Subtasks[1])
}

func TestTask_IsDeleted(t *testing.T) {
	t.Run("not deleted", func(t *testing.T) {
		task, _ := NewTask("Test", "Team A")
		assert.False(t, task.IsDeleted())
	})

	t.Run("is deleted", func(t *testing.T) {
		task, _ := NewTask("Test", "Team A")
		now := time.Now()
		task.DeletedAt = &now
		assert.True(t, task.IsDeleted())
	})
}

func TestTask_SetStartDate(t *testing.T) {
	t.Run("set start date when nil", func(t *testing.T) {
		task, _ := NewTask("Test", "Team A")
		assert.Nil(t, task.StartDate)

		beforeSet := time.Now()
		task.SetStartDate()
		afterSet := time.Now()

		require.NotNil(t, task.StartDate)
		assert.True(t, task.StartDate.After(beforeSet) || task.StartDate.Equal(beforeSet))
		assert.True(t, task.StartDate.Before(afterSet) || task.StartDate.Equal(afterSet))
	})

	t.Run("does not override existing start date", func(t *testing.T) {
		task, _ := NewTask("Test", "Team A")
		originalDate := time.Now().Add(-1 * time.Hour)
		task.StartDate = &originalDate

		task.SetStartDate()
		assert.Equal(t, originalDate, *task.StartDate)
	})
}

func TestTask_SetEndDate(t *testing.T) {
	t.Run("set end date when nil", func(t *testing.T) {
		task, _ := NewTask("Test", "Team A")
		assert.Nil(t, task.EndDate)

		beforeSet := time.Now()
		task.SetEndDate()
		afterSet := time.Now()

		require.NotNil(t, task.EndDate)
		assert.True(t, task.EndDate.After(beforeSet) || task.EndDate.Equal(beforeSet))
		assert.True(t, task.EndDate.Before(afterSet) || task.EndDate.Equal(afterSet))
	})

	t.Run("does not override existing end date", func(t *testing.T) {
		task, _ := NewTask("Test", "Team A")
		originalDate := time.Now().Add(-1 * time.Hour)
		task.EndDate = &originalDate

		task.SetEndDate()
		assert.Equal(t, originalDate, *task.EndDate)
	})
}

func TestTask_Delete(t *testing.T) {
	t.Run("delete task", func(t *testing.T) {
		task, _ := NewTask("Test", "Team A")
		originalUpdatedAt := task.UpdatedAt
		assert.Nil(t, task.DeletedAt)

		time.Sleep(1 * time.Millisecond)
		task.Delete()

		require.NotNil(t, task.DeletedAt)
		assert.True(t, task.IsDeleted())
		assert.True(t, task.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("delete already deleted task does not change date", func(t *testing.T) {
		task, _ := NewTask("Test", "Team A")
		task.Delete()
		originalDeletedAt := *task.DeletedAt
		originalUpdatedAt := task.UpdatedAt

		time.Sleep(1 * time.Millisecond)
		task.Delete()

		assert.Equal(t, originalDeletedAt, *task.DeletedAt)
		assert.Equal(t, originalUpdatedAt, task.UpdatedAt)
	})
}

func TestTask_PropagateStateToSubtasks(t *testing.T) {
	t.Run("propagate COMPLETED state to all subtasks", func(t *testing.T) {
		task, _ := NewTask("Test Task", "Team A")
		subtask1, _ := NewSubtask("Subtask 1")
		subtask2, _ := NewSubtask("Subtask 2")
		subtask1.State = StateInProgress
		subtask2.State = StatePending

		task.AddSubtask(subtask1)
		task.AddSubtask(subtask2)

		task.State = StateCompleted
		task.PropagateStateToSubtasks()

		assert.Equal(t, StateCompleted, subtask1.State)
		assert.Equal(t, StateCompleted, subtask2.State)
		assert.NotNil(t, subtask1.EndDate)
		assert.NotNil(t, subtask2.EndDate)
	})

	t.Run("propagate FAILED state to all subtasks", func(t *testing.T) {
		task, _ := NewTask("Test Task", "Team A")
		subtask1, _ := NewSubtask("Subtask 1")
		task.AddSubtask(subtask1)

		task.State = StateFailed
		task.PropagateStateToSubtasks()

		assert.Equal(t, StateFailed, subtask1.State)
		assert.NotNil(t, subtask1.EndDate)
	})

	t.Run("does not propagate non-final states", func(t *testing.T) {
		task, _ := NewTask("Test Task", "Team A")
		subtask1, _ := NewSubtask("Subtask 1")
		task.AddSubtask(subtask1)

		task.State = StateInProgress
		task.PropagateStateToSubtasks()

		assert.Equal(t, StatePending, subtask1.State)
	})

	t.Run("skips deleted subtasks", func(t *testing.T) {
		task, _ := NewTask("Test Task", "Team A")
		subtask1, _ := NewSubtask("Subtask 1")
		subtask1.Delete()
		task.AddSubtask(subtask1)

		task.State = StateCompleted
		task.PropagateStateToSubtasks()

		// Estado de subtarea eliminada no debería cambiar
		assert.Equal(t, StatePending, subtask1.State)
	})
}

func TestTask_UpdateState(t *testing.T) {
	t.Run("update to IN_PROGRESS sets start date", func(t *testing.T) {
		task, _ := NewTask("Test Task", "Team A")
		assert.Nil(t, task.StartDate)

		err := task.UpdateState(StateInProgress, "Team B")
		require.NoError(t, err)

		assert.Equal(t, StateInProgress, task.State)
		assert.Equal(t, "Team B", task.UpdatedBy)
		assert.NotNil(t, task.StartDate)
		assert.Nil(t, task.EndDate)
	})

	t.Run("update to COMPLETED sets end date and propagates", func(t *testing.T) {
		task, _ := NewTask("Test Task", "Team A")
		subtask1, _ := NewSubtask("Subtask 1")
		task.AddSubtask(subtask1)

		err := task.UpdateState(StateCompleted, "Team A")
		require.NoError(t, err)

		assert.Equal(t, StateCompleted, task.State)
		assert.NotNil(t, task.EndDate)
		assert.Equal(t, StateCompleted, subtask1.State)
		assert.NotNil(t, subtask1.EndDate)
	})

	t.Run("update to FAILED sets end date and propagates", func(t *testing.T) {
		task, _ := NewTask("Test Task", "Team A")
		subtask1, _ := NewSubtask("Subtask 1")
		task.AddSubtask(subtask1)

		err := task.UpdateState(StateFailed, "Team A")
		require.NoError(t, err)

		assert.Equal(t, StateFailed, task.State)
		assert.NotNil(t, task.EndDate)
		assert.Equal(t, StateFailed, subtask1.State)
	})

	t.Run("invalid state returns error", func(t *testing.T) {
		task, _ := NewTask("Test Task", "Team A")

		err := task.UpdateState(State("INVALID"), "Team A")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid state")
	})

	t.Run("empty updated_by returns error", func(t *testing.T) {
		task, _ := NewTask("Test Task", "Team A")

		err := task.UpdateState(StateInProgress, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "updated_by is required")
	})

	t.Run("updates updated_at timestamp", func(t *testing.T) {
		task, _ := NewTask("Test Task", "Team A")
		originalUpdatedAt := task.UpdatedAt

		time.Sleep(1 * time.Millisecond)
		err := task.UpdateState(StateInProgress, "Team B")
		require.NoError(t, err)

		assert.True(t, task.UpdatedAt.After(originalUpdatedAt))
	})
}
