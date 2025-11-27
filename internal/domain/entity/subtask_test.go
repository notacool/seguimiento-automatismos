package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubtask(t *testing.T) {
	tests := []struct {
		name        string
		subtaskName string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid name with alphanumeric",
			subtaskName: "ValidTask123",
			wantErr:     false,
		},
		{
			name:        "valid name with spaces",
			subtaskName: "Valid Task Name",
			wantErr:     false,
		},
		{
			name:        "valid name with hyphens",
			subtaskName: "valid-task-name",
			wantErr:     false,
		},
		{
			name:        "valid name with underscores",
			subtaskName: "valid_task_name",
			wantErr:     false,
		},
		{
			name:        "valid name mixed",
			subtaskName: "Task-2024_Test Process",
			wantErr:     false,
		},
		{
			name:        "empty name",
			subtaskName: "",
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "name too long",
			subtaskName: string(make([]byte, 257)),
			wantErr:     true,
			errContains: "exceeds 256 characters",
		},
		{
			name:        "name with invalid characters @",
			subtaskName: "Invalid@Name",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "name with invalid characters #",
			subtaskName: "Invalid#Name",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "name with special chars",
			subtaskName: "Invalid!Name$",
			wantErr:     true,
			errContains: "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subtask, err := NewSubtask(tt.subtaskName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, subtask)
			} else {
				require.NoError(t, err)
				require.NotNil(t, subtask)
				assert.NotEqual(t, uuid.Nil, subtask.ID)
				assert.Equal(t, tt.subtaskName, subtask.Name)
				assert.Equal(t, StatePending, subtask.State)
				assert.False(t, subtask.CreatedAt.IsZero())
				assert.False(t, subtask.UpdatedAt.IsZero())
				assert.Nil(t, subtask.StartDate)
				assert.Nil(t, subtask.EndDate)
				assert.Nil(t, subtask.DeletedAt)
			}
		})
	}
}

func TestSubtask_IsDeleted(t *testing.T) {
	t.Run("not deleted", func(t *testing.T) {
		subtask, _ := NewSubtask("Test")
		assert.False(t, subtask.IsDeleted())
	})

	t.Run("is deleted", func(t *testing.T) {
		subtask, _ := NewSubtask("Test")
		now := time.Now()
		subtask.DeletedAt = &now
		assert.True(t, subtask.IsDeleted())
	})
}

func TestSubtask_SetStartDate(t *testing.T) {
	t.Run("set start date when nil", func(t *testing.T) {
		subtask, _ := NewSubtask("Test")
		assert.Nil(t, subtask.StartDate)

		beforeSet := time.Now()
		subtask.SetStartDate()
		afterSet := time.Now()

		require.NotNil(t, subtask.StartDate)
		assert.True(t, subtask.StartDate.After(beforeSet) || subtask.StartDate.Equal(beforeSet))
		assert.True(t, subtask.StartDate.Before(afterSet) || subtask.StartDate.Equal(afterSet))
	})

	t.Run("does not override existing start date", func(t *testing.T) {
		subtask, _ := NewSubtask("Test")
		originalDate := time.Now().Add(-1 * time.Hour)
		subtask.StartDate = &originalDate

		subtask.SetStartDate()

		assert.Equal(t, originalDate, *subtask.StartDate)
	})
}

func TestSubtask_SetEndDate(t *testing.T) {
	t.Run("set end date when nil", func(t *testing.T) {
		subtask, _ := NewSubtask("Test")
		assert.Nil(t, subtask.EndDate)

		beforeSet := time.Now()
		subtask.SetEndDate()
		afterSet := time.Now()

		require.NotNil(t, subtask.EndDate)
		assert.True(t, subtask.EndDate.After(beforeSet) || subtask.EndDate.Equal(beforeSet))
		assert.True(t, subtask.EndDate.Before(afterSet) || subtask.EndDate.Equal(afterSet))
	})

	t.Run("does not override existing end date", func(t *testing.T) {
		subtask, _ := NewSubtask("Test")
		originalDate := time.Now().Add(-1 * time.Hour)
		subtask.EndDate = &originalDate

		subtask.SetEndDate()

		assert.Equal(t, originalDate, *subtask.EndDate)
	})
}

func TestSubtask_Delete(t *testing.T) {
	t.Run("delete subtask", func(t *testing.T) {
		subtask, _ := NewSubtask("Test")
		originalUpdatedAt := subtask.UpdatedAt
		assert.Nil(t, subtask.DeletedAt)

		time.Sleep(1 * time.Millisecond) // Asegurar que UpdatedAt cambie
		subtask.Delete()

		require.NotNil(t, subtask.DeletedAt)
		assert.True(t, subtask.IsDeleted())
		assert.True(t, subtask.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("delete already deleted subtask does not change date", func(t *testing.T) {
		subtask, _ := NewSubtask("Test")
		subtask.Delete()
		originalDeletedAt := *subtask.DeletedAt
		originalUpdatedAt := subtask.UpdatedAt

		time.Sleep(1 * time.Millisecond)
		subtask.Delete()

		assert.Equal(t, originalDeletedAt, *subtask.DeletedAt)
		assert.Equal(t, originalUpdatedAt, subtask.UpdatedAt)
	})
}
