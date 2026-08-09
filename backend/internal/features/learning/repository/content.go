package repository

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"anti-scam-trainer/backend/internal/features/learning/service"
	"time"

	"github.com/go-pg/pg"
)

func (r *PostgresRepository) ListContent() ([]domain.Topic, error) {
	var rows []topicRow
	_, err := r.db.Query(&rows, `SELECT id,slug,user_role,title,description,sort_order,content_status,archived_at FROM topics ORDER BY user_role,sort_order,id`)
	result := make([]domain.Topic, len(rows))
	for i, row := range rows {
		result[i] = topicFromRow(row)
	}
	return result, mapContentError(err)
}

func (r *PostgresRepository) Content(id int) (domain.TopicContent, error) {
	var row topicRow
	_, err := r.db.QueryOne(&row, `SELECT id,slug,user_role,title,description,sort_order,content_status,archived_at FROM topics WHERE id=?`, id)
	if err != nil {
		return domain.TopicContent{}, mapContentError(err)
	}
	theory, err := r.Theory(id)
	if err != nil {
		return domain.TopicContent{}, err
	}
	quiz, err := r.quiz(id, true)
	if err != nil {
		return domain.TopicContent{}, err
	}
	return domain.TopicContent{Topic: topicFromRow(row), Theory: theory, Quiz: quiz}, nil
}

func (r *PostgresRepository) CreateTopic(topic domain.Topic) (domain.Topic, error) {
	_, err := r.db.QueryOne(pg.Scan(&topic.ID), `INSERT INTO topics(slug,user_role,title,description,sort_order,content_status) VALUES(?,?,?,?,?,'draft') RETURNING id`, topic.Slug, topic.UserRole, topic.Title, topic.Description, topic.SortOrder)
	topic.Status = domain.TopicStatusDraft
	return topic, mapContentError(err)
}
func (r *PostgresRepository) UpdateTopic(topic domain.Topic) error {
	result, err := r.db.Exec(`UPDATE topics SET slug=?,user_role=?,title=?,description=?,sort_order=? WHERE id=? AND content_status='draft'`, topic.Slug, topic.UserRole, topic.Title, topic.Description, topic.SortOrder, topic.ID)
	if err == nil && result.RowsAffected() == 0 {
		err = service.ErrContentConflict
	}
	return mapContentError(err)
}
func (r *PostgresRepository) SetTopicStatus(id int, status string) error {
	var archived any
	if status == domain.TopicStatusArchived {
		archived = time.Now().UTC()
	}
	_, err := r.db.Exec(`UPDATE topics SET content_status=?,archived_at=? WHERE id=?`, status, archived, id)
	return mapContentError(err)
}
func (r *PostgresRepository) CreateTheoryBlock(block domain.TheoryBlock) (domain.TheoryBlock, error) {
	err := r.withDraftTopic(block.TopicID, func(tx *pg.Tx) error {
		_, err := tx.QueryOne(pg.Scan(&block.ID), `INSERT INTO theory_blocks(topic_id,sort_order,kind,title,body) VALUES(?,?,?,?,?) RETURNING id`, block.TopicID, block.SortOrder, block.Kind, block.Title, block.Body)
		return err
	})
	return block, mapContentError(err)
}
func (r *PostgresRepository) UpdateTheoryBlock(block domain.TheoryBlock) error {
	err := r.withDraftTopic(block.TopicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`UPDATE theory_blocks SET sort_order=?,kind=?,title=?,body=? WHERE id=? AND topic_id=?`, block.SortOrder, block.Kind, block.Title, block.Body, block.ID, block.TopicID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}
func (r *PostgresRepository) DeleteTheoryBlock(topicID, blockID int) error {
	err := r.withDraftTopic(topicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`DELETE FROM theory_blocks WHERE id=? AND topic_id=?`, blockID, topicID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}
func (r *PostgresRepository) CreateQuizQuestion(q domain.QuizQuestion) (domain.QuizQuestion, error) {
	err := r.withDraftTopic(q.TopicID, func(tx *pg.Tx) error {
		_, err := tx.QueryOne(pg.Scan(&q.ID), `INSERT INTO quiz_questions(topic_id,sort_order,text,explanation) VALUES(?,?,?,?) RETURNING id`, q.TopicID, q.SortOrder, q.Text, q.Explanation)
		return err
	})
	return q, mapContentError(err)
}
func (r *PostgresRepository) UpdateQuizQuestion(q domain.QuizQuestion) error {
	err := r.withDraftTopic(q.TopicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`UPDATE quiz_questions SET sort_order=?,text=?,explanation=? WHERE id=? AND topic_id=?`, q.SortOrder, q.Text, q.Explanation, q.ID, q.TopicID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}
func (r *PostgresRepository) DeleteQuizQuestion(topicID, questionID int) error {
	err := r.withDraftTopic(topicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`DELETE FROM quiz_questions WHERE id=? AND topic_id=?`, questionID, topicID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}
func (r *PostgresRepository) CreateQuizOption(o domain.QuizOption) (domain.QuizOption, error) {
	var topicID int
	if _, err := r.db.QueryOne(pg.Scan(&topicID), `SELECT topic_id FROM quiz_questions WHERE id=?`, o.QuestionID); err != nil {
		return domain.QuizOption{}, mapContentError(err)
	}
	err := r.withDraftTopic(topicID, func(tx *pg.Tx) error {
		_, err := tx.QueryOne(pg.Scan(&o.ID), `INSERT INTO quiz_options(question_id,sort_order,text,is_correct) VALUES(?,?,?,?) RETURNING id`, o.QuestionID, o.SortOrder, o.Text, o.Correct)
		return err
	})
	return o, mapContentError(err)
}
func (r *PostgresRepository) UpdateQuizOption(o domain.QuizOption) error {
	var topicID int
	if _, err := r.db.QueryOne(pg.Scan(&topicID), `SELECT topic_id FROM quiz_questions WHERE id=?`, o.QuestionID); err != nil {
		return mapContentError(err)
	}
	err := r.withDraftTopic(topicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`UPDATE quiz_options SET sort_order=?,text=?,is_correct=? WHERE id=? AND question_id=?`, o.SortOrder, o.Text, o.Correct, o.ID, o.QuestionID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}
func (r *PostgresRepository) DeleteQuizOption(questionID, optionID int) error {
	var topicID int
	if _, err := r.db.QueryOne(pg.Scan(&topicID), `SELECT topic_id FROM quiz_questions WHERE id=?`, questionID); err != nil {
		return mapContentError(err)
	}
	err := r.withDraftTopic(topicID, func(tx *pg.Tx) error {
		res, err := tx.Exec(`DELETE FROM quiz_options WHERE id=? AND question_id=?`, optionID, questionID)
		if err == nil && res.RowsAffected() == 0 {
			err = service.ErrTopicNotFound
		}
		return err
	})
	return mapContentError(err)
}

func (r *PostgresRepository) PublishTopic(id int) error {
	return r.db.RunInTransaction(func(tx *pg.Tx) error {
		var status string
		if _, err := tx.QueryOne(pg.Scan(&status), `SELECT content_status FROM topics WHERE id=? FOR UPDATE`, id); err != nil {
			return mapContentError(err)
		}
		if status != domain.TopicStatusDraft {
			return service.ErrContentConflict
		}
		var valid bool
		_, err := tx.QueryOne(pg.Scan(&valid), `SELECT
			(SELECT COUNT(*)=5 AND MIN(sort_order)=1 AND MAX(sort_order)=5 FROM theory_blocks WHERE topic_id=t.id)
			AND (SELECT COUNT(*)=5 AND MIN(sort_order)=1 AND MAX(sort_order)=5 FROM quiz_questions WHERE topic_id=t.id)
			AND NOT EXISTS (SELECT 1 FROM quiz_questions q WHERE q.topic_id=t.id AND ((SELECT COUNT(*) FROM quiz_options o WHERE o.question_id=q.id)<>4 OR (SELECT COUNT(*) FROM quiz_options o WHERE o.question_id=q.id AND o.is_correct)<>1 OR (SELECT MIN(sort_order)=1 AND MAX(sort_order)=4 FROM quiz_options o WHERE o.question_id=q.id) IS NOT TRUE))
			AND (SELECT COUNT(DISTINCT l.level_number)=4 FROM chats c JOIN levels l ON l.id=c.level_id WHERE c.topic_id=t.id AND c.user_role=t.user_role AND c.content_status='published' AND c.archived_at IS NULL)
			FROM topics t WHERE t.id=? AND t.content_status='draft'`, id)
		if err != nil {
			return mapContentError(err)
		}
		if !valid {
			return service.ErrContentConflict
		}
		_, err = tx.Exec(`UPDATE topics SET content_status='published',archived_at=NULL WHERE id=?`, id)
		return mapContentError(err)
	})
}

func (r *PostgresRepository) withDraftTopic(topicID int, action func(*pg.Tx) error) error {
	return r.db.RunInTransaction(func(tx *pg.Tx) error {
		var status string
		if _, err := tx.QueryOne(pg.Scan(&status), `SELECT content_status FROM topics WHERE id=? FOR UPDATE`, topicID); err != nil {
			return mapContentError(err)
		}
		if status != domain.TopicStatusDraft {
			return service.ErrContentConflict
		}
		return mapContentError(action(tx))
	})
}

func mapContentError(err error) error {
	if err == pg.ErrNoRows {
		return service.ErrTopicNotFound
	}
	if pgErr, ok := err.(pg.Error); ok && pgErr.IntegrityViolation() {
		return service.ErrContentConflict
	}
	return err
}
