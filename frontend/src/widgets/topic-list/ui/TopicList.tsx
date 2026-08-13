import { ArrowRight, BookOpenText, CheckCircle, PlayCircle } from '@phosphor-icons/react'
import { Link } from 'react-router-dom'
import type { Topic } from '@/entities/learning'
import styles from './TopicList.module.scss'

function getTopicState(topic: Topic) {
  if (topic.isCompleted) {
    return {
      label: 'Тема пройдена',
      className: styles.completed,
      icon: <CheckCircle aria-hidden="true" weight="fill" />,
    }
  }

  if (topic.isQuizPassed) {
    return {
      label: 'Доступны Уровни',
      className: styles.training,
      icon: <PlayCircle aria-hidden="true" weight="fill" />,
    }
  }

  if (topic.isTheoryRead) {
    return {
      label: 'Пройдите квиз',
      className: styles.quiz,
      icon: <BookOpenText aria-hidden="true" weight="fill" />,
    }
  }

  return {
    label: 'Теория и квиз',
    className: styles.new,
    icon: <BookOpenText aria-hidden="true" weight="fill" />,
  }
}

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
        const state = getTopicState(topic)
        const completedLevels = topic.levels.filter((level) => level.stars > 0).length
        const completedSteps =
          Number(topic.isTheoryRead) + Number(topic.isQuizPassed) + completedLevels
        const totalSteps = 2 + topic.levels.length
        const progress = totalSteps > 0 ? (completedSteps / totalSteps) * 100 : 0

        return (
          <Link
            key={topic.id}
            className={`${styles.card} ${state.className}`}
            to={`${basePath}/${topic.id}`}
          >
            <div className={styles.topline}>
              <span className={styles.number}>{String(topic.order).padStart(2, '0')}</span>
              <span className={styles.arrow} aria-hidden="true">
                <ArrowRight weight="bold" />
              </span>
            </div>

            <div className={styles.content}>
              <h2>{topic.title}</h2>
              <p>{topic.description}</p>
            </div>

            <div className={styles.footer}>
              <span className={styles.status}>
                {state.icon}
                {state.label}
              </span>
              <span className={styles.progressLabel}>
                {completedSteps} из {totalSteps}
              </span>
            </div>

            <span
              className={styles.progress}
              role="progressbar"
              aria-label={`Прогресс Темы «${topic.title}»`}
              aria-valuenow={completedSteps}
              aria-valuemin={0}
              aria-valuemax={totalSteps}
            >
              <span style={{ width: `${progress}%` }} />
            </span>
          </Link>
        )
      })}
    </div>
  )
}
