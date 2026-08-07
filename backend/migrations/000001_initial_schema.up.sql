CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR UNIQUE NOT NULL,
    username VARCHAR NOT NULL,
    completed_chats INTEGER DEFAULT 0
);

CREATE TABLE chats (
    id SERIAL PRIMARY KEY,
    title VARCHAR NOT NULL,
    description TEXT,
    difficulty VARCHAR NOT NULL,
    role VARCHAR NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE achievements (
    id SERIAL PRIMARY KEY,
    title VARCHAR NOT NULL,
    description TEXT,
    icon VARCHAR,
    condition_type VARCHAR NOT NULL,
    condition_value VARCHAR NOT NULL
);

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

CREATE TABLE user_achievements (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id INTEGER NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, achievement_id)
);

CREATE TABLE chat_steps (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    role VARCHAR NOT NULL,
    message_text TEXT NOT NULL,
    response_type VARCHAR
);

CREATE UNIQUE INDEX idx_chat_step_number ON chat_steps (chat_id, step_number);

CREATE TABLE chat_options (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    option_text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    explanation TEXT,
    points INTEGER DEFAULT 0
);

CREATE INDEX idx_chat_option_step ON chat_options (chat_id, step_number);

CREATE TABLE chat_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE RESTRICT,
    status VARCHAR NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP WITH TIME ZONE,
    score INTEGER DEFAULT 0
);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
