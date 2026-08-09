DROP TABLE IF EXISTS attempt_results;
DELETE FROM user_achievements WHERE achievement_id IN (SELECT id FROM achievements WHERE code IN ('first_training','five_trainings','perfect_score','first_topic_completed','all_buyer_topics','all_seller_topics','streak_3','streak_7'));
DELETE FROM achievements WHERE code IN ('first_training','five_trainings','perfect_score','first_topic_completed','all_buyer_topics','all_seller_topics','streak_3','streak_7');
ALTER TABLE achievements DROP CONSTRAINT IF EXISTS achievements_code_key;
ALTER TABLE achievements DROP COLUMN IF EXISTS code;

DELETE FROM chats WHERE product_context->>'seed_version' = 'issue-49';
ALTER TABLE chat_steps DROP COLUMN IF EXISTS counterparty_message;
DROP INDEX IF EXISTS chats_one_published_topic_level_idx;
CREATE UNIQUE INDEX chats_one_published_role_level_idx ON chats (user_role, level_id) WHERE content_status = 'published' AND archived_at IS NULL;
ALTER TABLE chats DROP COLUMN IF EXISTS topic_id;

DELETE FROM user_level_progress newer
USING user_level_progress older
WHERE newer.id > older.id AND newer.user_id = older.user_id AND newer.user_role = older.user_role AND newer.level_id = older.level_id;
ALTER TABLE user_level_progress DROP CONSTRAINT IF EXISTS user_level_progress_user_topic_level_key;
ALTER TABLE user_level_progress DROP COLUMN IF EXISTS topic_id;
ALTER TABLE user_level_progress ADD CONSTRAINT user_level_progress_user_id_user_role_level_id_key UNIQUE (user_id, user_role, level_id);

DROP TABLE IF EXISTS daily_activity;
DROP TABLE IF EXISTS quiz_attempts;
DROP TABLE IF EXISTS user_topic_progress;
DROP TABLE IF EXISTS quiz_options;
DROP TABLE IF EXISTS quiz_questions;
DROP TABLE IF EXISTS theory_blocks;
DROP TABLE IF EXISTS topics;

ALTER TABLE users
    DROP COLUMN IF EXISTS last_activity_date,
    DROP COLUMN IF EXISTS longest_streak,
    DROP COLUMN IF EXISTS current_streak,
    DROP COLUMN IF EXISTS training_role;
