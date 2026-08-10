import type { CSSProperties } from 'react'
import type { Topic } from '../model/types'
import styles from './TopicCompletionRing.module.scss'

export function TopicCompletionRing({ topic }: { topic: Topic }) {
  const total = topic.levels.length
  const completed = topic.levels.filter((level) => level.stars > 0).length
  const percentage = total ? Math.round((completed / total) * 100) : 0

  return (
    <span
      className={styles.progress}
      style={{ '--topic-progress': `${percentage * 3.6}deg` } as CSSProperties}
      role="img"
      aria-label={`Пройдено уровней: ${completed} из ${total}`}
    >
      <span>
        <b>{completed}</b>
        <small>/{total}</small>
      </span>
    </span>
  )
}
