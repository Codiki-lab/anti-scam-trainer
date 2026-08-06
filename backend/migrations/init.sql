-- TABLE: users
-- Stores registered users.
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR UNIQUE NOT NULL, -- External platform identifier (e.g. Telegram ID)
    username VARCHAR NOT NULL,
    completed_chats INTEGER DEFAULT 0
);

-- TABLE: chats
-- Stores available chat scenarios.
CREATE TABLE chats (
    id SERIAL PRIMARY KEY,
    title VARCHAR NOT NULL,
    description TEXT,
    difficulty VARCHAR NOT NULL, -- e.g., 'Easy', 'Medium', 'Hard'
    role VARCHAR NOT NULL, -- The role the user plays in the scenario
    is_active BOOLEAN DEFAULT TRUE
);

-- TABLE: achievements
-- Stores available achievements.
CREATE TABLE achievements (
    id SERIAL PRIMARY KEY,
    title VARCHAR NOT NULL,
    description TEXT,
    icon VARCHAR, -- Path or identifier for the icon
    condition_type VARCHAR NOT NULL, -- e.g., 'COMPLETED_CHAT', 'TOTAL_SCORE'
    condition_value VARCHAR NOT NULL -- The required value for the condition
);

-- TABLE: statistics
-- Stores aggregated user statistics (1:1 relationship with users).
CREATE TABLE statistics (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_chats INTEGER DEFAULT 0,
    completed_chats INTEGER DEFAULT 0,
    failed_chats INTEGER DEFAULT 0,
    total_messages INTEGER DEFAULT 0,
    time_spent INTEGER DEFAULT 0, -- Stored in seconds
    success_rate NUMERIC(5, 2) DEFAULT 0.00 -- Percentage rate
);

-- TABLE: leaderboard
-- Stores user ranking information (1:1 relationship with users).
CREATE TABLE leaderboard (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    rank INTEGER DEFAULT 0,
    score INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- TABLE: user_achievements
-- Junction table (N:M relationship between users and achievements).
CREATE TABLE user_achievements (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id INTEGER NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Ensures a user doesn't receive the same achievement twice
    UNIQUE (user_id, achievement_id)
);
-- TABLE: chat_steps
-- Stores ordered messages that make up a chat scenario (Conversation flow).
CREATE TABLE chat_steps (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    role VARCHAR NOT NULL, -- Role speaking (e.g., 'User', 'NPC')
    message_text TEXT NOT NULL,
    response_type VARCHAR -- e.g., 'OPTION', 'FREE_TEXT', 'END'
);
-- Constraint to ensure step_number is unique per chat
CREATE UNIQUE INDEX idx_chat_step_number ON chat_steps (chat_id, step_number);


-- TABLE: chat_options
-- Stores answer options and scoring for specific steps.
CREATE TABLE chat_options (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    option_text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    explanation TEXT,
    points INTEGER DEFAULT 0
);
-- Constraint to ensure options are unique per step
CREATE UNIQUE INDEX idx_chat_option_step ON chat_options (chat_id, step_number);

-- TABLE: chat_sessions
-- Represents a user's attempt to complete a chat.
CREATE TABLE chat_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE RESTRICT,
    status VARCHAR NOT NULL, -- e.g., 'IN_PROGRESS', 'COMPLETED', 'ABANDONED'
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP WITH TIME ZONE,
    score INTEGER DEFAULT 0
);

-- TABLE: messages
-- Stores all messages exchanged during a chat session (Conversation history).
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR NOT NULL, -- 'USER' or 'NPC'
    message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

