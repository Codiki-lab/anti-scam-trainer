import { useState } from 'react'
import type { MicroQuestion } from '@/entities/training'
import { uiStyles } from '@/shared/ui-kit'
import { useAnswerMicroQuestion } from '../model/useAnswerMicroQuestion'
import styles from './MicroQuestionCard.module.scss'

interface MicroQuestionCardProps {
  attemptId: number
  question: MicroQuestion
  isPreview?: boolean
  onConflict: () => void
}

export function MicroQuestionCard({
  attemptId,
  question,
  isPreview = false,
  onConflict,
}: MicroQuestionCardProps) {
  const [previewAnswer, setPreviewAnswer] = useState<{ isCorrect: boolean; safeAction: string }>()
  const request = useAnswerMicroQuestion(attemptId, onConflict)
  const answer = isPreview ? previewAnswer : request.answer

  if (answer) {
    return (
      <section className={styles.card} aria-live="polite">
        <p className={uiStyles.eyebrow}>Короткая проверка</p>
        <h2>{answer.isCorrect ? 'Верно' : 'Обратите внимание'}</h2>
        <p>{answer.safeAction}</p>
      </section>
    )
  }

  return (
    <section className={styles.card}>
      <p className={uiStyles.eyebrow}>Закрепите правило</p>
      <h2>{question.question}</h2>
      <div className={styles.options}>
        {question.options.map((option, index) => (
          <button
            key={option}
            disabled={request.isSubmitting}
            type="button"
            onClick={() => {
              if (isPreview) {
                setPreviewAnswer({
                  isCorrect: index === 1,
                  safeAction: 'Проверяйте сделку и оплату только внутри сервиса.',
                })
                return
              }
              void request.submit(index as 0 | 1)
            }}
          >
            {option}
          </button>
        ))}
      </div>
      {request.error && (
        <p className={uiStyles.formError} role="alert">
          {request.error}
        </p>
      )}
      <p className={styles.note}>Ответ не влияет на Баллы и доступность кнопки продолжения.</p>
    </section>
  )
}
