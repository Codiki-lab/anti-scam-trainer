import { Link } from 'react-router-dom'
import type { Topic } from '@/entities/learning'
import { Stars } from '@/shared/stars'
import { uiStyles } from '@/shared/ui-kit'
import styles from './Dashboard.module.scss'

export function TrainingRow({ topic, basePath }: { topic: Topic; basePath: string }) {
  const stars = Math.max(...topic.levels.map((level) => level.stars), 0)

  return (
    <div className={styles.trainingRow}>
      <div>
        <small>Практика по теме</small>
        <h2>{topic.title}</h2>
        <p>{topic.description}</p>
      </div>
      <div className={styles.trainingAction}>
        <Stars value={stars} />
        <Link className={uiStyles.primaryButton} to={`${basePath}/chats`}>
          {stars ? 'Продолжить' : 'Начать'}
        </Link>
      </div>
    </div>
  )
}
