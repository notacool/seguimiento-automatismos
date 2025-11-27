-- Drop subtasks table and all related objects
DROP INDEX IF EXISTS idx_subtasks_cleanup;
DROP INDEX IF EXISTS idx_subtasks_task_state;
DROP INDEX IF EXISTS idx_subtasks_deleted_at;
DROP INDEX IF EXISTS idx_subtasks_created_at;
DROP INDEX IF EXISTS idx_subtasks_state;
DROP INDEX IF EXISTS idx_subtasks_task_id;
DROP TABLE IF EXISTS subtasks CASCADE;
