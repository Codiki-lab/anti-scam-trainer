import { useEffect, useState } from 'react'
import {
  useCreateAdminQuizOptionMutation,
  useCreateAdminQuizQuestionMutation,
  useDeleteAdminQuizOptionMutation,
  useDeleteAdminQuizQuestionMutation,
  useUpdateAdminQuizOptionMutation,
  useUpdateAdminQuizQuestionMutation,
  type AdminQuizOption,
  type AdminQuizQuestion,
  type AdminTopic,
} from '@/entities/admin-content'
import { getApiErrorMessage } from '@/shared/http-error'
import { DangerAction, EditorActions, PrimaryAction, TextAreaField, TextField } from './AdminFields'
import styles from './AdminPage.module.scss'

const emptyQuestion = (sortOrder: number): AdminQuizQuestion => ({
  sortOrder,
  text: '',
  explanation: '',
  options: [],
})

const emptyOption = (sortOrder: number): AdminQuizOption => ({
  sortOrder,
  text: '',
  isCorrect: false,
})

export function QuizEditor({ topic, editable }: { topic: AdminTopic; editable: boolean }) {
  const [adding, setAdding] = useState(false)
  const topicId = topic.id
  if (!topicId) return null

  return (
    <section className={styles.contentSection}>
      <div className={styles.contentSectionHeader}>
        <div>
          <h3>Quiz</h3>
          <p>{topic.quiz.length} из 5 обязательных вопросов, по 4 варианта в каждом</p>
        </div>
        {editable && topic.quiz.length < 5 && !adding && (
          <button type="button" onClick={() => setAdding(true)}>
            + Добавить вопрос
          </button>
        )}
      </div>
      <div className={styles.nestedList}>
        {topic.quiz.map((question) => (
          <QuestionCard
            key={question.id}
            editable={editable}
            question={question}
            topicId={topicId}
          />
        ))}
        {adding && (
          <QuestionCard
            editable
            question={emptyQuestion(topic.quiz.length + 1)}
            topicId={topicId}
            onCancel={() => setAdding(false)}
            onCreated={() => setAdding(false)}
          />
        )}
      </div>
    </section>
  )
}

interface QuestionCardProps {
  topicId: number
  question: AdminQuizQuestion
  editable: boolean
  onCancel?: () => void
  onCreated?: () => void
}

function QuestionCard({ topicId, question, editable, onCancel, onCreated }: QuestionCardProps) {
  const [draft, setDraft] = useState(question)
  const [addingOption, setAddingOption] = useState(false)
  const [message, setMessage] = useState<string>()
  const [createQuestion, createState] = useCreateAdminQuizQuestionMutation()
  const [updateQuestion, updateState] = useUpdateAdminQuizQuestionMutation()
  const [deleteQuestion, deleteState] = useDeleteAdminQuizQuestionMutation()

  useEffect(() => setDraft(question), [question])

  const save = async () => {
    setMessage(undefined)
    try {
      if (question.id) {
        await updateQuestion({ topicId, itemId: question.id, value: draft }).unwrap()
        setMessage('Вопрос сохранён.')
      } else {
        await createQuestion({ topicId, value: draft }).unwrap()
        onCreated?.()
      }
    } catch (requestError) {
      setMessage(getApiErrorMessage(requestError))
    }
  }

  const remove = async () => {
    if (!question.id || !window.confirm('Удалить вопрос Quiz со всеми вариантами?')) return
    try {
      await deleteQuestion({ topicId, itemId: question.id }).unwrap()
    } catch (requestError) {
      setMessage(getApiErrorMessage(requestError))
    }
  }

  const pending = createState.isLoading || updateState.isLoading || deleteState.isLoading
  const questionId = question.id

  return (
    <details className={styles.nestedCard} open={!question.id}>
      <summary>
        <span>{String(draft.sortOrder).padStart(2, '0')}</span>
        <b>{draft.text || 'Новый вопрос'}</b>
        <em>{question.options.length}/4 варианта</em>
      </summary>
      <div className={styles.nestedBody}>
        <div className={styles.formGrid}>
          <TextField
            disabled={!editable}
            label="Порядок"
            max={5}
            min={1}
            type="number"
            value={draft.sortOrder}
            onChange={(event) => setDraft({ ...draft, sortOrder: Number(event.target.value) })}
          />
          <TextAreaField
            className={styles.wideField}
            disabled={!editable}
            label="Вопрос"
            value={draft.text}
            onChange={(event) => setDraft({ ...draft, text: event.target.value })}
          />
          <TextAreaField
            className={styles.wideField}
            disabled={!editable}
            label="Объяснение после ответа"
            value={draft.explanation}
            onChange={(event) => setDraft({ ...draft, explanation: event.target.value })}
          />
        </div>
        {message && <p className={styles.formMessage}>{message}</p>}
        {editable && (
          <EditorActions>
            <PrimaryAction disabled={pending} onClick={save}>
              Сохранить вопрос
            </PrimaryAction>
            {onCancel && <button onClick={onCancel}>Отмена</button>}
            {question.id && (
              <DangerAction disabled={pending} onClick={remove}>
                Удалить вопрос
              </DangerAction>
            )}
          </EditorActions>
        )}

        {questionId && (
          <div className={styles.optionList}>
            {question.options.map((option) => (
              <QuizOptionRow
                key={option.id}
                editable={editable}
                option={option}
                questionId={questionId}
                topicId={topicId}
              />
            ))}
            {addingOption && (
              <QuizOptionRow
                editable
                option={emptyOption(question.options.length + 1)}
                questionId={questionId}
                topicId={topicId}
                onCancel={() => setAddingOption(false)}
                onCreated={() => setAddingOption(false)}
              />
            )}
            {editable && question.options.length < 4 && !addingOption && (
              <button
                className={styles.addNested}
                type="button"
                onClick={() => setAddingOption(true)}
              >
                + Добавить вариант
              </button>
            )}
          </div>
        )}
      </div>
    </details>
  )
}

interface QuizOptionRowProps {
  topicId: number
  questionId: number
  option: AdminQuizOption
  editable: boolean
  onCancel?: () => void
  onCreated?: () => void
}

function QuizOptionRow({
  topicId,
  questionId,
  option,
  editable,
  onCancel,
  onCreated,
}: QuizOptionRowProps) {
  const [draft, setDraft] = useState(option)
  const [message, setMessage] = useState<string>()
  const [createOption, createState] = useCreateAdminQuizOptionMutation()
  const [updateOption, updateState] = useUpdateAdminQuizOptionMutation()
  const [deleteOption, deleteState] = useDeleteAdminQuizOptionMutation()

  useEffect(() => setDraft(option), [option])

  const save = async () => {
    try {
      if (option.id) {
        await updateOption({ topicId, questionId, optionId: option.id, value: draft }).unwrap()
        setMessage('Сохранено.')
      } else {
        await createOption({ topicId, questionId, value: draft }).unwrap()
        onCreated?.()
      }
    } catch (requestError) {
      setMessage(getApiErrorMessage(requestError))
    }
  }

  const remove = async () => {
    if (!option.id) return
    try {
      await deleteOption({ topicId, questionId, optionId: option.id }).unwrap()
    } catch (requestError) {
      setMessage(getApiErrorMessage(requestError))
    }
  }

  const pending = createState.isLoading || updateState.isLoading || deleteState.isLoading

  return (
    <div className={styles.optionRow}>
      <input
        aria-label="Порядок варианта"
        disabled={!editable}
        max={4}
        min={1}
        type="number"
        value={draft.sortOrder}
        onChange={(event) => setDraft({ ...draft, sortOrder: Number(event.target.value) })}
      />
      <input
        aria-label="Текст варианта"
        disabled={!editable}
        placeholder="Текст варианта ответа"
        value={draft.text}
        onChange={(event) => setDraft({ ...draft, text: event.target.value })}
      />
      <label className={styles.checkbox}>
        <input
          checked={draft.isCorrect}
          disabled={!editable}
          type="checkbox"
          onChange={(event) => setDraft({ ...draft, isCorrect: event.target.checked })}
        />
        Верный
      </label>
      {editable && (
        <>
          <button disabled={pending} type="button" onClick={save}>
            Сохранить
          </button>
          {onCancel ? (
            <button type="button" onClick={onCancel}>
              Отмена
            </button>
          ) : (
            <button
              className={styles.inlineDanger}
              disabled={pending}
              type="button"
              onClick={remove}
            >
              Удалить
            </button>
          )}
        </>
      )}
      {message && <small>{message}</small>}
    </div>
  )
}
