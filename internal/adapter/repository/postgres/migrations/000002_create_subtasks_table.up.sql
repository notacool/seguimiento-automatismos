-- Create subtasks table
CREATE TABLE IF NOT EXISTS subtasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    name VARCHAR(256) NOT NULL CHECK (char_length(name) > 0 AND name ~ '^[a-zA-Z0-9 _-]+$'),
    state VARCHAR(50) NOT NULL CHECK (state IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED')),

    -- Timestamps
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT valid_subtask_dates CHECK (
        (start_date IS NULL OR start_date >= created_at) AND
        (end_date IS NULL OR (start_date IS NOT NULL AND end_date >= start_date))
    ),
    CONSTRAINT valid_subtask_final_state_dates CHECK (
        (state IN ('COMPLETED', 'FAILED', 'CANCELLED') AND end_date IS NOT NULL) OR
        (state NOT IN ('COMPLETED', 'FAILED', 'CANCELLED'))
    )
);

-- Create indexes for common queries
CREATE INDEX idx_subtasks_task_id ON subtasks(task_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_subtasks_state ON subtasks(state) WHERE deleted_at IS NULL;
CREATE INDEX idx_subtasks_created_at ON subtasks(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_subtasks_deleted_at ON subtasks(deleted_at) WHERE deleted_at IS NOT NULL;

-- Create composite index for task + state queries
CREATE INDEX idx_subtasks_task_state ON subtasks(task_id, state) WHERE deleted_at IS NULL;

-- Create index for soft delete cleanup (pg_cron job)
CREATE INDEX idx_subtasks_cleanup ON subtasks(deleted_at)
    WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '30 days';

-- Add comments to table
COMMENT ON TABLE subtasks IS 'Subtasks table for granular tracking of task steps';
COMMENT ON COLUMN subtasks.id IS 'Unique identifier (UUID)';
COMMENT ON COLUMN subtasks.task_id IS 'Foreign key to parent task';
COMMENT ON COLUMN subtasks.state IS 'Current subtask state: PENDING, IN_PROGRESS, COMPLETED, FAILED, CANCELLED';
COMMENT ON COLUMN subtasks.start_date IS 'Date when subtask transitioned to IN_PROGRESS';
COMMENT ON COLUMN subtasks.end_date IS 'Date when subtask reached a final state';
COMMENT ON COLUMN subtasks.deleted_at IS 'Soft delete timestamp (records purged after 30 days)';
