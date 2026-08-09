ALTER TABLE chats
    ADD COLUMN content_status VARCHAR NOT NULL DEFAULT 'draft' CHECK (content_status IN ('draft', 'published', 'archived')),
    ADD COLUMN archived_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE chat_options DROP COLUMN is_correct;
ALTER TABLE session_answers DROP COLUMN is_correct;

ALTER TABLE chat_options
    ADD CONSTRAINT chat_options_points_check CHECK (points IN (0, 25, 50, 75, 100));

CREATE UNIQUE INDEX chats_one_published_role_level_idx
    ON chats (user_role, level_id)
    WHERE content_status = 'published' AND archived_at IS NULL;

CREATE UNIQUE INDEX chat_sessions_one_in_progress_user_chat_idx
    ON chat_sessions (user_id, chat_id)
    WHERE status = 'IN_PROGRESS' AND mode = 'scenario';

INSERT INTO chats (title, description, difficulty, role, is_active, level_id, user_role, content_status)
SELECT 'Покупка: безопасная сделка', 'Базовая тренировка покупателя', '1', 'buyer', TRUE, id, 'buyer', 'published' FROM levels WHERE level_number = 1
UNION ALL SELECT 'Покупка: проверка условий', 'Продолжение тренировки покупателя', '2', 'buyer', TRUE, id, 'buyer', 'published' FROM levels WHERE level_number = 2
UNION ALL SELECT 'Продажа: безопасная сделка', 'Базовая тренировка продавца', '1', 'seller', TRUE, id, 'seller', 'published' FROM levels WHERE level_number = 1
UNION ALL SELECT 'Продажа: проверка условий', 'Продолжение тренировки продавца', '2', 'seller', TRUE, id, 'seller', 'published' FROM levels WHERE level_number = 2;

INSERT INTO chat_steps (chat_id, step_number, response_type, step_goal, max_points)
SELECT chats.id, step_number, 'multiple_choice', 'Выберите безопасное действие', 100
FROM chats CROSS JOIN generate_series(1, 3) AS step_number
WHERE chats.content_status = 'published';

INSERT INTO chat_options (step_id, option_text, explanation, points, sort_order)
SELECT steps.id, option_text, explanation, points, sort_order
FROM chat_steps AS steps
CROSS JOIN LATERAL (VALUES
    ('Перейти по внешней ссылке', 'Не переходите по внешним ссылкам из переписки.', 0, 1),
    ('Продолжить общение только внутри сервиса', 'Безопаснее сохранять общение и оплату внутри сервиса.', 100, 2)
) AS options(option_text, explanation, points, sort_order)
JOIN chats ON chats.id = steps.chat_id
WHERE chats.content_status = 'published';
