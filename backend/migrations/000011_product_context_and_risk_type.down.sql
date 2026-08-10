ALTER TABLE chats DROP CONSTRAINT chats_risk_type_check;
ALTER TABLE chats DROP COLUMN risk_type;

UPDATE free_play_configs AS config
SET product_context = saved.product_context
FROM migration_000011_free_play_contexts AS saved
WHERE config.user_role = saved.user_role;
DROP TABLE migration_000011_free_play_contexts;

ALTER TABLE chat_sessions DROP CONSTRAINT chat_sessions_free_text_count_check;
ALTER TABLE chat_sessions ADD CONSTRAINT chat_sessions_free_text_count_check CHECK (free_text_count BETWEEN 0 AND 6);

INSERT INTO chat_steps(id,chat_id,step_number,response_type,step_goal,counterparty_message,max_points,ai_instruction,fallback_message)
SELECT id,chat_id,step_number,response_type,step_goal,counterparty_message,max_points,ai_instruction,fallback_message
FROM migration_000011_removed_steps;
DROP TABLE migration_000011_removed_steps;
