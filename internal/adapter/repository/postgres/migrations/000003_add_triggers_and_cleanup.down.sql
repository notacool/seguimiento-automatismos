-- Remove pg_cron job
SELECT cron.unschedule('cleanup-soft-deletes');

-- Drop triggers
DROP TRIGGER IF EXISTS update_subtasks_updated_at ON subtasks;
DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;

-- Drop functions
DROP FUNCTION IF EXISTS cleanup_soft_deleted_records();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Note: We don't drop pg_cron extension as it might be used by other databases
-- DROP EXTENSION IF EXISTS pg_cron;
