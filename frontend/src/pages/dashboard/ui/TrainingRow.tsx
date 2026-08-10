import { Link } from 'react-router-dom'
import { getCompletedLevelCount, type Topic } from '@/entities/learning'
import { uiStyles } from '@/shared/ui-kit'
import styles from './Dashboard.module.scss'

export function TrainingRow({ topic, basePath }: { topic: Topic; basePath: string }) {
  const completedLevels = getCompletedLevelCount(topic)

  return (
    <div className={styles.trainingRow}>
      <div>
        <small>Практика по теме</small>
        <h2>{topic.title}</h2>
        <p>{topic.description}</p>
      </div>
      <div className={styles.trainingAction}>
        <span
          className={styles.levelDots}
          aria-label={`Пройдено ${completedLevels} из ${topic.levels.length} уровней`}
        >
          {topic.levels.map((level) => (
            <i
              key={level.number}
              aria-hidden="true"
              className={level.stars > 0 ? styles.completedLevelDot : styles.levelDot}
            />
          ))}
        </span>
        <Link className={uiStyles.primaryButton} to={`${basePath}/chats`}>
          {completedLevels ? 'Продолжить' : 'Начать'}
        </Link>
      </div>
    </div>
  )
}
