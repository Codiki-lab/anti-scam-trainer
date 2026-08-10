import { LockKey } from '@phosphor-icons/react'
import { Link } from 'react-router-dom'
import type { LevelState } from '@/entities/training'
import type { Topic } from '@/entities/learning'
import { Stars } from '@/shared/stars'
import { StatusBadge, uiStyles } from '@/shared/ui-kit'
import styles from './TrainingList.module.scss'

interface TrainingListProps {
  topic: Topic
  levels: LevelState[]
  isStarting: boolean
  onStart: (level: number) => void
  basePath: string
}

const mechanics = {
  multiple_choice: 'Ответьте собеседнику одной из готовых реплик.',
  similar_choice: 'Сравните похожие реплики и отправьте наиболее безопасную.',
  mixed: 'Сначала отправьте готовую реплику, затем ответьте своими словами.',
  free_text: 'Ведите диалог самостоятельно без готовых формулировок.',
}

export function TrainingList({ topic, levels, isStarting, onStart, basePath }: TrainingListProps) {
  return (
    <div className={styles.list}>
      {topic.levels.map((progress) => {
        const level = levels.find((item) => item.number === progress.number)
        const isOpened = level?.isOpened ?? false

        return (
          <div className={`${styles.row} ${isOpened ? '' : styles.locked}`} key={progress.number}>
            <div>
              <small>Уровень {progress.number}</small>
              <h2>{level?.scenarioTitle ?? topic.title}</h2>
              <p>{level?.scenarioDescription ?? topic.description}</p>
              {level && <p className={styles.mechanic}>{mechanics[level.responseType]}</p>}
              <StatusBadge
                tone={isOpened ? (level?.inProgressAttemptId ? 'progress' : 'success') : 'locked'}
              >
                {!isOpened && <LockKey aria-hidden="true" size={15} weight="bold" />}
                {isOpened
                  ? 'Доступен'
                  : progress.number === 1
                    ? 'Пройдите Quiz минимум на 80%'
                    : `Получите хотя бы одну Звезду за Уровень ${progress.number - 1}`}
              </StatusBadge>
            </div>
            <div className={styles.action}>
              <Stars value={progress.stars} />
              <small>
                Лучший Балл: {progress.bestScore} · Прохождений: {progress.attempts}
              </small>
              {progress.lastAttemptId && (
                <Link to={`${basePath}/sessions/${progress.lastAttemptId}/result`}>
                  Последний Result →
                </Link>
              )}
              <button
                className={`${uiStyles.primaryButton} ${isOpened ? '' : styles.lockedButton}`}
                disabled={isStarting || !isOpened}
                type="button"
                onClick={() => onStart(progress.number)}
              >
                {!isOpened && <LockKey aria-hidden="true" size={18} weight="bold" />}
                {isStarting
                  ? 'Запускаем…'
                  : !isOpened
                    ? 'Закрыто'
                    : level?.inProgressAttemptId
                      ? 'Продолжить'
                      : progress.attempts > 0
                        ? 'Попробовать снова'
                        : 'Начать'}
              </button>
            </div>
          </div>
        )
      })}
    </div>
  )
}
