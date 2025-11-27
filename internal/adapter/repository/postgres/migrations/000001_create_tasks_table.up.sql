-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(256) NOT NULL CHECK (char_length(name) > 0 AND name ~ '^[a-zA-Z0-9 _-]+$'),
    state VARCHAR(50) NOT NULL CHECK (state IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED')),

    -- Audit fields
    created_by VARCHAR(256) NOT NULL CHECK (char_length(created_by) > 0),
    updated_by VARCHAR(256),

    -- Timestamps
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT valid_dates CHECK (
        (start_date IS NULL OR start_date >= created_at) AND
        (end_date IS NULL OR (start_date IS NOT NULL AND end_date >= start_date))
    ),
    CONSTRAINT valid_final_state_dates CHECK (
        (state IN ('COMPLETED', 'FAILED', 'CANCELLED') AND end_date IS NOT NULL) OR
        (state NOT IN ('COMPLETED', 'FAILED', 'CANCELLED'))
    )
);

-- Create index for common queries
CREATE INDEX idx_tasks_state ON tasks(state) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_name ON tasks(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_deleted_at ON tasks(deleted_at) WHERE deleted_at IS NOT NULL;

-- Create index for soft delete cleanup (pg_cron job)
CREATE INDEX idx_tasks_cleanup ON tasks(deleted_at)
    WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '30 days';

-- Add comment to table
COMMENT ON TABLE tasks IS 'Main tasks table for automation process tracking';
COMMENT ON COLUMN tasks.id IS 'Unique identifier (UUID)';
COMMENT ON COLUMN tasks.state IS 'Current task state: PENDING, IN_PROGRESS, COMPLETED, FAILED, CANCELLED';
COMMENT ON COLUMN tasks.created_by IS 'Team or person who created the task';
COMMENT ON COLUMN tasks.updated_by IS 'Team or person who last updated the task';
COMMENT ON COLUMN tasks.start_date IS 'Date when task transitioned to IN_PROGRESS';
COMMENT ON COLUMN tasks.end_date IS 'Date when task reached a final state';
COMMENT ON COLUMN tasks.deleted_at IS 'Soft delete timestamp (records purged after 30 days)';
