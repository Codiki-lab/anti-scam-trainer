#!/usr/bin/env bash
set -euo pipefail

project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_root"

set -a
source backend/.env
set +a

test_database="antiscam_migration_000010_acceptance"
local_host="${POSTGRES_HOST:-127.0.0.1}"
if [[ "$local_host" == "anti-scam-trainer-postgres" ]]; then
  local_host="127.0.0.1"
fi
local_port="${POSTGRES_PORT:-5432}"
database_url="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@anti-scam-trainer-postgres:5432/${test_database}?sslmode=disable"
compose=(docker compose --env-file backend/.env -f deploy/docker-compose.yml)
psql_base=(psql -h "$local_host" -p "$local_port" -U "$POSTGRES_USER" -v ON_ERROR_STOP=1)

export PGPASSWORD="$POSTGRES_PASSWORD"
cleanup() {
  "${psql_base[@]}" -d postgres -c "DROP DATABASE IF EXISTS ${test_database}" >/dev/null
}
trap cleanup EXIT

cleanup
"${psql_base[@]}" -d postgres -c "CREATE DATABASE ${test_database}" >/dev/null

PROJECT_ROOT="$project_root" POSTGRES_DB="$test_database" "${compose[@]}" run --rm anti-scam-trainer-migrate \
  -path=/migrations -database="$database_url" up 9 >/dev/null

"${psql_base[@]}" -d "$test_database" <<'SQL'
UPDATE theory_blocks SET kind='risk' WHERE id=(SELECT b.id FROM theory_blocks b JOIN topics t ON t.id=b.topic_id WHERE t.slug='buyer-phishing-links' AND b.sort_order=1);

WITH displaced AS (
  SELECT c.id FROM chats c JOIN topics t ON t.id=c.topic_id JOIN levels l ON l.id=c.level_id
  WHERE t.slug='buyer-prepayment' AND l.level_number=1 AND c.content_status='published'
)
UPDATE chats SET content_status='archived',archived_at=NOW(),is_active=FALSE WHERE id IN (SELECT id FROM displaced);

INSERT INTO chats(title,description,difficulty,role,is_active,level_id,user_role,content_status,scam_scheme,product_context,ai_system_prompt,final_rubric,topic_id)
SELECT 'Пользовательский опубликованный Сценарий','Не является seed-контентом','1','buyer',TRUE,l.id,'buyer','published','custom','{"owner":"admin"}'::jsonb,'custom','{}'::jsonb,t.id
FROM topics t JOIN levels l ON l.level_number=1 WHERE t.slug='buyer-prepayment';

WITH u AS (
  INSERT INTO users(username,password_hash,access_role,training_role) VALUES('migration10-old-history','hash','user','buyer') RETURNING id
), old AS (
  SELECT c.id chat_id,c.topic_id,c.level_id,c.user_role FROM chats c JOIN topics t ON t.id=c.topic_id
  WHERE c.content_status='published' AND c.product_context->>'seed_version'='issue-49' ORDER BY c.id LIMIT 1
), a AS (
  INSERT INTO chat_sessions(user_id,chat_id,status,started_at,finished_at,score,max_score,mode,current_step_number,user_role,free_text_count)
  SELECT u.id,old.chat_id,'COMPLETED',NOW(),NOW(),100,100,'scenario',1,old.user_role,0 FROM u,old RETURNING id,user_id
), progress AS (
  INSERT INTO user_level_progress(user_id,level_id,user_role,best_score,stars,attempts,passed_at,topic_id)
  SELECT a.user_id,old.level_id,old.user_role,100,3,1,NOW(),old.topic_id FROM a,old
)
INSERT INTO attempt_results(attempt_id,result) SELECT id,'{"score":100,"stars":3}'::jsonb FROM a;
SQL

PROJECT_ROOT="$project_root" POSTGRES_DB="$test_database" "${compose[@]}" run --rm anti-scam-trainer-migrate \
  -path=/migrations -database="$database_url" up 1 >/dev/null

"${psql_base[@]}" -d "$test_database" <<'SQL'
DO $$ BEGIN
  IF (SELECT COUNT(*) FROM chats WHERE content_status='published' AND archived_at IS NULL) <> 48 THEN RAISE EXCEPTION 'published curriculum count changed'; END IF;
  IF (SELECT COUNT(*) FROM chats WHERE content_status='published' AND product_context->>'owner'='admin') <> 1 THEN RAISE EXCEPTION 'admin-authored scenario was replaced'; END IF;
  IF (SELECT COUNT(*) FROM attempt_results r JOIN chat_sessions a ON a.id=r.attempt_id JOIN users u ON u.id=a.user_id WHERE u.username='migration10-old-history') <> 1 THEN RAISE EXCEPTION 'old Result was lost on up'; END IF;
END $$;

WITH u AS (
  INSERT INTO users(username,password_hash,access_role,training_role) VALUES('migration10-new-history','hash','user','seller') RETURNING id
), fresh AS (
  SELECT c.id chat_id,c.user_role FROM chats c WHERE c.product_context->>'content_version'='issue-103-complete' ORDER BY c.id LIMIT 1
), a AS (
  INSERT INTO chat_sessions(user_id,chat_id,status,started_at,finished_at,score,max_score,mode,current_step_number,user_role,free_text_count)
  SELECT u.id,fresh.chat_id,'COMPLETED',NOW(),NOW(),75,100,'scenario',1,fresh.user_role,0 FROM u,fresh RETURNING id
)
INSERT INTO attempt_results(attempt_id,result) SELECT id,'{"score":75,"stars":2}'::jsonb FROM a;
SQL

PROJECT_ROOT="$project_root" POSTGRES_DB="$test_database" "${compose[@]}" run --rm anti-scam-trainer-migrate \
  -path=/migrations -database="$database_url" down 1 >/dev/null

"${psql_base[@]}" -d "$test_database" <<'SQL'
DO $$ BEGIN
  IF (SELECT COUNT(*) FROM attempt_results r JOIN chat_sessions a ON a.id=r.attempt_id JOIN users u ON u.id=a.user_id WHERE u.username IN ('migration10-old-history','migration10-new-history')) <> 2 THEN RAISE EXCEPTION 'historical Results were lost on down'; END IF;
  IF (SELECT COUNT(*) FROM chats c JOIN chat_sessions a ON a.chat_id=c.id JOIN users u ON u.id=a.user_id WHERE u.username='migration10-new-history' AND c.content_status='archived') <> 1 THEN RAISE EXCEPTION 'used replacement scenario was not retained as archive'; END IF;
  IF (SELECT kind FROM theory_blocks b JOIN topics t ON t.id=b.topic_id WHERE t.slug='buyer-phishing-links' AND b.sort_order=1) <> 'risk' THEN RAISE EXCEPTION 'Theory kind was not restored'; END IF;
  IF (SELECT COUNT(*) FROM chats WHERE content_status='published' AND product_context->>'owner'='admin') <> 1 THEN RAISE EXCEPTION 'admin-authored scenario changed on down'; END IF;
END $$;
SQL

PROJECT_ROOT="$project_root" POSTGRES_DB="$test_database" "${compose[@]}" run --rm anti-scam-trainer-migrate \
  -path=/migrations -database="$database_url" up 1 >/dev/null

"${psql_base[@]}" -d "$test_database" <<'SQL'
DO $$ BEGIN
  IF (SELECT COUNT(*) FROM chats WHERE content_status='published' AND archived_at IS NULL) <> 48 THEN RAISE EXCEPTION 'published curriculum count changed after re-up'; END IF;
  IF (SELECT COUNT(*) FROM attempt_results r JOIN chat_sessions a ON a.id=r.attempt_id JOIN users u ON u.id=a.user_id WHERE u.username IN ('migration10-old-history','migration10-new-history')) <> 2 THEN RAISE EXCEPTION 'historical Results were lost after re-up'; END IF;
END $$;
SQL

echo "migration 000010 acceptance: ok"
