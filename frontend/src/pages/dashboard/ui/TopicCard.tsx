import { Link } from 'react-router-dom'
import type { Topic } from '@/entities/learning'
import { uiStyles } from '@/shared/ui-kit'
import styles from './Dashboard.module.scss'

export function TopicCard({ topic, basePath }: { topic: Topic; basePath: string }) {
  return (
    <Link className={styles.topicCard} to={`${basePath}/lessons/${topic.id}`}>
      <span className={styles.topicIcon}>{String(topic.order).padStart(2, '0')}</span>
      <small>Тема {topic.order}</small>
      <h3>{topic.title}</h3>
      <p>{topic.description}</p>
      <span className={uiStyles.tag}>{topic.isCompleted ? 'Пройдено' : 'В процессе'}</span>
    </Link>
  )
}
