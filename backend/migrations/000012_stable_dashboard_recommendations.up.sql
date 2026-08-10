CREATE TABLE dashboard_recommendations (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_date DATE NOT NULL,
    user_role VARCHAR NOT NULL CHECK (user_role IN ('buyer','seller')),
    action JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, activity_date, user_role)
);
