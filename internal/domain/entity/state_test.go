package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		state State
		want  bool
	}{
		{
			name:  "PENDING is valid",
			state: StatePending,
			want:  true,
		},
		{
			name:  "IN_PROGRESS is valid",
			state: StateInProgress,
			want:  true,
		},
		{
			name:  "COMPLETED is valid",
			state: StateCompleted,
			want:  true,
		},
		{
			name:  "FAILED is valid",
			state: StateFailed,
			want:  true,
		},
		{
			name:  "CANCELLED is valid",
			state: StateCancelled,
			want:  true,
		},
		{
			name:  "invalid state",
			state: State("INVALID"),
			want:  false,
		},
		{
			name:  "empty state",
			state: State(""),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestState_IsFinal(t *testing.T) {
	tests := []struct {
		name  string
		state State
		want  bool
	}{
		{
			name:  "PENDING is not final",
			state: StatePending,
			want:  false,
		},
		{
			name:  "IN_PROGRESS is not final",
			state: StateInProgress,
			want:  false,
		},
		{
			name:  "COMPLETED is final",
			state: StateCompleted,
			want:  true,
		},
		{
			name:  "FAILED is final",
			state: StateFailed,
			want:  true,
		},
		{
			name:  "CANCELLED is final",
			state: StateCancelled,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.IsFinal()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		name  string
		state State
		want  string
	}{
		{
			name:  "PENDING to string",
			state: StatePending,
			want:  "PENDING",
		},
		{
			name:  "IN_PROGRESS to string",
			state: StateInProgress,
			want:  "IN_PROGRESS",
		},
		{
			name:  "COMPLETED to string",
			state: StateCompleted,
			want:  "COMPLETED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
