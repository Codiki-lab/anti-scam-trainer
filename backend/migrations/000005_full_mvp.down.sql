DELETE FROM chat_sessions WHERE mode = 'free_play' OR chat_id IN (
    SELECT chats.id FROM chats JOIN levels ON levels.id = chats.level_id WHERE levels.level_number IN (3, 4)
);
DROP TABLE free_play_configs;
DELETE FROM chats WHERE id IN (
    SELECT chats.id FROM chats JOIN levels ON levels.id = chats.level_id WHERE levels.level_number IN (3, 4)
);

DROP INDEX chat_sessions_one_in_progress_free_play_idx;
DROP INDEX session_answers_session_turn_idx;
ALTER TABLE session_answers DROP COLUMN turn_number;
ALTER TABLE session_answers ALTER COLUMN step_id SET NOT NULL;
ALTER TABLE session_answers ADD CONSTRAINT session_answers_session_id_step_id_key UNIQUE (session_id, step_id);

ALTER TABLE chat_sessions DROP CONSTRAINT chat_sessions_mode_chat_check;
ALTER TABLE chat_sessions DROP COLUMN final_breakdown, DROP COLUMN free_text_count, DROP COLUMN user_role;
ALTER TABLE chat_sessions ADD CONSTRAINT chat_sessions_mode_chat_check CHECK (
    (mode = 'scenario' AND chat_id IS NOT NULL AND is_scam IS NULL)
    OR (mode = 'free_play' AND chat_id IS NULL AND is_scam IS NOT NULL)
);

ALTER TABLE chats DROP COLUMN final_rubric, DROP COLUMN ai_system_prompt;
