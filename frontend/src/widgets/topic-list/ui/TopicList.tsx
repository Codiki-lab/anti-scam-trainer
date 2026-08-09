import { Link } from 'react-router-dom'
import type { Topic } from '@/entities/learning'
import { uiStyles } from '@/shared/ui-kit'
import styles from './TopicList.module.scss'

export function TopicList({
  topics,
  basePath = '/lessons',
}: {
  topics: Topic[]
  basePath?: string
}) {
  return (
    <div className={styles.grid}>
      {topics.map((topic) => {
        const status = topic.isCompleted
          ? 'Тема пройдена'
          : topic.isQuizPassed
            ? 'Доступны тренировки'
            : topic.isTheoryRead
              ? 'Пройдите квиз'
              : 'Теория и квиз'

        return (
          <Link key={topic.id} className={styles.card} to={`${basePath}/${topic.id}`}>
            <span className={styles.number}>{String(topic.order).padStart(2, '0')}</span>
            <div>
              <h2>{topic.title}</h2>
              <p>{topic.description}</p>
              <span className={uiStyles.tag}>{status}</span>
            </div>
            <b aria-hidden="true">→</b>
          </Link>
        )
      })}
    </div>
  )
}
