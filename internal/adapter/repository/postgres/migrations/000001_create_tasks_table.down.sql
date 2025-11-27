-- Drop tasks table and all related objects
DROP INDEX IF EXISTS idx_tasks_cleanup;
DROP INDEX IF EXISTS idx_tasks_deleted_at;
DROP INDEX IF EXISTS idx_tasks_name;
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_state;
DROP TABLE IF EXISTS tasks CASCADE;
