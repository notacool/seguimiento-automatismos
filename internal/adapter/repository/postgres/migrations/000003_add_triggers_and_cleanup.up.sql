-- [REMOVED] pg_cron no disponible en la imagen actual. El job debe programarse externamente si se requiere.

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subtasks_updated_at
    BEFORE UPDATE ON subtasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

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

-- [REMOVED] El job programado con pg_cron debe implementarse externamente si se requiere.

COMMENT ON FUNCTION cleanup_soft_deleted_records() IS
    'Permanently deletes tasks and subtasks that have been soft-deleted for more than 30 days';
COMMENT ON FUNCTION update_updated_at_column() IS
    'Automatically updates the updated_at timestamp on record modification';
