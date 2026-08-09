ALTER TABLE topics ADD COLUMN content_status VARCHAR NOT NULL DEFAULT 'published'
    CHECK (content_status IN ('draft', 'published', 'archived'));
ALTER TABLE topics ADD COLUMN archived_at TIMESTAMP WITH TIME ZONE;
UPDATE topics SET content_status = CASE WHEN published THEN 'published' ELSE 'draft' END;
ALTER TABLE topics DROP CONSTRAINT topics_user_role_sort_order_key;
ALTER TABLE topics DROP COLUMN published;

CREATE UNIQUE INDEX topics_published_role_order_idx
    ON topics(user_role, sort_order) WHERE content_status = 'published';

CREATE TABLE daily_tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_date DATE NOT NULL,
    user_role VARCHAR NOT NULL CHECK (user_role IN ('buyer', 'seller')),
    action_type VARCHAR NOT NULL CHECK (action_type IN ('resume_attempt', 'read_theory', 'take_quiz', 'start_level', 'start_free_play')),
    topic_id INTEGER REFERENCES topics(id) ON DELETE RESTRICT,
    level_number INTEGER CHECK (level_number BETWEEN 1 AND 4),
    attempt_id INTEGER REFERENCES chat_sessions(id) ON DELETE RESTRICT,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, activity_date, user_role)
);

CREATE INDEX daily_tasks_incomplete_target_idx
    ON daily_tasks(user_id, activity_date, user_role) WHERE completed_at IS NULL;
