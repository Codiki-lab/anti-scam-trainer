DELETE FROM chats WHERE id IN (SELECT id FROM migration_000009_new_chats);

UPDATE chats c SET
    content_status=m.previous_content_status,
    archived_at=m.previous_archived_at,
    is_active=m.previous_is_active
FROM migration_000009_archived_chats m
WHERE c.id=m.id;

DROP TABLE migration_000009_new_chats;
DROP TABLE migration_000009_archived_chats;

ALTER TABLE chat_sessions
    DROP COLUMN compact_summary,
    DROP COLUMN dialogue_phase,
    DROP CONSTRAINT chat_sessions_free_text_count_check;
ALTER TABLE chat_sessions
    ADD CONSTRAINT chat_sessions_free_text_count_check CHECK (free_text_count BETWEEN 0 AND 5);

ALTER TABLE chat_options
    DROP CONSTRAINT chat_options_counterparty_reaction_length,
    DROP COLUMN counterparty_reaction;
