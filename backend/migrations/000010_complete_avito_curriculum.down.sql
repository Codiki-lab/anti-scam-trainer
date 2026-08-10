UPDATE chats SET content_status='archived',archived_at=CURRENT_TIMESTAMP,is_active=FALSE
WHERE id IN (SELECT id FROM migration_000010_new_chats);

DELETE FROM chats c WHERE c.id IN (SELECT id FROM migration_000010_new_chats)
AND NOT EXISTS (SELECT 1 FROM chat_sessions a WHERE a.chat_id=c.id);

UPDATE chats c SET content_status=m.previous_content_status,archived_at=m.previous_archived_at,is_active=m.previous_is_active
FROM migration_000010_archived_chats m WHERE c.id=m.id;

UPDATE topics t SET title=s.title,description=s.description
FROM migration_000010_topic_snapshot s WHERE t.id=s.id;

UPDATE theory_blocks b SET kind=s.kind,title=s.title,body=s.body
FROM migration_000010_theory_snapshot s WHERE b.id=s.id;

UPDATE quiz_questions q SET text=s.text,explanation=s.explanation
FROM migration_000010_question_snapshot s WHERE q.id=s.id;

UPDATE quiz_options o SET text=s.text,is_correct=s.is_correct
FROM migration_000010_quiz_option_snapshot s WHERE o.id=s.id;

UPDATE free_play_configs f SET product_context=s.product_context,system_prompt=s.system_prompt,final_rubric=s.final_rubric
FROM migration_000010_free_play_snapshot s WHERE f.user_role=s.user_role;

DROP TABLE migration_000010_new_chats;
DROP TABLE migration_000010_archived_chats;
DROP TABLE migration_000010_free_play_snapshot;
DROP TABLE migration_000010_quiz_option_snapshot;
DROP TABLE migration_000010_question_snapshot;
DROP TABLE migration_000010_theory_snapshot;
DROP TABLE migration_000010_topic_snapshot;
