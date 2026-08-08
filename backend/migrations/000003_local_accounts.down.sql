DROP INDEX users_one_admin_idx;

ALTER TABLE users
    DROP CONSTRAINT users_username_key,
    DROP COLUMN password_hash,
    DROP COLUMN access_role,
    ADD COLUMN user_id VARCHAR,
    ADD COLUMN completed_chats INTEGER DEFAULT 0;

UPDATE users SET user_id = id::VARCHAR;

ALTER TABLE users
    ALTER COLUMN user_id SET NOT NULL,
    ADD CONSTRAINT users_user_id_key UNIQUE (user_id);
