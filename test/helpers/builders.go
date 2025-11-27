package helpers

import (
	"time"

	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
)

// TaskBuilder facilita la creación de tareas para tests.
type TaskBuilder struct {
	task *entity.Task
}

// NewTaskBuilder crea un builder de tareas con valores por defecto.
func NewTaskBuilder() *TaskBuilder {
	task, _ := entity.NewTask("Test Task", "test-user")
	return &TaskBuilder{task: task}
}

// WithID asigna un ID específico.
func (b *TaskBuilder) WithID(id uuid.UUID) *TaskBuilder {
	b.task.ID = id
	return b
}

// WithName asigna un nombre específico.
func (b *TaskBuilder) WithName(name string) *TaskBuilder {
	b.task.Name = name
	return b
}

// WithState asigna un estado específico.
func (b *TaskBuilder) WithState(state entity.State) *TaskBuilder {
	b.task.State = state
	return b
}

// WithCreatedBy asigna el creador.
func (b *TaskBuilder) WithCreatedBy(user string) *TaskBuilder {
	b.task.CreatedBy = user
	return b
}

// WithStartDate asigna fecha de inicio.
func (b *TaskBuilder) WithStartDate(date time.Time) *TaskBuilder {
	b.task.StartDate = &date
	return b
}

// WithEndDate asigna fecha de fin.
func (b *TaskBuilder) WithEndDate(date time.Time) *TaskBuilder {
	b.task.EndDate = &date
	return b
}

// WithSubtask añade una subtarea.
func (b *TaskBuilder) WithSubtask(subtask *entity.Subtask) *TaskBuilder {
	b.task.AddSubtask(subtask)
	return b
}

// Build construye la tarea.
func (b *TaskBuilder) Build() *entity.Task {
	return b.task
}

// SubtaskBuilder facilita la creación de subtareas para tests.
type SubtaskBuilder struct {
	subtask *entity.Subtask
}

// NewSubtaskBuilder crea un builder de subtareas con valores por defecto.
func NewSubtaskBuilder() *SubtaskBuilder {
	subtask, _ := entity.NewSubtask("Test Subtask")
	return &SubtaskBuilder{subtask: subtask}
}

// WithID asigna un ID específico.
func (s *SubtaskBuilder) WithID(id uuid.UUID) *SubtaskBuilder {
	s.subtask.ID = id
	return s
}

// WithName asigna un nombre específico.
func (s *SubtaskBuilder) WithName(name string) *SubtaskBuilder {
	s.subtask.Name = name
	return s
}

// WithState asigna un estado específico.
func (s *SubtaskBuilder) WithState(state entity.State) *SubtaskBuilder {
	s.subtask.State = state
	return s
}

// WithStartDate asigna fecha de inicio.
func (s *SubtaskBuilder) WithStartDate(date time.Time) *SubtaskBuilder {
	s.subtask.StartDate = &date
	return s
}

// WithEndDate asigna fecha de fin.
func (s *SubtaskBuilder) WithEndDate(date time.Time) *SubtaskBuilder {
	s.subtask.EndDate = &date
	return s
}

// Build construye la subtarea.
func (s *SubtaskBuilder) Build() *entity.Subtask {
	return s.subtask
}
