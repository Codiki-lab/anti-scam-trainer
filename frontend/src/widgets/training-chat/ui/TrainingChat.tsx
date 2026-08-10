import { useEffect, useRef, useState } from 'react'
import type { TrainingAnswer, TrainingSession } from '@/entities/training'
import { ConfirmDialog, uiStyles } from '@/shared/ui-kit'
import styles from './TrainingChat.module.scss'

interface TrainingChatProps {
  session: TrainingSession
  isSubmitting: boolean
  error: string
  cooldown: number
  onSubmit: (answer: TrainingAnswer) => Promise<boolean>
  onAbandon: () => Promise<void>
}

const dealMethodLabels = {
  delivery: 'Avito Доставка',
  meetup: 'Личная встреча',
  pickup: 'Самовывоз',
}

const productImageKeys = new Set([
  'smartphone',
  'electronics',
  'appliance',
  'camera',
  'bicycle',
  'laptop',
  'headphones',
  'console',
])

export function TrainingChat({
  session,
  isSubmitting,
  error,
  cooldown,
  onSubmit,
  onAbandon,
}: TrainingChatProps) {
  const [freeText, setFreeText] = useState('')
  const [pendingText, setPendingText] = useState('')
  const [selectedOptionId, setSelectedOptionId] = useState<number>()
  const [finishOpen, setFinishOpen] = useState(false)
  const [abandonOpen, setAbandonOpen] = useState(false)
  const submittingText = useRef(false)
  const acceptsText = session.step.options.length === 0

  useEffect(() => {
    setSelectedOptionId(undefined)
    setFreeText('')
  }, [session.step.id])

  const submitText = async (finish = false) => {
    const text = freeText.trim()
    if (!text || isSubmitting || submittingText.current || cooldown > 0) return
    submittingText.current = true
    setPendingText(text)
    try {
      const succeeded = await onSubmit({ type: 'text', stepId: session.step.id, text, finish })
      if (succeeded) setFreeText('')
    } finally {
      submittingText.current = false
      setPendingText('')
      setFinishOpen(false)
    }
  }

  const selectedOption = session.step.options.find((option) => option.id === selectedOptionId)

  let answerIndex = 0
  const dialogueHistory = session.messages.map((message) => {
    if (message.role !== 'user') return message
    const answer = session.answers[answerIndex++]
    return answer
      ? { role: 'user' as const, text: answer.optionText || answer.freeText || 'Ответ сохранён' }
      : message
  })
  if (pendingText) dialogueHistory.push({ role: 'user', text: pendingText })
  const productImageKey = productImageKeys.has(session.productContext.imageKey ?? '')
    ? session.productContext.imageKey
    : 'electronics'
  const price = session.productContext.price
    ? new Intl.NumberFormat('ru-RU').format(session.productContext.price) + ' ₽'
    : 'Цена не указана'

  return (
    <section className={styles.layout}>
      <div className={styles.main}>
        <div className={styles.header}>
          <div>
            <small>
              {session.topicTitle} · {session.level ? `Уровень ${session.level}` : 'Свободная игра'}
            </small>
            <h2>{session.scenarioTitle}</h2>
            <p>{session.scenarioDescription}</p>
          </div>
          <div className={styles.headerActions}>
            <span className={uiStyles.tag}>
              Шаг {session.progress.currentStep} из {session.progress.totalSteps}
            </span>
            <button
              type="button"
              className={styles.abandonButton}
              onClick={() => setAbandonOpen(true)}
            >
              Выйти
            </button>
          </div>
        </div>
        <div className={styles.compactProduct} aria-label="Контекст товара">
          <b>{session.productContext.itemTitle}</b>
          <span>{price}</span>
          <small>{dealMethodLabels[session.productContext.dealMethod]}</small>
        </div>
        <div className={styles.messages}>
          {dialogueHistory.map((message, index) => (
            <p
              key={`${message.role}-${index}`}
              className={`${styles.bubble} ${message.role === 'assistant' ? styles.other : styles.mine}`}
            >
              {message.text}
            </p>
          ))}
        </div>
        {session.step.options.length > 0 && (
          <div className={styles.replyOptions} aria-label="Готовые ответы пользователя">
            <p>Выберите, что ответить</p>
            {session.step.options.map((option) => (
              <button
                key={option.id}
                disabled={isSubmitting}
                type="button"
                className={selectedOptionId === option.id ? styles.selectedOption : undefined}
                aria-pressed={selectedOptionId === option.id}
                onClick={() => setSelectedOptionId(option.id)}
              >
                {option.text}
              </button>
            ))}
            <button
              className={`${uiStyles.primaryButton} ${styles.confirmReplyButton}`}
              disabled={!selectedOption || isSubmitting}
              type="button"
              onClick={() => {
                if (!selectedOption) return
                void onSubmit({
                  type: 'option',
                  stepId: session.step.id,
                  optionId: selectedOption.id,
                })
              }}
            >
              {isSubmitting ? 'Отправляем…' : 'Отправить реплику'}
            </button>
          </div>
        )}
        {acceptsText && (
          <div className={styles.composer}>
            <textarea
              maxLength={400}
              aria-label="Ответ пользователя"
              placeholder="Напишите ответ собеседнику…"
              value={freeText}
              onChange={(event) => setFreeText(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter' && !event.shiftKey) {
                  event.preventDefault()
                  void submitText(false)
                }
              }}
            />
            <button
              className={`${uiStyles.primaryButton} ${styles.sendButton}`}
              disabled={isSubmitting || cooldown > 0 || !freeText.trim()}
              type="button"
              onClick={() => void submitText(false)}
            >
              {isSubmitting ? 'Отправляем…' : cooldown > 0 ? `Через ${cooldown} сек.` : 'Отправить'}
            </button>
            {session.canFinishEarly && (
              <button
                className={uiStyles.secondaryButton}
                type="button"
                onClick={() => setFinishOpen(true)}
                disabled={isSubmitting || cooldown > 0 || !freeText.trim()}
              >
                Завершить этим ответом
              </button>
            )}
          </div>
        )}
        {isSubmitting && (
          <p className={styles.typing} role="status">
            Собеседник печатает…
          </p>
        )}
        {error && <p className={uiStyles.formError}>{error}</p>}
      </div>
      <aside className={styles.side} aria-label="Контекст товара и сделки">
        <p className={uiStyles.eyebrow}>Контекст сделки</p>
        <div className={styles.productImage} aria-hidden="true">
          <svg viewBox="0 0 160 120" role="presentation">
            <use href={`/assets/product-illustrations.svg#${productImageKey}`} />
          </svg>
        </div>
        <h3>{session.productContext.itemTitle}</h3>
        <strong className={styles.price}>{price}</strong>
        <dl>
          <div>
            <dt>Категория</dt>
            <dd>{session.productContext.category}</dd>
          </div>
          <div>
            <dt>Сделка</dt>
            <dd>{dealMethodLabels[session.productContext.dealMethod]}</dd>
          </div>
          {session.productContext.location && (
            <div>
              <dt>Локация</dt>
              <dd>{session.productContext.location}</dd>
            </div>
          )}
          <div>
            <dt>Ваша роль</dt>
            <dd>{session.userRole === 'buyer' ? 'Покупатель' : 'Продавец'}</dd>
          </div>
          <div>
            <dt>Собеседник</dt>
            <dd>{session.counterpartyRole === 'buyer' ? 'Покупатель' : 'Продавец'}</dd>
          </div>
        </dl>
        <div className={styles.riskNote}>
          <b>Помните</b>
          <br />
          Не сообщайте секретные данные и не уходите на сторонние страницы.
        </div>
      </aside>
      <ConfirmDialog
        open={finishOpen}
        title="Завершить Прохождение?"
        description="Текущий текст станет последним Ответом пользователя и перейдёт в Result."
        confirmLabel="Завершить"
        isPending={isSubmitting}
        onClose={() => setFinishOpen(false)}
        onConfirm={() => void submitText(true)}
      />
      <ConfirmDialog
        open={abandonOpen}
        title="Отказаться от Прохождения?"
        description="Незавершённое Прохождение будет закрыто без Result и не изменит Прогресс."
        confirmLabel="Отказаться"
        isPending={isSubmitting}
        onClose={() => setAbandonOpen(false)}
        onConfirm={() => void onAbandon()}
      />
    </section>
  )
}
