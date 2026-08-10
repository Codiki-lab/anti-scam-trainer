ALTER TABLE chats ADD COLUMN risk_type VARCHAR;

CREATE TABLE migration_000011_free_play_contexts AS
SELECT user_role, product_context FROM free_play_configs;
ALTER TABLE migration_000011_free_play_contexts ADD PRIMARY KEY(user_role);

CREATE TABLE migration_000011_removed_steps AS
SELECT s.* FROM chat_steps s
JOIN chats c ON c.id=s.chat_id
JOIN levels l ON l.id=c.level_id
WHERE c.content_status='published' AND c.archived_at IS NULL AND l.level_number=4 AND s.step_number=6;
ALTER TABLE migration_000011_removed_steps ADD PRIMARY KEY(id);
DELETE FROM chat_steps WHERE id IN (SELECT id FROM migration_000011_removed_steps);

ALTER TABLE chat_sessions DROP CONSTRAINT chat_sessions_free_text_count_check;
ALTER TABLE chat_sessions ADD CONSTRAINT chat_sessions_free_text_count_check CHECK (free_text_count BETWEEN 0 AND 5);

UPDATE chats SET risk_type = CASE
    WHEN product_context->>'risk_type' IN ('phishing','prepayment','fake_payment','delivery','external_messenger','account_takeover','sms_code','social_engineering')
        THEN product_context->>'risk_type'
    WHEN scam_scheme IN ('phishing','prepayment','fake_payment','delivery','external_messenger','account_takeover','sms_code','social_engineering')
        THEN scam_scheme
    ELSE 'social_engineering'
END;

UPDATE chats SET product_context = product_context || jsonb_build_object(
    'item_title', COALESCE(NULLIF(product_context->>'item_title',''), NULLIF(product_context->>'item',''), title),
    'category', COALESCE(NULLIF(product_context->>'category',''), CASE
        WHEN product_context->>'topic' LIKE '%delivery%' THEN 'Товары с доставкой'
        WHEN product_context->>'topic' LIKE '%codes%' THEN 'Электроника'
        ELSE 'Объявления Avito'
    END),
    'deal_method', COALESCE(NULLIF(product_context->>'deal_method',''), 'delivery'),
    'currency', CASE WHEN COALESCE((product_context->>'price')::integer,0)>0 THEN 'RUB' ELSE '' END,
    'location', COALESCE(NULLIF(product_context->>'location',''), 'Москва'),
    'image_key', CASE
        WHEN lower(COALESCE(product_context->>'item','')) ~ '(iphone|смартфон)' THEN 'smartphone'
        WHEN lower(COALESCE(product_context->>'item','')) ~ '(ноутбук|macbook)' THEN 'laptop'
        WHEN lower(COALESCE(product_context->>'item','')) ~ '(playstation|приставк)' THEN 'console'
        WHEN lower(COALESCE(product_context->>'item','')) ~ '(airpods|наушник)' THEN 'headphones'
        WHEN lower(COALESCE(product_context->>'item','')) ~ '(фотоаппарат|камера)' THEN 'camera'
        WHEN lower(COALESCE(product_context->>'item','')) ~ '(велосипед)' THEN 'bicycle'
        WHEN lower(COALESCE(product_context->>'item','')) ~ '(кофемашин|пылесос)' THEN 'appliance'
        ELSE 'electronics'
    END
);

UPDATE free_play_configs SET product_context = CASE user_role
    WHEN 'buyer' THEN jsonb_build_object('item_title','Nintendo Switch OLED','category','Игровые приставки','deal_method','delivery','price',28000,'currency','RUB','location','Москва','image_key','console')
    ELSE jsonb_build_object('item_title','Смартфон Samsung Galaxy S24','category','Смартфоны','deal_method','delivery','price',59000,'currency','RUB','location','Москва','image_key','smartphone')
END;

ALTER TABLE chats ALTER COLUMN risk_type SET NOT NULL;
ALTER TABLE chats ADD CONSTRAINT chats_risk_type_check CHECK (risk_type IN (
    'phishing','prepayment','fake_payment','delivery','external_messenger','account_takeover','sms_code','social_engineering'
));
