DROP TABLE daily_tasks;
DROP INDEX topics_published_role_order_idx;
ALTER TABLE topics ADD COLUMN published BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE topics SET published = TRUE;
ALTER TABLE topics ADD CONSTRAINT topics_user_role_sort_order_key UNIQUE(user_role, sort_order);
ALTER TABLE topics DROP COLUMN archived_at;
ALTER TABLE topics DROP COLUMN content_status;
