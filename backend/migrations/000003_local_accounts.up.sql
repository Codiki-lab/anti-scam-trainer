ALTER TABLE users
    DROP COLUMN user_id,
    DROP COLUMN completed_chats,
    ADD COLUMN password_hash VARCHAR NOT NULL DEFAULT '',
    ADD COLUMN access_role VARCHAR NOT NULL DEFAULT 'user' CHECK (access_role IN ('user', 'admin'));

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users GROUP BY LOWER(username) HAVING COUNT(*) > 1) THEN
        RAISE EXCEPTION 'cannot migrate users: usernames collide after lowercase normalization';
    END IF;
END $$;

UPDATE users SET username = LOWER(username);

ALTER TABLE users
    ADD CONSTRAINT users_username_key UNIQUE (username);

CREATE UNIQUE INDEX users_one_admin_idx ON users ((access_role)) WHERE access_role = 'admin';
