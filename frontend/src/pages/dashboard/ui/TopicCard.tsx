import { Link } from 'react-router-dom'
import { getCompletedLevelCount, TopicCompletionRing, type Topic } from '@/entities/learning'
import styles from './Dashboard.module.scss'

export function TopicCard({ topic, basePath }: { topic: Topic; basePath: string }) {
  const completedLevels = getCompletedLevelCount(topic)

  return (
    <Link className={styles.topicCard} to={`${basePath}/lessons/${topic.id}`}>
      <div className={styles.topicCardHeader}>
        <TopicCompletionRing topic={topic} />
        <small>Тема {String(topic.order).padStart(2, '0')}</small>
      </div>
      <div className={styles.topicCardBody}>
        <h3>{topic.title}</h3>
        <p>{topic.description}</p>
        <span className={styles.topicProgressLabel}>
          {completedLevels} из {topic.levels.length} Уровней
        </span>
      </div>
    </Link>
  )
}
