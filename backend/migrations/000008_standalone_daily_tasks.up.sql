ALTER TABLE daily_tasks RENAME TO daily_tasks_technical_000007;

CREATE TABLE daily_tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_date DATE NOT NULL,
    user_role VARCHAR NOT NULL CHECK (user_role IN ('buyer', 'seller')),
    messages JSONB NOT NULL,
    verdict BOOLEAN NOT NULL,
    signals JSONB NOT NULL,
    safe_action TEXT NOT NULL,
    user_answer BOOLEAN,
    is_correct BOOLEAN,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK ((user_answer IS NULL) = (is_correct IS NULL)),
    CHECK ((user_answer IS NULL) = (completed_at IS NULL)),
    UNIQUE(user_id, activity_date)
);
