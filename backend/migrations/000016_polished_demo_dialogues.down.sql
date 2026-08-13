UPDATE chats
SET content_status='archived',archived_at=CURRENT_TIMESTAMP,is_active=FALSE
WHERE id IN (SELECT id FROM migration_000016_new_chats);

DELETE FROM chats c
WHERE c.id IN (SELECT id FROM migration_000016_new_chats)
  AND NOT EXISTS (SELECT 1 FROM chat_sessions a WHERE a.chat_id=c.id);

UPDATE chats c
SET content_status=m.content_status,archived_at=m.archived_at,is_active=m.is_active
FROM migration_000016_archived_chats m
WHERE c.id=m.id;

DROP TABLE migration_000016_new_chats;
DROP TABLE migration_000016_archived_chats;
