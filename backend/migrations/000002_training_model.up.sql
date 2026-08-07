CREATE TABLE levels (
    id SERIAL PRIMARY KEY,
    level_number INTEGER NOT NULL UNIQUE CHECK (level_number BETWEEN 1 AND 4),
    response_type VARCHAR NOT NULL CHECK (response_type IN ('multiple_choice', 'similar_choice', 'mixed', 'free_text')),
    ai_mode VARCHAR NOT NULL CHECK (ai_mode IN ('disabled', 'evaluate_and_reply', 'generate')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

INSERT INTO levels (level_number, response_type, ai_mode) VALUES
    (1, 'multiple_choice', 'disabled'),
    (2, 'similar_choice', 'disabled'),
    (3, 'mixed', 'evaluate_and_reply'),
    (4, 'free_text', 'generate');

CREATE TABLE user_level_progress (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    level_id INTEGER NOT NULL REFERENCES levels(id) ON DELETE RESTRICT,
    user_role VARCHAR NOT NULL CHECK (user_role IN ('buyer', 'seller')),
    best_score INTEGER NOT NULL DEFAULT 0 CHECK (best_score >= 0),
    stars INTEGER NOT NULL DEFAULT 0 CHECK (stars BETWEEN 0 AND 3),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    passed_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (user_id, user_role, level_id)
);

ALTER TABLE chats
    ADD COLUMN level_id INTEGER REFERENCES levels(id) ON DELETE RESTRICT,
    ADD COLUMN user_role VARCHAR CHECK (user_role IN ('buyer', 'seller')),
    ADD COLUMN mode VARCHAR NOT NULL DEFAULT 'scenario' CHECK (mode = 'scenario'),
    ADD COLUMN scam_scheme VARCHAR,
    ADD COLUMN product_context JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD CONSTRAINT chats_role_check CHECK (role IN ('buyer', 'seller')),
    ADD CONSTRAINT chats_user_role_matches_role CHECK (user_role IS NULL OR user_role = role);

ALTER TABLE chat_steps
    DROP COLUMN role,
    DROP COLUMN message_text,
    ADD COLUMN step_goal TEXT NOT NULL DEFAULT '',
    ADD COLUMN ai_instruction TEXT,
    ADD COLUMN fallback_message TEXT,
    ADD COLUMN max_points INTEGER NOT NULL DEFAULT 0 CHECK (max_points >= 0);

UPDATE chat_steps
SET response_type = 'multiple_choice'
WHERE response_type IS NULL;

ALTER TABLE chat_steps
    ALTER COLUMN response_type SET DEFAULT 'multiple_choice',
    ALTER COLUMN response_type SET NOT NULL,
    ADD CONSTRAINT chat_steps_response_type_check
        CHECK (response_type IN ('multiple_choice', 'similar_choice', 'mixed', 'free_text'));

DROP INDEX idx_chat_option_step;

ALTER TABLE chat_options
    DROP COLUMN chat_id,
    DROP COLUMN step_number,
    ADD COLUMN step_id INTEGER NOT NULL REFERENCES chat_steps(id) ON DELETE CASCADE,
    ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
    ADD CONSTRAINT chat_options_step_sort_order_key UNIQUE (step_id, sort_order);

ALTER TABLE chat_sessions
    ALTER COLUMN chat_id DROP NOT NULL,
    ADD COLUMN mode VARCHAR NOT NULL DEFAULT 'scenario' CHECK (mode IN ('scenario', 'free_play')),
    ADD COLUMN current_step_number INTEGER NOT NULL DEFAULT 0 CHECK (current_step_number >= 0),
    ADD COLUMN is_scam BOOLEAN,
    ADD COLUMN max_score INTEGER NOT NULL DEFAULT 0 CHECK (max_score >= 0),
    ADD CONSTRAINT chat_sessions_status_check CHECK (status IN ('IN_PROGRESS', 'COMPLETED', 'ABANDONED')),
    ADD CONSTRAINT chat_sessions_mode_chat_check CHECK (
        (mode = 'scenario' AND chat_id IS NOT NULL AND is_scam IS NULL)
        OR (mode = 'free_play' AND chat_id IS NULL AND is_scam IS NOT NULL)
    );

CREATE TABLE session_answers (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    step_id INTEGER NOT NULL REFERENCES chat_steps(id) ON DELETE RESTRICT,
    option_id INTEGER REFERENCES chat_options(id) ON DELETE RESTRICT,
    free_text TEXT,
    is_correct BOOLEAN NOT NULL,
    awarded_points INTEGER NOT NULL DEFAULT 0,
    ai_evaluation TEXT,
    explanation TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT session_answers_one_response_check CHECK (num_nonnulls(option_id, free_text) = 1),
    UNIQUE (session_id, step_id)
);

DROP TABLE leaderboard;
DROP TABLE statistics;
