import { LockKey } from '@phosphor-icons/react'
import type { LevelState } from '@/entities/training'
import type { Topic } from '@/entities/learning'
import { Stars } from '@/shared/stars'
import { uiStyles } from '@/shared/ui-kit'
import styles from './TrainingList.module.scss'

interface TrainingListProps {
  topic: Topic
  levels: LevelState[]
  isStarting: boolean
  onStart: (level: number) => void
}

export function TrainingList({ topic, levels, isStarting, onStart }: TrainingListProps) {
  return (
    <div className={styles.list}>
      {topic.levels.map((progress) => {
        const level = levels.find((item) => item.number === progress.number)
        const isOpened = level?.isOpened ?? progress.isOpened

        return (
          <div className={`${styles.row} ${isOpened ? '' : styles.locked}`} key={progress.number}>
            <div>
              <small>Уровень {progress.number}</small>
              <h2>{topic.title}</h2>
              <p>{topic.description}</p>
              <span className={`${uiStyles.tag} ${isOpened ? '' : styles.lockedTag}`}>
                {!isOpened && <LockKey aria-hidden="true" size={15} weight="bold" />}
                {isOpened ? 'Доступен' : 'Откроется после прошлого уровня'}
              </span>
            </div>
            <div className={styles.action}>
              <Stars value={progress.stars} />
              <button
                className={`${uiStyles.primaryButton} ${isOpened ? '' : styles.lockedButton}`}
                disabled={isStarting || !isOpened}
                type="button"
                onClick={() => onStart(progress.number)}
              >
                {!isOpened && <LockKey aria-hidden="true" size={18} weight="bold" />}
                {isStarting ? 'Запускаем…' : isOpened ? 'Начать' : 'Закрыто'}
              </button>
            </div>
          </div>
        )
      })}
    </div>
  )
}
