-- Enable pg_cron extension (requires superuser privileges)
CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for tasks table
CREATE TRIGGER update_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for subtasks table
CREATE TRIGGER update_subtasks_updated_at
    BEFORE UPDATE ON subtasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to cleanup old soft-deleted records
CREATE OR REPLACE FUNCTION cleanup_soft_deleted_records()
RETURNS void AS $$
DECLARE
    deleted_tasks_count INTEGER;
    deleted_subtasks_count INTEGER;
BEGIN
    -- Delete tasks older than 30 days
    DELETE FROM tasks
    WHERE deleted_at IS NOT NULL
      AND deleted_at < NOW() - INTERVAL '30 days';
    GET DIAGNOSTICS deleted_tasks_count = ROW_COUNT;

    -- Delete subtasks older than 30 days
    DELETE FROM subtasks
    WHERE deleted_at IS NOT NULL
      AND deleted_at < NOW() - INTERVAL '30 days';
    GET DIAGNOSTICS deleted_subtasks_count = ROW_COUNT;

    -- Log the cleanup (optional, requires logging table or use RAISE NOTICE)
    RAISE NOTICE 'Cleanup completed: % tasks and % subtasks deleted',
        deleted_tasks_count, deleted_subtasks_count;
END;
$$ LANGUAGE plpgsql;

-- Schedule cleanup job to run daily at 2:00 AM
-- Note: This requires the database user to have permissions to use pg_cron
SELECT cron.schedule(
    'cleanup-soft-deletes',           -- job name
    '0 2 * * *',                      -- cron schedule (daily at 2 AM)
    'SELECT cleanup_soft_deleted_records();'  -- SQL command
);

-- Add comment
COMMENT ON FUNCTION cleanup_soft_deleted_records() IS
    'Permanently deletes tasks and subtasks that have been soft-deleted for more than 30 days';
COMMENT ON FUNCTION update_updated_at_column() IS
    'Automatically updates the updated_at timestamp on record modification';
