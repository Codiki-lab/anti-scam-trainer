import { Link } from 'react-router-dom'
import { useGetRecommendationQuery } from '@/entities/learning'
import type { UserRole } from '@/entities/user'
import { ErrorState } from '@/shared/error-state'
import { getLearningActionPath } from '../lib/getLearningActionPath'
import styles from './LearningRecommendationCard.module.scss'

interface LearningRecommendationCardProps {
  role: UserRole
  basePath: string
  disabled?: boolean
}

export function LearningRecommendationCard({
  role,
  basePath,
  disabled = false,
}: LearningRecommendationCardProps) {
  const query = useGetRecommendationQuery(role, { skip: disabled })

  if (disabled || query.isLoading) return null
  if (query.error) {
    return (
      <ErrorState
        message="Не удалось подобрать следующий шаг."
        onRetry={() => void query.refetch()}
      />
    )
  }
  if (!query.data) return null

  const recommendation = query.data
  return (
    <aside className={styles.card}>
      <span>{recommendation.isFallback ? 'Безопасный старт' : 'Рекомендуем дальше'}</span>
      <h3>{recommendation.topic.title}</h3>
      <p>{recommendation.explanation}</p>
      <Link to={getLearningActionPath(recommendation.nextAction, basePath)}>Продолжить →</Link>
    </aside>
  )
}
