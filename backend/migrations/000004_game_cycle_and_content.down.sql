DELETE FROM chat_options WHERE step_id IN (SELECT id FROM chat_steps WHERE chat_id IN (SELECT id FROM chats WHERE content_status = 'published'));
DELETE FROM chat_steps WHERE chat_id IN (SELECT id FROM chats WHERE content_status = 'published');
DELETE FROM chats WHERE content_status = 'published';
DROP INDEX chat_sessions_one_in_progress_user_chat_idx;
DROP INDEX chats_one_published_role_level_idx;
ALTER TABLE chat_options DROP CONSTRAINT chat_options_points_check;
ALTER TABLE session_answers ADD COLUMN is_correct BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE chat_options ADD COLUMN is_correct BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE chats DROP COLUMN archived_at, DROP COLUMN content_status;
