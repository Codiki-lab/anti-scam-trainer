import { useState } from 'react'
import type { Quiz, QuizOutcome, QuizSubmission } from '@/entities/learning'
import { uiStyles } from '@/shared/ui-kit'
import styles from './QuizPanel.module.scss'

interface QuizPanelProps {
  quiz: Quiz
  isSubmitting: boolean
  onSubmit: (submission: QuizSubmission) => Promise<QuizOutcome>
  onPassed: () => void
}

export function QuizPanel({ quiz, isSubmitting, onSubmit, onPassed }: QuizPanelProps) {
  const [questionIndex, setQuestionIndex] = useState(0)
  const [selectedChoice, setSelectedChoice] = useState<number | null>(null)
  const [answers, setAnswers] = useState<QuizSubmission['answers']>([])
  const [outcome, setOutcome] = useState<QuizOutcome>()
  const [error, setError] = useState('')
  const question = quiz.questions[questionIndex]

  const restart = () => {
    setQuestionIndex(0)
    setSelectedChoice(null)
    setAnswers([])
    setOutcome(undefined)
    setError('')
  }

  const proceed = async () => {
    if (selectedChoice === null) return
    const nextAnswers = [...answers, { questionId: question.id, choiceId: selectedChoice }]

    if (questionIndex === quiz.questions.length - 1) {
      setError('')
      try {
        setOutcome(await onSubmit({ answers: nextAnswers }))
      } catch {
        setError('Не удалось отправить ответы. Попробуйте ещё раз.')
      }
    } else {
      setAnswers(nextAnswers)
      setQuestionIndex((index) => index + 1)
      setSelectedChoice(null)
    }
  }

  if (outcome) {
    return (
      <section className={styles.quiz}>
        <p className={uiStyles.eyebrow}>Результат квиза</p>
        <h1>{outcome.score}%</h1>
        <p>
          {outcome.isPassed
            ? 'Проверка пройдена — практика открыта.'
            : `Нужно набрать ${quiz.passThreshold}%. Повторите Теорию и попробуйте снова.`}
        </p>
        <div className={uiStyles.buttonRow}>
          {outcome.isPassed && (
            <button className={uiStyles.primaryButton} type="button" onClick={onPassed}>
              К тренировкам
            </button>
          )}
          {!outcome.isPassed && (
            <button className={uiStyles.primaryButton} type="button" onClick={restart}>
              Пройти ещё раз
            </button>
          )}
        </div>
      </section>
    )
  }

  return (
    <section className={styles.quiz}>
      <p className={uiStyles.eyebrow}>
        Проверка знаний · {questionIndex + 1} из {quiz.questions.length}
      </p>
      <div className={styles.progress}>
        <i style={{ width: `${((questionIndex + 1) / quiz.questions.length) * 100}%` }} />
      </div>
      <h1>{question.text}</h1>
      <div className={styles.answerList}>
        {question.choices.map((choice, index) => (
          <button
            key={choice.id}
            className={selectedChoice === choice.id ? styles.selected : undefined}
            type="button"
            onClick={() => setSelectedChoice(choice.id)}
          >
            <span>{String.fromCharCode(65 + index)}</span>
            {choice.text}
          </button>
        ))}
      </div>
      {error && <p className={uiStyles.formError}>{error}</p>}
      <button
        className={uiStyles.primaryButton}
        disabled={selectedChoice === null || isSubmitting}
        type="button"
        onClick={() => void proceed()}
      >
        {isSubmitting
          ? 'Отправляем…'
          : questionIndex === quiz.questions.length - 1
            ? 'Завершить'
            : 'Ответить'}
      </button>
    </section>
  )
}
