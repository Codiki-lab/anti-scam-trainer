import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useGetLevelsQuery } from '@/entities/training'
import { useCurrentAccount } from '@/entities/user'
import { useTopics } from '@/features/learning-content'
import { useStartTraining } from '@/features/start-training'
import { ErrorState } from '@/shared/error-state'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'
import { uiStyles } from '@/shared/ui-kit'
import { parsePositiveInteger } from '@/shared/url'
import { TrainingList } from '@/widgets/training-list'
import type { TrainingPreview } from '../model/types'
import styles from './TrainingPage.module.scss'

export function TrainingPage({ preview }: { preview?: TrainingPreview }) {
  const { account } = useCurrentAccount()
  const role = account.trainingRole
  const isPreview = useIsPreview()
  const [searchParams, setSearchParams] = useSearchParams()
  const { topics, isLoading: areTopicsLoading } = useTopics(role, preview?.topics)
  const selectedTopicId = parsePositiveInteger(searchParams.get('topic'))
  const topic = topics.find((item) => item.id === selectedTopicId) ?? topics[0]
  const topicId = topic?.id ?? 0
  const levelsQuery = useGetLevelsQuery({ role, topicId }, { skip: isPreview || topicId < 1 })
  const levels = isPreview ? (preview?.levels ?? []) : (levelsQuery.data ?? [])
  const { start, isStarting } = useStartTraining(topicId, role)
  const [error, setError] = useState('')

  const startLevel = async (level: number) => setError((await start(level)) ?? '')

  if (areTopicsLoading || levelsQuery.isLoading)
    return <p className={uiStyles.muted}>Загружаем Уровни…</p>
  if (!isPreview && levelsQuery.error) {
    return (
      <ErrorState
        message={getApiErrorMessage(levelsQuery.error)}
        onRetry={() => void levelsQuery.refetch()}
      />
    )
  }
  if (!topic) return <p className={uiStyles.formError}>Тема для тренировки не найдена.</p>

  return (
    <>
      <section className={uiStyles.pageHeading}>
        <p className={uiStyles.eyebrow}>Практика</p>
        <h1>Тренировки</h1>
        <p className={uiStyles.muted}>
          В каждой из шести тем — четыре уровня сложности. Всего 24 тренировки для каждой роли.
        </p>
      </section>
      <div className={styles.trainingToolbar}>
        <label className={styles.topicField}>
          Тема тренировки
          <select
            value={topic.id}
            onChange={(event) => setSearchParams({ topic: event.target.value })}
          >
            {topics.map((item) => (
              <option key={item.id} value={item.id}>
                {item.title}
              </option>
            ))}
          </select>
        </label>
        <p>
          Показаны 4 уровня выбранной темы. Следующий уровень открывается после получения хотя бы
          одной звезды.
        </p>
      </div>
      {error && <p className={uiStyles.formError}>{error}</p>}
      <TrainingList topic={topic} levels={levels} isStarting={isStarting} onStart={startLevel} />
    </>
  )
}
