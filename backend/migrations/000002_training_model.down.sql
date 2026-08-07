CREATE TABLE statistics (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_chats INTEGER DEFAULT 0,
    completed_chats INTEGER DEFAULT 0,
    failed_chats INTEGER DEFAULT 0,
    total_messages INTEGER DEFAULT 0,
    time_spent INTEGER DEFAULT 0,
    success_rate NUMERIC(5, 2) DEFAULT 0.00
);

CREATE TABLE leaderboard (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    rank INTEGER DEFAULT 0,
    score INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE session_answers;

DELETE FROM chat_sessions
WHERE mode = 'free_play';

ALTER TABLE chat_sessions
    DROP CONSTRAINT chat_sessions_mode_chat_check,
    DROP CONSTRAINT chat_sessions_status_check,
    DROP COLUMN max_score,
    DROP COLUMN is_scam,
    DROP COLUMN current_step_number,
    DROP COLUMN mode,
    ALTER COLUMN chat_id SET NOT NULL;

ALTER TABLE chat_options
    DROP CONSTRAINT chat_options_step_sort_order_key,
    ADD COLUMN chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    ADD COLUMN step_number INTEGER;

UPDATE chat_options AS option
SET chat_id = step.chat_id,
    step_number = step.step_number
FROM chat_steps AS step
WHERE step.id = option.step_id;

ALTER TABLE chat_options
    ALTER COLUMN chat_id SET NOT NULL,
    ALTER COLUMN step_number SET NOT NULL,
    DROP COLUMN sort_order,
    DROP COLUMN step_id;

CREATE INDEX idx_chat_option_step ON chat_options (chat_id, step_number);

ALTER TABLE chat_steps
    DROP COLUMN max_points,
    DROP COLUMN fallback_message,
    DROP COLUMN ai_instruction,
    DROP COLUMN step_goal,
    DROP CONSTRAINT chat_steps_response_type_check,
    ALTER COLUMN response_type DROP NOT NULL,
    ALTER COLUMN response_type DROP DEFAULT,
    ADD COLUMN role VARCHAR NOT NULL DEFAULT 'NPC',
    ADD COLUMN message_text TEXT NOT NULL DEFAULT '';

ALTER TABLE chats
    DROP CONSTRAINT chats_user_role_matches_role,
    DROP CONSTRAINT chats_role_check,
    DROP COLUMN product_context,
    DROP COLUMN scam_scheme,
    DROP COLUMN mode,
    DROP COLUMN user_role,
    DROP COLUMN level_id;

DROP TABLE user_level_progress;
DROP TABLE levels;
