import { useState } from 'react'
import type { TrainingAnswer, TrainingSession } from '@/entities/training'
import { uiStyles } from '@/shared/ui-kit'
import styles from './TrainingChat.module.scss'

interface TrainingChatProps {
  session: TrainingSession
  isSubmitting: boolean
  error: string
  onSubmit: (answer: TrainingAnswer) => Promise<boolean>
}

export function TrainingChat({ session, isSubmitting, error, onSubmit }: TrainingChatProps) {
  const [freeText, setFreeText] = useState('')
  const acceptsText = session.mode === 'mixed' || session.mode === 'free_text'

  const submitText = async () => {
    const text = freeText.trim()
    if (!text) return
    const wasAccepted = await onSubmit({ type: 'text', stepId: session.step.id, text })
    if (wasAccepted) setFreeText('')
  }

  return (
    <section className={styles.layout}>
      <div className={styles.main}>
        <div className={styles.header}>
          <div>
            <small>Прохождение · Шаг {session.step.number}</small>
            <h2>Безопасная сделка</h2>
          </div>
          <span className={uiStyles.tag}>В процессе</span>
        </div>
        <div className={styles.messages}>
          {session.messages.map((message, index) => (
            <p
              key={`${message.role}-${index}`}
              className={`${styles.bubble} ${message.role === 'assistant' ? styles.other : styles.mine}`}
            >
              {message.text}
            </p>
          ))}
        </div>
        {session.step.options.length > 0 && (
          <div className={styles.options}>
            {session.step.options.map((option) => (
              <button
                key={option.id}
                disabled={isSubmitting}
                type="button"
                onClick={() =>
                  void onSubmit({ type: 'option', stepId: session.step.id, optionId: option.id })
                }
              >
                {option.text}
              </button>
            ))}
          </div>
        )}
        {acceptsText && (
          <div className={styles.options}>
            <textarea
              maxLength={400}
              placeholder="Напишите безопасный ответ…"
              value={freeText}
              onChange={(event) => setFreeText(event.target.value)}
            />
            <button
              className={`${uiStyles.primaryButton} ${styles.sendButton}`}
              disabled={isSubmitting || !freeText.trim()}
              type="button"
              onClick={() => void submitText()}
            >
              Отправить
            </button>
          </div>
        )}
        {error && <p className={uiStyles.formError}>{error}</p>}
      </div>
      <aside className={styles.side}>
        <p className={uiStyles.eyebrow}>Контекст сделки</p>
        <h3>Сохраните сделку безопасной</h3>
        <p>{session.step.counterpartyMessage}</p>
        <div className={styles.riskNote}>
          <b>Помните</b>
          <br />
          Не сообщайте секретные данные и не уходите на сторонние страницы.
        </div>
      </aside>
    </section>
  )
}
