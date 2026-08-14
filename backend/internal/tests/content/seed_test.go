package content_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	learningrepository "anti-scam-trainer/backend/internal/features/learning/repository"
	learningservice "anti-scam-trainer/backend/internal/features/learning/service"
	scenariosrepository "anti-scam-trainer/backend/internal/features/scenarios/repository"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-pg/pg"
)

// TestPublishedContentMatrix runs against the disposable migrated database used by acceptance.
func TestPublishedContentMatrix(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer func() { _ = db.Close() }()
	var topics, theory, quiz, scenarios int
	_, err := db.QueryOne(pg.Scan(&topics, &theory, &quiz, &scenarios), `SELECT (SELECT COUNT(*) FROM topics),(SELECT COUNT(*) FROM theory_blocks),(SELECT COUNT(*) FROM quiz_questions),(SELECT COUNT(*) FROM chats WHERE content_status='published' AND archived_at IS NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	if topics != 12 || theory != 60 || quiz != 60 || scenarios != 48 {
		t.Fatalf("content counts=(%d,%d,%d,%d), want (12,60,60,48)", topics, theory, quiz, scenarios)
	}
	var invalid int
	var genericDialogueFragments int
	_, err = db.QueryOne(pg.Scan(&genericDialogueFragments), `SELECT COUNT(*) FROM (
		SELECT c.id,l.level_number,COUNT(DISTINCT s.id) steps,MIN(s.step_number) first_step,MAX(s.step_number) last_step,
			COUNT(DISTINCT o.id) options,
			COUNT(DISTINCT s.id) FILTER (WHERE s.response_type='multiple_choice') multiple_choice_steps,
			COUNT(DISTINCT s.id) FILTER (WHERE s.response_type='similar_choice') similar_choice_steps,
			COUNT(DISTINCT s.id) FILTER (WHERE s.response_type='free_text') free_text_steps
		FROM chats c JOIN levels l ON l.id=c.level_id JOIN chat_steps s ON s.chat_id=c.id LEFT JOIN chat_options o ON o.step_id=s.id
		WHERE c.content_status='published' AND c.archived_at IS NULL GROUP BY c.id,l.level_number
		HAVING COUNT(DISTINCT s.id)<>CASE l.level_number WHEN 1 THEN 3 WHEN 2 THEN 2 WHEN 3 THEN 3 ELSE 5 END OR MIN(s.step_number)<>1
			OR MAX(s.step_number)<>CASE l.level_number WHEN 1 THEN 3 WHEN 2 THEN 2 WHEN 3 THEN 3 ELSE 5 END
			OR COUNT(DISTINCT o.id)<>CASE l.level_number WHEN 1 THEN 9 WHEN 2 THEN 6 WHEN 3 THEN 3 ELSE 0 END
			OR COUNT(DISTINCT s.id) FILTER (WHERE s.response_type='multiple_choice')<>CASE l.level_number WHEN 1 THEN 3 WHEN 3 THEN 1 ELSE 0 END
			OR COUNT(DISTINCT s.id) FILTER (WHERE s.response_type='similar_choice')<>CASE WHEN l.level_number=2 THEN 2 ELSE 0 END
			OR COUNT(DISTINCT s.id) FILTER (WHERE s.response_type='free_text')<>CASE l.level_number WHEN 3 THEN 2 WHEN 4 THEN 5 ELSE 0 END
	) bad`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("invalid scenario structures=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM chats c
		WHERE c.content_status='published' AND c.archived_at IS NULL AND (
			c.risk_type NOT IN ('phishing','prepayment','fake_payment','delivery','external_messenger','account_takeover','sms_code','social_engineering')
			OR NULLIF(c.product_context->>'item_title','') IS NULL
			OR NULLIF(c.product_context->>'category','') IS NULL
			OR NULLIF(c.product_context->>'deal_method','') IS NULL
			OR c.product_context ? 'url'
		)`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("published scenarios with invalid risk or product context=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM (
		SELECT s.id FROM chat_steps s JOIN chats c ON c.id=s.chat_id JOIN levels l ON l.id=c.level_id LEFT JOIN chat_options o ON o.step_id=s.id
		WHERE c.content_status='published' GROUP BY s.id,l.level_number,s.step_number HAVING COUNT(o.id)<>CASE WHEN l.level_number IN(1,2) OR (l.level_number=3 AND s.step_number=1) THEN 3 ELSE 0 END
	) bad`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("steps with an invalid option count=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id WHERE c.content_status='published' AND (o.points NOT IN(0,25,50,75,100) OR char_length(o.option_text)>140)`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("invalid options=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM chat_options l2o JOIN chat_steps l2s ON l2s.id=l2o.step_id JOIN chats l2c ON l2c.id=l2s.chat_id JOIN levels l2l ON l2l.id=l2c.level_id WHERE l2c.content_status='published' AND l2l.level_number=2 AND (l2o.points NOT IN(0,25,50,75,100) OR EXISTS(SELECT 1 FROM chat_options l1o JOIN chat_steps l1s ON l1s.id=l1o.step_id JOIN chats l1c ON l1c.id=l1s.chat_id JOIN levels l1l ON l1l.id=l1c.level_id WHERE l1c.topic_id=l2c.topic_id AND l1l.level_number=1 AND l1o.option_text=l2o.option_text))`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("level 2 has non-similar-choice scoring or duplicates level 1 options: %d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM (
		SELECT c.topic_id FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id
		WHERE c.content_status='published' AND c.archived_at IS NULL
		GROUP BY c.topic_id HAVING COUNT(*)<>COUNT(DISTINCT o.option_text)
	) repeated`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("topics with repeated prepared replies across steps=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM chat_steps s JOIN chats c ON c.id=s.chat_id WHERE c.content_status='published' AND (NULLIF(s.counterparty_message,'') IS NULL OR char_length(s.counterparty_message)>280 OR (s.response_type IN('mixed','free_text') AND (NULLIF(s.ai_instruction,'') IS NULL OR NULLIF(s.fallback_message,'') IS NULL)))`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("invalid steps=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM (
		SELECT title text FROM topics UNION ALL SELECT description FROM topics UNION ALL SELECT body FROM theory_blocks
		UNION ALL SELECT text FROM quiz_questions UNION ALL SELECT text FROM quiz_options
		UNION ALL SELECT counterparty_message FROM chat_steps UNION ALL SELECT fallback_message FROM chat_steps
		UNION ALL SELECT option_text FROM chat_options
	) content WHERE text ~* '(https?://|www\.|[0-9]{10,})' OR text ~* '(avito доставка|безопасная сделка).{0,20}(опасн|мошенн)'`)
	if err != nil {
		t.Fatal(err)
	}
	if invalid != 0 {
		t.Fatalf("unsafe content fragments=%d", invalid)
	}
	_, err = db.QueryOne(pg.Scan(&invalid), `SELECT COUNT(*) FROM (
		SELECT option_text AS text FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id
		WHERE c.content_status='published' AND c.archived_at IS NULL
		UNION ALL
		SELECT counterparty_reaction FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id
		WHERE c.content_status='published' AND c.archived_at IS NULL AND counterparty_reaction IS NOT NULL
		UNION ALL
		SELECT counterparty_message FROM chat_steps s JOIN chats c ON c.id=s.chat_id
		WHERE c.content_status='published' AND c.archived_at IS NULL
	) content WHERE text ~ '^(Точно:|Сразу скажу:)' OR text IN (
		'Откройте форму получения 67 000 ₽ и подтвердите свою карту.',
		'Проверяйте, но я всё равно предлагаю оформить быстрее.',
		'Хорошо, но для продолжения всё равно понадобится подтверждение.',
		'Тогда переходите к оформлению по моей инструкции.',
		'Собеседник не принимает отказ и начинает торопить.',
		'Собеседник усиливает срочность.',
		'Собеседник требует завершить действие немедленно.',
		'Это моё окончательное решение.',
		'Пока продолжать сделку не буду.',
		'Так и начну разговор.',
		'Сначала ограничусь этим ответом.',
		'С этого и начну.'
	)`)
	if err != nil {
		t.Fatal(err)
	}
	if genericDialogueFragments != 0 {
		t.Fatalf("published scenarios with mechanical prefixes or generic reactions=%d", genericDialogueFragments)
	}
	var theorySignatures, quizSignatures, scenarioSignatures int
	_, err = db.QueryOne(pg.Scan(&theorySignatures, &quizSignatures, &scenarioSignatures), `SELECT
		(SELECT COUNT(DISTINCT signature) FROM (SELECT topic_id,string_agg(body,'|' ORDER BY sort_order) signature FROM theory_blocks GROUP BY topic_id) x),
		(SELECT COUNT(DISTINCT signature) FROM (SELECT topic_id,string_agg(text||' '||explanation,'|' ORDER BY sort_order) signature FROM quiz_questions GROUP BY topic_id) x),
		(SELECT COUNT(DISTINCT signature) FROM (SELECT c.topic_id,string_agg(s.counterparty_message,'|' ORDER BY l.level_number,s.step_number) signature FROM chats c JOIN levels l ON l.id=c.level_id JOIN chat_steps s ON s.chat_id=c.id WHERE c.content_status='published' GROUP BY c.topic_id) x)`)
	if err != nil {
		t.Fatal(err)
	}
	if theorySignatures != 12 || quizSignatures != 12 || scenarioSignatures != 12 {
		t.Fatalf("topic-specific signatures=(%d,%d,%d), want all 12", theorySignatures, quizSignatures, scenarioSignatures)
	}

	var scenarioIDs []int
	if _, err := db.Query(&scenarioIDs, `SELECT id FROM chats WHERE content_status='published' AND archived_at IS NULL ORDER BY id`); err != nil {
		t.Fatal(err)
	}
	validator := scenariosrepository.NewPostgres(db)
	for _, scenarioID := range scenarioIDs {
		valid, err := validator.ValidContent(scenarioID)
		if err != nil {
			t.Fatalf("validate scenario %d: %v", scenarioID, err)
		}
		if !valid {
			t.Fatalf("published scenario %d does not pass publication validator", scenarioID)
		}
	}
}

func TestAvitoScenariosReplaceTenTemplatesAndPreserveArchivedReferences(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer func() { _ = db.Close() }()
	var publishedReplacements, archivedTemplates, preservedPublished, steps, options, reactions int
	_, err := db.QueryOne(pg.Scan(&publishedReplacements, &archivedTemplates, &preservedPublished, &steps, &options, &reactions), `SELECT
		(SELECT COUNT(*) FROM chats WHERE content_status='published' AND archived_at IS NULL AND product_context?'content_key'),
		(SELECT COUNT(*) FROM migration_000009_archived_chats m JOIN chats c ON c.id=m.id WHERE c.content_status='archived'),
		(SELECT COUNT(*) FROM chats WHERE content_status='published' AND archived_at IS NULL AND NOT (product_context?'content_key')),
		(SELECT COUNT(*) FROM chat_steps s JOIN chats c ON c.id=s.chat_id WHERE c.product_context?'content_key'),
		(SELECT COUNT(*) FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id WHERE c.product_context?'content_key'),
		(SELECT COUNT(*) FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id WHERE c.product_context?'content_key' AND o.counterparty_reaction IS NOT NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	if publishedReplacements != 0 || archivedTemplates != 10 || preservedPublished != 48 || steps != 24 || options != 60 || reactions != 36 {
		t.Fatalf("replacement counts=(published=%d archived=%d preserved=%d steps=%d options=%d reactions=%d)", publishedReplacements, archivedTemplates, preservedPublished, steps, options, reactions)
	}
	var exact int
	_, err = db.QueryOne(pg.Scan(&exact), `SELECT COUNT(*) FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id
		WHERE c.product_context->>'content_key'='buyer-l1-fake-delivery' AND s.step_number=1 AND o.sort_order=1
		AND o.option_text='Можно оформить покупку через Avito Доставку внутри приложения?'
		AND o.counterparty_reaction='Можно, но обычное оформление у меня сейчас почему-то не работает.' AND o.points=100`)
	if err != nil || exact != 1 {
		t.Fatalf("prepared option exact match=%d err=%v", exact, err)
	}
	var contentDigest string
	_, err = db.QueryOne(pg.Scan(&contentDigest), `SELECT md5(string_agg(row_data, E'\n' ORDER BY row_data)) FROM (
		SELECT concat_ws('|','chat',c.product_context->>'content_key',c.title,c.description,c.user_role,l.level_number,c.scam_scheme,c.product_context::text,c.ai_system_prompt,c.final_rubric::text) row_data
		FROM chats c JOIN levels l ON l.id=c.level_id WHERE c.product_context?'content_key'
		UNION ALL
		SELECT concat_ws('|','step',c.product_context->>'content_key',s.step_number,s.response_type,s.counterparty_message,coalesce(s.ai_instruction,''),s.fallback_message,s.max_points)
		FROM chat_steps s JOIN chats c ON c.id=s.chat_id WHERE c.product_context?'content_key'
		UNION ALL
		SELECT concat_ws('|','option',c.product_context->>'content_key',s.step_number,o.sort_order,o.option_text,coalesce(o.counterparty_reaction,''),o.explanation,o.points)
		FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id WHERE c.product_context?'content_key'
	) prepared_content`)
	if err != nil || contentDigest != "c09ec77fc4205c931ca810c19d9b6877" {
		t.Fatalf("prepared content digest=%q err=%v; every scenario, step, option, reaction, and explanation must match the approved fixture", contentDigest, err)
	}
	var userID, archivedID, attemptID int
	username := fmt.Sprintf("archived-scenario-reference-%d", time.Now().UnixNano())
	_, err = db.QueryOne(pg.Scan(&userID), `INSERT INTO users(username,password_hash,access_role,training_role) VALUES(?,'hash','user','buyer') RETURNING id`, username)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = db.Exec(`DELETE FROM users WHERE id=?`, userID) }()
	_, err = db.QueryOne(pg.Scan(&archivedID), `SELECT id FROM migration_000009_archived_chats ORDER BY id LIMIT 1`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.QueryOne(pg.Scan(&attemptID), `INSERT INTO chat_sessions(user_id,chat_id,status,started_at,current_step_number,mode,user_role,free_text_count) SELECT ?,c.id,'ABANDONED',NOW(),1,'scenario',c.user_role,0 FROM chats c WHERE c.id=? RETURNING id`, userID, archivedID)
	if err != nil || attemptID == 0 {
		t.Fatalf("historical attempt reference=(%d,%v)", attemptID, err)
	}
}

func TestCompleteAvitoCurriculumReplacesEveryPublishedScenario(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer func() { _ = db.Close() }()

	var topics, theoryBlocks, questions, quizOptions, scenarios, steps, options, reactions, archived int
	_, err := db.QueryOne(pg.Scan(&topics, &theoryBlocks, &questions, &quizOptions, &scenarios, &steps, &options, &reactions, &archived), `SELECT
		(SELECT COUNT(*) FROM topics WHERE content_status='published'),
		(SELECT COUNT(*) FROM theory_blocks b JOIN topics t ON t.id=b.topic_id WHERE t.content_status='published'),
		(SELECT COUNT(*) FROM quiz_questions q JOIN topics t ON t.id=q.topic_id WHERE t.content_status='published'),
		(SELECT COUNT(*) FROM quiz_options o JOIN quiz_questions q ON q.id=o.question_id JOIN topics t ON t.id=q.topic_id WHERE t.content_status='published'),
		(SELECT COUNT(*) FROM chats WHERE content_status='published' AND archived_at IS NULL AND product_context->>'content_version'='issue-103-complete'),
		(SELECT COUNT(*) FROM chat_steps s JOIN chats c ON c.id=s.chat_id WHERE c.content_status='published' AND c.product_context->>'content_version'='issue-103-complete'),
		(SELECT COUNT(*) FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id WHERE c.content_status='published' AND c.product_context->>'content_version'='issue-103-complete'),
		(SELECT COUNT(*) FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id WHERE c.content_status='published' AND c.product_context->>'content_version'='issue-103-complete' AND o.counterparty_reaction IS NOT NULL),
		(SELECT COUNT(*) FROM migration_000010_archived_chats)`)
	if err != nil {
		t.Fatal(err)
	}
	if topics != 12 || theoryBlocks != 60 || questions != 60 || quizOptions != 240 || scenarios != 48 || steps != 156 || options != 216 || reactions != 144 || archived != 48 {
		t.Fatalf("complete curriculum=(topics=%d theory=%d questions=%d quiz_options=%d scenarios=%d steps=%d options=%d reactions=%d archived=%d)", topics, theoryBlocks, questions, quizOptions, scenarios, steps, options, reactions, archived)
	}

	var invalidQuiz, invalidLevels, legacyPublished int
	_, err = db.QueryOne(pg.Scan(&invalidQuiz, &invalidLevels, &legacyPublished), `SELECT
		(SELECT COUNT(*) FROM (SELECT q.id FROM quiz_questions q JOIN quiz_options o ON o.question_id=q.id GROUP BY q.id HAVING COUNT(*)<>4 OR COUNT(*) FILTER (WHERE o.is_correct)<>1) invalid),
		(SELECT COUNT(*) FROM (SELECT c.topic_id,l.level_number,COUNT(*) amount FROM chats c JOIN levels l ON l.id=c.level_id WHERE c.content_status='published' AND c.archived_at IS NULL GROUP BY c.topic_id,l.level_number HAVING COUNT(*)<>1) invalid),
		(SELECT COUNT(*) FROM chats WHERE content_status='published' AND archived_at IS NULL AND product_context->>'content_version' IS DISTINCT FROM 'issue-103-complete')`)
	if err != nil || invalidQuiz != 0 || invalidLevels != 0 || legacyPublished != 0 {
		t.Fatalf("curriculum invariants=(quiz=%d levels=%d legacy=%d err=%v)", invalidQuiz, invalidLevels, legacyPublished, err)
	}
	var directReply string
	_, err = db.QueryOne(pg.Scan(&directReply), `SELECT o.option_text FROM chat_options o
		JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id JOIN topics t ON t.id=c.topic_id JOIN levels l ON l.id=c.level_id
		WHERE t.slug='buyer-phishing-links' AND l.level_number=1 AND s.step_number=1 AND o.sort_order=1 AND c.content_status='published'`)
	if err != nil || directReply != "Да, доставка подходит. Оформите заказ через Avito, я проверю его в приложении." {
		t.Fatalf("prepared option is not a direct user reply: %q err=%v", directReply, err)
	}
	var thinTheory, duplicateQuizOptions, repeatedReaction, distinctQuestions, distinctQuizSets int
	_, err = db.QueryOne(pg.Scan(&thinTheory, &duplicateQuizOptions, &repeatedReaction, &distinctQuestions, &distinctQuizSets), `SELECT
		(SELECT COUNT(*) FROM theory_blocks b JOIN topics t ON t.id=b.topic_id WHERE t.content_status='published' AND char_length(b.body)<110),
		(SELECT COUNT(*) FROM (SELECT q.id FROM quiz_questions q JOIN quiz_options o ON o.question_id=q.id GROUP BY q.id HAVING COUNT(DISTINCT o.text)<>4) invalid),
		(SELECT COUNT(*) FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id JOIN chat_steps next ON next.chat_id=s.chat_id AND next.step_number=s.step_number+1 WHERE c.content_status='published' AND o.counterparty_reaction IS NOT NULL AND o.counterparty_reaction=next.counterparty_message),
		(SELECT COUNT(DISTINCT q.text) FROM quiz_questions q JOIN topics t ON t.id=q.topic_id WHERE t.content_status='published'),
		(SELECT COUNT(DISTINCT signature) FROM (SELECT q.id,string_agg(o.text,'|' ORDER BY o.sort_order) signature FROM quiz_questions q JOIN quiz_options o ON o.question_id=q.id GROUP BY q.id) sets)`)
	if err != nil || thinTheory != 0 || duplicateQuizOptions != 0 || repeatedReaction != 0 || distinctQuestions != 60 || distinctQuizSets != 60 {
		t.Fatalf("curriculum quality=(thin_theory=%d duplicate_options=%d repeated_reaction=%d questions=%d option_sets=%d err=%v)", thinTheory, duplicateQuizOptions, repeatedReaction, distinctQuestions, distinctQuizSets, err)
	}

	var digest string
	_, err = db.QueryOne(pg.Scan(&digest), `SELECT md5(string_agg(row_data,E'\n' ORDER BY row_data)) FROM (
		SELECT concat_ws('|','topic',t.slug,t.user_role,t.title,t.description,t.sort_order) row_data FROM topics t WHERE t.content_status='published'
		UNION ALL SELECT concat_ws('|','theory',t.slug,b.sort_order,b.kind,b.title,b.body) FROM theory_blocks b JOIN topics t ON t.id=b.topic_id WHERE t.content_status='published'
		UNION ALL SELECT concat_ws('|','question',t.slug,q.sort_order,q.text,q.explanation) FROM quiz_questions q JOIN topics t ON t.id=q.topic_id WHERE t.content_status='published'
		UNION ALL SELECT concat_ws('|','quiz_option',t.slug,q.sort_order,o.sort_order,o.text,o.is_correct) FROM quiz_options o JOIN quiz_questions q ON q.id=o.question_id JOIN topics t ON t.id=q.topic_id WHERE t.content_status='published'
		UNION ALL SELECT concat_ws('|','scenario',t.slug,l.level_number,c.title,c.description,c.scam_scheme,c.product_context::text,c.ai_system_prompt,c.final_rubric::text) FROM chats c JOIN topics t ON t.id=c.topic_id JOIN levels l ON l.id=c.level_id WHERE c.content_status='published' AND c.archived_at IS NULL
		UNION ALL SELECT concat_ws('|','step',t.slug,l.level_number,s.step_number,s.response_type,s.step_goal,s.counterparty_message,coalesce(s.ai_instruction,''),s.fallback_message,s.max_points) FROM chat_steps s JOIN chats c ON c.id=s.chat_id JOIN topics t ON t.id=c.topic_id JOIN levels l ON l.id=c.level_id WHERE c.content_status='published' AND c.archived_at IS NULL
		UNION ALL SELECT concat_ws('|','scenario_option',t.slug,l.level_number,s.step_number,o.sort_order,o.option_text,coalesce(o.counterparty_reaction,''),o.explanation,o.points) FROM chat_options o JOIN chat_steps s ON s.id=o.step_id JOIN chats c ON c.id=s.chat_id JOIN topics t ON t.id=c.topic_id JOIN levels l ON l.id=c.level_id WHERE c.content_status='published' AND c.archived_at IS NULL
		UNION ALL SELECT concat_ws('|','free_play',user_role,product_context::text,system_prompt,final_rubric::text) FROM free_play_configs
	) curriculum`)
	if err != nil || digest != "22ce37b549873d5dfe118c1196f6a30a" {
		t.Fatalf("complete curriculum digest=%q err=%v", digest, err)
	}

	var userID, archivedScenarioID, historicalAttemptID int
	username := fmt.Sprintf("complete-curriculum-history-%d", time.Now().UnixNano())
	_, err = db.QueryOne(pg.Scan(&userID), `INSERT INTO users(username,password_hash,access_role,training_role) VALUES(?,'hash','user','buyer') RETURNING id`, username)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = db.Exec(`DELETE FROM users WHERE id=?`, userID) }()
	_, err = db.QueryOne(pg.Scan(&archivedScenarioID), `SELECT id FROM migration_000010_archived_chats ORDER BY id LIMIT 1`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.QueryOne(pg.Scan(&historicalAttemptID), `INSERT INTO chat_sessions(user_id,chat_id,status,started_at,current_step_number,mode,user_role,free_text_count) SELECT ?,c.id,'ABANDONED',NOW(),1,'scenario',c.user_role,0 FROM chats c WHERE c.id=? RETURNING id`, userID, archivedScenarioID)
	if err != nil || historicalAttemptID == 0 {
		t.Fatalf("complete curriculum historical reference=(%d,%v)", historicalAttemptID, err)
	}
}

func TestLearningActivityAwardsStreakAchievements(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer func() { _ = db.Close() }()
	var userID, topicID int
	_, err := db.QueryOne(pg.Scan(&userID), `INSERT INTO users(username,password_hash,access_role,training_role,current_streak,longest_streak,last_activity_date) VALUES('streak-learning-test','hash','user','buyer',2,2,'2026-08-08') RETURNING id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = db.Exec(`DELETE FROM users WHERE id=?`, userID) }()
	_, err = db.QueryOne(pg.Scan(&topicID), `SELECT id FROM topics WHERE slug='buyer-phishing-links'`)
	if err != nil {
		t.Fatal(err)
	}
	repository := learningrepository.NewPostgres(db)
	activityDate := time.Date(2026, 8, 9, 0, 0, 0, 0, time.FixedZone("Europe/Moscow", 3*60*60))
	streak, _, err := repository.MarkTheoryRead(userID, topicID, activityDate)
	if err != nil || streak.Current != 3 {
		t.Fatalf("theory streak = (%#v,%v)", streak, err)
	}
	var awarded int
	_, err = db.QueryOne(pg.Scan(&awarded), `SELECT COUNT(*) FROM user_achievements ua JOIN achievements a ON a.id=ua.achievement_id WHERE ua.user_id=? AND a.code='streak_3'`, userID)
	if err != nil || awarded != 1 {
		t.Fatalf("streak_3 awards=%d, err=%v", awarded, err)
	}

	_, err = db.Exec(`UPDATE users SET current_streak=6,longest_streak=6,last_activity_date='2026-08-09' WHERE id=?`, userID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO user_level_progress(user_id,level_id,user_role,topic_id,best_score,stars,attempts,passed_at) SELECT ?,id,'buyer',?,80,1,1,NOW() FROM levels`, userID, topicID)
	if err != nil {
		t.Fatal(err)
	}
	var answers []domain.QuizAnswer
	_, err = db.Query(&answers, `SELECT q.id question_id,o.id option_id FROM quiz_questions q JOIN quiz_options o ON o.question_id=q.id AND o.is_correct=TRUE WHERE q.topic_id=? ORDER BY q.sort_order`, topicID)
	if err != nil {
		t.Fatal(err)
	}
	quizDate := activityDate.AddDate(0, 0, 1)
	if _, err = repository.SubmitQuiz(userID, topicID, answers, quizDate); err != nil {
		t.Fatal(err)
	}
	_, err = db.QueryOne(pg.Scan(&awarded), `SELECT COUNT(*) FROM user_achievements ua JOIN achievements a ON a.id=ua.achievement_id WHERE ua.user_id=? AND a.code='streak_7'`, userID)
	if err != nil || awarded != 1 {
		t.Fatalf("streak_7 awards=%d, err=%v", awarded, err)
	}
	var completed bool
	_, err = db.QueryOne(pg.Scan(&completed), `SELECT completed_at IS NOT NULL FROM user_topic_progress WHERE user_id=? AND topic_id=?`, userID, topicID)
	if err != nil || !completed {
		t.Fatalf("quiz did not persist migrated topic completion: completed=%v err=%v", completed, err)
	}
	_, err = db.QueryOne(pg.Scan(&awarded), `SELECT COUNT(*) FROM user_achievements ua JOIN achievements a ON a.id=ua.achievement_id WHERE ua.user_id=? AND a.code='first_topic_completed'`, userID)
	if err != nil || awarded != 1 {
		t.Fatalf("first_topic_completed awards=%d, err=%v", awarded, err)
	}
}

func TestPostgresDailyTaskAndTopicLifecycle(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer func() { _ = db.Close() }()
	var userID, topicID int
	if _, err := db.QueryOne(pg.Scan(&userID), `INSERT INTO users(username,password_hash,access_role,training_role) VALUES('daily-content-test','hash','user','buyer') RETURNING id`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`UPDATE topics SET content_status='published',archived_at=NULL WHERE id=?`, topicID)
		_, _ = db.Exec(`DELETE FROM users WHERE id=?`, userID)
	}()
	if _, err := db.QueryOne(pg.Scan(&topicID), `SELECT id FROM topics WHERE user_role='buyer' ORDER BY sort_order LIMIT 1`); err != nil {
		t.Fatal(err)
	}
	repository := learningrepository.NewPostgres(db)
	now := time.Date(2026, 8, 9, 9, 0, 0, 0, time.UTC)
	learning := learningservice.NewWithClock(repository, func() time.Time { return now })
	_, _, _, _, first, err := learning.Dashboard(userID, domain.UserRoleBuyer)
	if err != nil || first == nil || len(first.Messages) == 0 || first.Completed || first.Answer != nil || first.Correct != nil {
		t.Fatalf("first daily task=(%#v,%v)", first, err)
	}
	_, _, _, _, refresh, err := learning.Dashboard(userID, domain.UserRoleBuyer)
	if err != nil || refresh.Date != first.Date || refresh.Role != first.Role || refresh.Answer != nil || refresh.Correct != nil {
		t.Fatalf("refreshed daily task=(%#v,%v)", refresh, err)
	}
	if _, _, err = repository.MarkTheoryRead(userID, topicID, now); err != nil {
		t.Fatalf("create topic progress before lifecycle checks: %v", err)
	}
	content := learningservice.NewContent(repository)
	if err = content.Archive(topicID); err != nil {
		t.Fatal(err)
	}
	var progress int
	if _, err = db.QueryOne(pg.Scan(&progress), `SELECT COUNT(*) FROM user_topic_progress WHERE user_id=? AND topic_id=?`, userID, topicID); err != nil || progress != 1 {
		t.Fatalf("archived progress=(%d,%v)", progress, err)
	}
	if err = content.Restore(topicID); err != nil {
		t.Fatal(err)
	}
	if err = content.Publish(topicID); err != nil {
		t.Fatalf("republish complete seed topic: %v", err)
	}
}

func TestProgressStatsExcludeAbandonedAttempts(t *testing.T) {
	database := os.Getenv("POSTGRES_TEST_NAME")
	if database == "" {
		t.Skip("POSTGRES_TEST_NAME is not set")
	}
	db := pg.Connect(&pg.Options{Addr: os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Database: database})
	defer func() { _ = db.Close() }()
	var userID, chatID int
	_, err := db.QueryOne(pg.Scan(&userID), `INSERT INTO users(username,password_hash,access_role,training_role) VALUES('progress-stats-test','hash','user','buyer') RETURNING id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`DELETE FROM chat_sessions WHERE user_id=?`, userID)
		_, _ = db.Exec(`DELETE FROM users WHERE id=?`, userID)
	}()
	_, err = db.QueryOne(pg.Scan(&chatID), `SELECT c.id FROM chats c JOIN topics t ON t.id=c.topic_id JOIN levels l ON l.id=c.level_id WHERE t.slug='buyer-phishing-links' AND l.level_number=1 AND c.content_status='published'`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO chat_sessions(user_id,chat_id,status,mode,current_step_number,score,user_role,finished_at) VALUES
		(?,?,'COMPLETED','scenario',3,80,'buyer',NOW()-INTERVAL '1 minute'),
		(?,?,'ABANDONED','scenario',2,100,'buyer',NOW())`, userID, chatID, userID, chatID)
	if err != nil {
		t.Fatal(err)
	}
	recent, average, err := learningrepository.NewPostgres(db).RecentAttempts(userID, domain.UserRoleBuyer)
	if err != nil || len(recent) != 1 || recent[0].Score != 80 || average != 80 {
		t.Fatalf("progress stats = (%#v,%v,%v), want one completed score 80", recent, average, err)
	}
}
