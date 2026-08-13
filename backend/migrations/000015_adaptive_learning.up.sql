CREATE TABLE mistake_pattern_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attempt_id BIGINT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    topic_id BIGINT REFERENCES topics(id) ON DELETE SET NULL,
    user_role VARCHAR NOT NULL CHECK (user_role IN ('buyer','seller')),
    pattern_code VARCHAR(80) NOT NULL,
    is_safe BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(attempt_id, pattern_code, is_safe)
);
CREATE INDEX mistake_pattern_events_profile_idx ON mistake_pattern_events(user_id, user_role, pattern_code, created_at DESC);

CREATE VIEW mistake_pattern_stats AS
WITH attempts AS (
    SELECT user_id,user_role,attempt_id,MAX(created_at) completed_at
    FROM mistake_pattern_events
    GROUP BY user_id,user_role,attempt_id
), ranked_attempts AS (
    SELECT user_id,user_role,attempt_id,
           ROW_NUMBER() OVER (PARTITION BY user_id,user_role ORDER BY completed_at DESC,attempt_id DESC) recent_rank
    FROM attempts
)
SELECT e.user_id,e.user_role,e.pattern_code,
       COUNT(*) FILTER (WHERE NOT e.is_safe)::int unsafe_count,
       COUNT(*) FILTER (WHERE e.is_safe)::int safe_count,
       COUNT(DISTINCT e.attempt_id) FILTER (WHERE NOT e.is_safe AND recent.recent_rank <= 5)::int recent_unsafe
FROM mistake_pattern_events e
JOIN ranked_attempts recent ON recent.user_id=e.user_id AND recent.user_role=e.user_role AND recent.attempt_id=e.attempt_id
GROUP BY e.user_id,e.user_role,e.pattern_code;

CREATE TABLE skill_check_pairs (
    id BIGSERIAL PRIMARY KEY,
    topic_id BIGINT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    pair_number SMALLINT NOT NULL CHECK (pair_number BETWEEN 1 AND 3),
    before_messages JSONB NOT NULL,
    before_is_scam BOOLEAN NOT NULL,
    before_pattern_code VARCHAR(80) NOT NULL,
    after_messages JSONB NOT NULL,
    after_is_scam BOOLEAN NOT NULL,
    after_pattern_code VARCHAR(80) NOT NULL,
    UNIQUE(topic_id, pair_number),
    CHECK (before_messages <> after_messages)
);

CREATE TABLE skill_checks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id BIGINT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    pair_id BIGINT NOT NULL REFERENCES skill_check_pairs(id) ON DELETE RESTRICT,
    before_answer BOOLEAN,
    after_answer BOOLEAN,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    UNIQUE(user_id, topic_id, id)
);
CREATE INDEX skill_checks_user_topic_idx ON skill_checks(user_id, topic_id, started_at DESC);
CREATE UNIQUE INDEX skill_checks_one_active_idx ON skill_checks(user_id, topic_id) WHERE completed_at IS NULL;

INSERT INTO skill_check_pairs(topic_id,pair_number,before_messages,before_is_scam,before_pattern_code,after_messages,after_is_scam,after_pattern_code)
SELECT t.id, n.pair_number,
       CASE n.pair_number WHEN 1 THEN
         jsonb_build_array(jsonb_build_object('role','assistant','text',format('По теме «%s»: откройте форму доставки из сообщения, бронь действует пять минут.',t.title)),jsonb_build_object('role','user','text','Хорошо, сейчас посмотрю.'))
       ELSE
         jsonb_build_array(jsonb_build_object('role','assistant','text',format('По теме «%s»: назовите код подтверждения, иначе заказ отменится.',t.title)),jsonb_build_object('role','user','text','Код уже пришёл.'))
       END,
       true,
       CASE n.pair_number WHEN 1 THEN 'external_link' ELSE 'account_takeover' END,
       CASE n.pair_number WHEN 1 THEN
         jsonb_build_array(jsonb_build_object('role','assistant','text',format('По теме «%s»: отсканируйте QR для получения товара, предложение скоро исчезнет.',t.title)),jsonb_build_object('role','user','text','QR ведёт на страницу подтверждения.'))
       ELSE
         jsonb_build_array(jsonb_build_object('role','assistant','text',format('По теме «%s»: пришлите данные карты для возврата, оператор ждёт.',t.title)),jsonb_build_object('role','user','text','Возврат обещают сегодня.'))
       END,
       true,
       CASE n.pair_number WHEN 1 THEN 'external_link' ELSE 'account_takeover' END
FROM topics t CROSS JOIN (VALUES (1),(2)) AS n(pair_number);
