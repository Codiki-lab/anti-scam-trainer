import {
  ArrowRight,
  ChatCircleText,
  CheckCircle,
  ListChecks,
  LockKey,
  PencilSimpleLine,
} from '@phosphor-icons/react'
import type { LevelState } from '@/entities/training'
import type { Topic } from '@/entities/learning'
import { Stars } from '@/shared/stars'
import styles from './TrainingList.module.scss'

interface TrainingListProps {
  topic: Topic
  levels: LevelState[]
  isStarting: boolean
  onStart: (level: number) => void
}

const responseModePresentation = {
  multiple_choice: { label: 'Готовые варианты', icon: ListChecks },
  similar_choice: { label: 'Похожие варианты', icon: ListChecks },
  mixed: { label: 'Смешанный ответ', icon: PencilSimpleLine },
  free_text: { label: 'Свободный текст', icon: ChatCircleText },
} as const

export function TrainingList({ topic, levels, isStarting, onStart }: TrainingListProps) {
  return (
    <div className={styles.list}>
      {topic.levels.map((progress) => {
        const level = levels.find((item) => item.number === progress.number)
        const isOpened = level?.isOpened ?? progress.isOpened
        const mode = level ? responseModePresentation[level.responseMode] : undefined
        const ModeIcon = mode?.icon
        const isInProgress = Boolean(level?.inProgressAttemptId)
        const isCompleted = progress.stars > 0
        const actionLabel = isInProgress ? 'Продолжить' : isCompleted ? 'Пройти ещё раз' : 'Начать'

        return (
          <article
            className={`${styles.card} ${isOpened ? '' : styles.locked}`}
            key={progress.number}
          >
            <div className={styles.cardHeader}>
              <span className={styles.levelNumber}>{String(progress.number).padStart(2, '0')}</span>
              {mode && ModeIcon && (
                <span className={styles.mode}>
                  <ModeIcon aria-hidden="true" weight="bold" />
                  {mode.label}
                </span>
              )}
            </div>

            <div className={styles.content}>
              <small>
                Уровень {progress.number} · {topic.title}
              </small>
              <h2>{level?.scenarioTitle || topic.title}</h2>
              <p>{level?.scenarioDescription || topic.description}</p>
            </div>

            <div className={styles.result}>
              <Stars value={progress.stars} />
              {isCompleted && <span>{progress.bestScore} Баллов</span>}
            </div>

            <div className={styles.footer}>
              <span className={`${styles.status} ${isCompleted ? styles.completedStatus : ''}`}>
                {!isOpened ? (
                  <LockKey aria-hidden="true" weight="bold" />
                ) : isCompleted ? (
                  <CheckCircle aria-hidden="true" weight="fill" />
                ) : (
                  <ChatCircleText aria-hidden="true" weight="fill" />
                )}
                {!isOpened
                  ? `Откроется после Уровня ${Math.max(1, progress.number - 1)}`
                  : isInProgress
                    ? 'Прохождение не завершено'
                    : isCompleted
                      ? 'Уровень пройден'
                      : 'Готов к прохождению'}
              </span>
              <button
                className={`${styles.action} ${isOpened ? '' : styles.lockedButton}`}
                disabled={isStarting || !isOpened}
                type="button"
                onClick={() => onStart(progress.number)}
              >
                {!isOpened && <LockKey aria-hidden="true" size={18} weight="bold" />}
                {isStarting ? 'Запускаем…' : isOpened ? actionLabel : 'Закрыто'}
                {isOpened && <ArrowRight aria-hidden="true" weight="bold" />}
              </button>
            </div>
          </article>
        )
      })}
    </div>
  )
}
