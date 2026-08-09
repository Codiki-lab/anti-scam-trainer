ALTER TABLE chats
    ADD COLUMN ai_system_prompt TEXT,
    ADD COLUMN final_rubric JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE chat_sessions DROP CONSTRAINT chat_sessions_mode_chat_check;
ALTER TABLE chat_sessions
    ADD COLUMN user_role VARCHAR CHECK (user_role IN ('buyer', 'seller')),
    ADD COLUMN free_text_count INTEGER NOT NULL DEFAULT 0 CHECK (free_text_count BETWEEN 0 AND 5),
    ADD COLUMN final_breakdown JSONB,
    ADD CONSTRAINT chat_sessions_mode_chat_check CHECK (
        (mode = 'scenario' AND chat_id IS NOT NULL AND is_scam IS NULL)
        OR (mode = 'free_play' AND chat_id IS NULL AND is_scam IS NOT NULL AND user_role IS NOT NULL)
    );

UPDATE chat_sessions AS sessions
SET user_role = chats.user_role
FROM chats
WHERE sessions.chat_id = chats.id;

ALTER TABLE session_answers DROP CONSTRAINT session_answers_session_id_step_id_key;
ALTER TABLE session_answers ALTER COLUMN step_id DROP NOT NULL;
ALTER TABLE session_answers ADD COLUMN turn_number INTEGER NOT NULL DEFAULT 0 CHECK (turn_number >= 0);
WITH numbered AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY created_at, id) AS turn_number
    FROM session_answers
)
UPDATE session_answers AS answers
SET turn_number = numbered.turn_number
FROM numbered
WHERE answers.id = numbered.id;
CREATE UNIQUE INDEX session_answers_session_turn_idx ON session_answers (session_id, turn_number);

CREATE UNIQUE INDEX chat_sessions_one_in_progress_free_play_idx
    ON chat_sessions (user_id, user_role)
    WHERE status = 'IN_PROGRESS' AND mode = 'free_play';

CREATE TABLE free_play_configs (
    user_role VARCHAR PRIMARY KEY CHECK (user_role IN ('buyer', 'seller')),
    product_context JSONB NOT NULL,
    system_prompt TEXT NOT NULL,
    final_rubric JSONB NOT NULL
);

INSERT INTO free_play_configs (user_role, product_context, system_prompt, final_rubric) VALUES
    ('buyer', '{"item":"товар из объявления","channel":"чат сервиса"}', 'Веди правдоподобный разговор с покупателем и не раскрывай учебную механику.', '{"checks_risk":100,"blindly_refuses":50,"leaves_service":0}'),
    ('seller', '{"item":"товар пользователя","channel":"чат сервиса"}', 'Веди правдоподобный разговор с продавцом и не раскрывай учебную механику.', '{"protects_secrets":100,"blindly_refuses":50,"shares_code":0}');

UPDATE chats
SET title = CASE
        WHEN user_role = 'buyer' AND level_id = (SELECT id FROM levels WHERE level_number = 1) THEN 'Покупатель: безопасная доставка'
        WHEN user_role = 'buyer' THEN 'Покупатель: проверка оплаты'
        WHEN user_role = 'seller' AND level_id = (SELECT id FROM levels WHERE level_number = 1) THEN 'Продавец: защита данных карты'
        ELSE 'Продавец: подмена службы поддержки'
    END,
    description = CASE
        WHEN user_role = 'buyer' AND level_id = (SELECT id FROM levels WHERE level_number = 1) THEN 'Продавец уводит оплату доставки на внешний сайт.'
        WHEN user_role = 'buyer' THEN 'Собеседник предлагает похожие способы оплаты с разным уровнем риска.'
        WHEN user_role = 'seller' AND level_id = (SELECT id FROM levels WHERE level_number = 1) THEN 'Покупатель просит данные карты для якобы получения перевода.'
        ELSE 'Лже-поддержка просит подтвердить сделку вне сервиса.'
    END,
    scam_scheme = CASE WHEN user_role = 'buyer' THEN 'фишинговая оплата' ELSE 'кража платёжных данных' END,
    product_context = jsonb_build_object('category', CASE WHEN user_role = 'buyer' THEN 'электроника' ELSE 'товары для дома' END),
    ai_system_prompt = 'Ты виртуальный собеседник учебного сценария. Верни только строгий JSON: awarded_points, explanation, reply, risk_signals.',
    final_rubric = '{"safe_channel":100,"external_link":0}'::jsonb
WHERE content_status = 'published';

UPDATE chat_steps AS steps
SET response_type = CASE WHEN levels.level_number = 1 THEN 'multiple_choice' ELSE 'similar_choice' END,
    step_goal = CASE steps.step_number
        WHEN 1 THEN 'Распознать попытку увести общение из сервиса'
        WHEN 2 THEN 'Не раскрыть платёжные данные'
        ELSE 'Выбрать штатное завершение сделки'
    END,
    fallback_message = CASE steps.step_number
        WHEN 1 THEN 'Так будет быстрее — откройте ссылку из сообщения.'
        WHEN 2 THEN 'Для перевода напишите данные карты и код подтверждения.'
        ELSE 'Подтвердите сделку сейчас, иначе объявление заблокируют.'
    END
FROM chats JOIN levels ON levels.id = chats.level_id
WHERE steps.chat_id = chats.id AND chats.content_status = 'published';

UPDATE chat_options AS options
SET option_text = CASE
        WHEN levels.level_number = 2 AND options.sort_order = 1 AND steps.step_number = 1 THEN 'Проверить адрес ссылки вручную и открыть её для оформления'
        WHEN levels.level_number = 2 AND options.sort_order = 2 AND steps.step_number = 1 THEN 'Попросить заново оформить доставку внутри приложения'
        WHEN levels.level_number = 2 AND options.sort_order = 1 AND steps.step_number = 2 THEN 'Передать только номер карты, но не сообщать код подтверждения'
        WHEN levels.level_number = 2 AND options.sort_order = 2 AND steps.step_number = 2 THEN 'Использовать штатное получение оплаты без передачи реквизитов собеседнику'
        WHEN levels.level_number = 2 AND options.sort_order = 1 THEN 'Самостоятельно найти поддержку по названию из сообщения'
        WHEN levels.level_number = 2 THEN 'Открыть раздел поддержки непосредственно в приложении'
        WHEN options.sort_order = 1 AND steps.step_number = 1 THEN 'Открыть присланную ссылку и проверить страницу'
        WHEN options.sort_order = 2 AND steps.step_number = 1 THEN 'Остаться в чате и использовать только штатную оплату сервиса'
        WHEN options.sort_order = 1 AND steps.step_number = 2 THEN 'Сообщить реквизиты и код, чтобы не задерживать сделку'
        WHEN options.sort_order = 2 AND steps.step_number = 2 THEN 'Не передавать коды и данные карты собеседнику'
        WHEN options.sort_order = 1 THEN 'Поспешить из-за угрозы блокировки'
        ELSE 'Остановить сделку и обратиться в поддержку через приложение'
    END,
    explanation = CASE WHEN levels.level_number = 2 AND options.sort_order = 1
        THEN 'Ответ выглядит осторожным, но всё ещё оставляет канал для фишинга или передачи реквизитов.'
        WHEN options.sort_order = 1
        THEN 'Давление, внешние ссылки и запрос секретных данных — признаки риска.'
        ELSE 'Штатный канал и отказ передавать секреты уменьшают риск.'
    END,
    points = CASE WHEN levels.level_number = 2 AND options.sort_order = 1 THEN 25 WHEN options.sort_order = 1 THEN 0 ELSE 100 END
FROM chat_steps AS steps
JOIN chats ON chats.id = steps.chat_id
JOIN levels ON levels.id = chats.level_id
WHERE options.step_id = steps.id AND chats.content_status = 'published';

INSERT INTO chats (title, description, difficulty, role, is_active, level_id, user_role, content_status, scam_scheme, product_context, ai_system_prompt, final_rubric)
SELECT data.title, data.description, levels.level_number::text, data.user_role, TRUE, levels.id, data.user_role, 'published', data.scam_scheme, data.product_context::jsonb,
       CASE WHEN levels.level_number = 4
            THEN 'Ты играешь виртуального собеседника-мошенника в учебном диалоге. Не раскрывай роль напрямую. Оцени безопасное поведение по рубрике и верни только JSON с awarded_points (0,25,50,75,100), explanation, reply и risk_signals.'
            ELSE 'Ты виртуальный собеседник тренажёра. Оцени безопасное поведение по рубрике и верни только JSON с awarded_points (0,25,50,75,100), explanation, reply и risk_signals.'
       END,
       data.final_rubric::jsonb
FROM (VALUES
    ('Покупатель: смешанный ответ', 'Продавец предлагает оформить доставку через внешний сайт.', 'buyer', 3, 'фишинговая доставка', '{"item":"смартфон","price":42000}', '{"refuse_link":100,"share_code":0}'),
    ('Покупатель: свободный диалог', 'Мошенник торопит с предоплатой редкого товара.', 'buyer', 4, 'предоплата вне сервиса', '{"item":"игровая приставка","price":65000}', '{"stay_in_service":100,"prepay":0}'),
    ('Продавец: смешанный ответ', 'Лже-покупатель просит данные карты для перевода.', 'seller', 3, 'кража платёжных данных', '{"item":"кофемашина","price":18000}', '{"refuse_card_data":100,"share_code":0}'),
    ('Продавец: свободный диалог', 'Мошенник изображает поддержку и требует подтверждение.', 'seller', 4, 'лже-поддержка', '{"item":"велосипед","price":35000}', '{"official_support":100,"external_link":0}')
) AS data(title, description, user_role, level_number, scam_scheme, product_context, final_rubric)
JOIN levels ON levels.level_number = data.level_number;

INSERT INTO chat_steps (chat_id, step_number, response_type, step_goal, ai_instruction, fallback_message, max_points)
SELECT chats.id, 1, levels.response_type,
       CASE WHEN levels.level_number = 3 THEN 'Сформулировать безопасный ответ своими словами или выбрать вариант' ELSE 'Вести свободный безопасный диалог' END,
       'Начисли 100 за отказ от внешних ссылок, кодов и предоплаты; 75 за безопасный, но неполный ответ; 0 за передачу секретов или оплату вне сервиса.',
       CASE WHEN chats.user_role = 'buyer' THEN 'Я уже оформил доставку — откройте ссылку и оплатите.' ELSE 'Чтобы получить деньги, подтвердите карту по ссылке.' END,
       100
FROM chats JOIN levels ON levels.id = chats.level_id
WHERE chats.content_status = 'published' AND levels.level_number IN (3, 4);

INSERT INTO chat_options (step_id, option_text, explanation, points, sort_order)
SELECT steps.id, options.option_text, options.explanation, options.points, options.sort_order
FROM chat_steps AS steps
JOIN chats ON chats.id = steps.chat_id
JOIN levels ON levels.id = chats.level_id AND levels.level_number = 3
CROSS JOIN LATERAL (VALUES
    ('Перейти по ссылке, если адрес выглядит знакомо', 'Внешняя страница может копировать интерфейс сервиса.', 0, 1),
    ('Отказаться от ссылки и продолжить только штатным способом', 'Безопасный ответ сохраняет сделку внутри сервиса.', 100, 2)
) AS options(option_text, explanation, points, sort_order);
