import { useEffect, useState } from 'react'
import { useAnswerDailyTaskMutation, type DailyTask } from '@/entities/progress'
import { getApiErrorCode, getApiErrorMessage } from '@/shared/http-error'
import { uiStyles } from '@/shared/ui-kit'
import styles from './DailyTaskModal.module.scss'

interface DailyTaskModalProps {
  task: DailyTask
  isOpen: boolean
  onClose: () => void
  onConflict: () => Promise<unknown>
}

export function DailyTaskModal({ task, isOpen, onClose, onConflict }: DailyTaskModalProps) {
  const [answerDailyTask, answerState] = useAnswerDailyTaskMutation()
  const [currentTask, setCurrentTask] = useState(task)
  const [error, setError] = useState('')

  useEffect(() => {
    setCurrentTask(task)
    setError('')
  }, [task])

  if (!isOpen) return null

  const submit = async (answer: boolean) => {
    setError('')

    try {
      const result = await answerDailyTask(answer).unwrap()
      setCurrentTask(result.dailyTask)
    } catch (reason) {
      if (getApiErrorCode(reason) === 'STATE_CONFLICT') {
        await onConflict()
        setError('Задание уже было выполнено. Показан сохранённый разбор.')
        return
      }
      setError(getApiErrorMessage(reason))
    }
  }

  const verdictText = currentTask.verdict ? 'Это обман' : 'Ситуация выглядит безопасной'

  return (
    <div className={styles.backdrop} role="presentation" onMouseDown={onClose}>
      <section
        aria-labelledby="daily-task-title"
        aria-modal="true"
        className={styles.dialog}
        onMouseDown={(event) => event.stopPropagation()}
        role="dialog"
      >
        <header className={styles.header}>
          <div>
            <p className={uiStyles.eyebrow}>Ежедневное задание</p>
            <h2 id="daily-task-title">Мошенник или нет?</h2>
          </div>
          <button
            aria-label="Закрыть задание"
            className={styles.close}
            onClick={onClose}
            type="button"
          >
            ×
          </button>
        </header>

        <p className={styles.intro}>
          Прочитайте Снимок диалога и решите, есть ли в ситуации признаки обмана.
        </p>
        <div aria-label="Снимок диалога" className={styles.messages}>
          {currentTask.messages.map((message, index) => (
            <p
              key={`${message.role}-${index}`}
              className={`${styles.message} ${message.role === 'assistant' ? styles.other : styles.mine}`}
            >
              {message.text}
            </p>
          ))}
        </div>

        {currentTask.isCompleted ? (
          <section className={styles.review} aria-live="polite">
            <span className={styles.completed}>Выполнено</span>
            <h3>{verdictText}</h3>
            <p>
              Ваш ответ: <b>{currentTask.answer ? 'обман' : 'не обман'}</b>
              {currentTask.isCorrect !== undefined && (
                <> · {currentTask.isCorrect ? 'верно' : 'стоит быть внимательнее'}</>
              )}
            </p>
            {currentTask.signals.length > 0 && (
              <div>
                <h4>Признаки ситуации</h4>
                <ul>
                  {currentTask.signals.map((signal) => (
                    <li key={signal}>{signal}</li>
                  ))}
                </ul>
              </div>
            )}
            {currentTask.safeAction && (
              <p className={styles.safeAction}>
                <b>Безопасное действие:</b> {currentTask.safeAction}
              </p>
            )}
          </section>
        ) : (
          <div className={styles.answers}>
            <p>Как вы оцениваете эту ситуацию?</p>
            <div className={styles.answerButtons}>
              <button
                className={styles.scamButton}
                disabled={answerState.isLoading}
                onClick={() => void submit(true)}
                type="button"
              >
                Это обман
              </button>
              <button
                className={styles.safeButton}
                disabled={answerState.isLoading}
                onClick={() => void submit(false)}
                type="button"
              >
                Ситуация безопасна
              </button>
            </div>
          </div>
        )}
        {error && (
          <p className={uiStyles.formError} role="alert">
            {error}
          </p>
        )}
      </section>
    </div>
  )
}
