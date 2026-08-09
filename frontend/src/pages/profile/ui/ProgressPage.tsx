import type { Progress } from '@/entities/progress'
import { useCurrentAccount } from '@/entities/user'
import { useProgressData } from '@/features/view-progress'
import { ErrorState } from '@/shared/error-state'
import { uiStyles } from '@/shared/ui-kit'
import { Metric } from './Metric'
import styles from './Profile.module.scss'

export function ProgressPage({ previewProgress }: { previewProgress?: Progress }) {
  const { account } = useCurrentAccount()
  const { progress, isLoading, error, retry } = useProgressData(
    account.trainingRole,
    previewProgress,
  )

  if (isLoading) return <p className={uiStyles.muted}>Загружаем прогресс…</p>
  if (error) return <ErrorState message={error} onRetry={() => void retry()} />
  if (!progress) return <p className={uiStyles.formError}>Не удалось загрузить прогресс.</p>

  const completion = Math.round(
    (progress.summary.completedTopics / Math.max(progress.summary.totalTopics, 1)) * 100,
  )

  return (
    <>
      <section className={uiStyles.pageHeading}>
        <p className={uiStyles.eyebrow}>Ваш путь</p>
        <h1>Прогресс</h1>
        <p className={uiStyles.muted}>
          Продолжайте практиковаться: короткие повторения укрепляют навык.
        </p>
      </section>

      <div className={styles.metrics}>
        <Metric value={`${progress.summary.completedLevels}`} label="уровней пройдено" />
        <Metric value={`${Math.round(progress.summary.averageScore)}%`} label="средний результат" />
        <Metric value={`${progress.summary.stars}`} label="звёзд получено" />
        <Metric value={`${progress.summary.completedTopics}`} label="тем завершено" />
      </div>

      <section className={styles.progressPanel}>
        <div>
          <h2>Пройденные темы</h2>
          <p>
            {progress.summary.completedTopics} из {progress.summary.totalTopics} тем завершены
          </p>
        </div>
        <div className={styles.wideTrack}>
          <i style={{ width: `${completion}%` }} />
        </div>
      </section>

      <section>
        <h2>Последние тренировки</h2>
        <div className={styles.history}>
          {progress.recentAttempts.map((attempt) => {
            const topic = progress.topics.find((item) => item.id === attempt.topicId)
            return (
              <p key={attempt.attemptId}>
                <span>
                  {topic?.title ?? `Тема ${attempt.topicId}`}
                  <small>Уровень {attempt.level}</small>
                </span>
                <b className={attempt.score === 100 ? styles.success : undefined}>
                  {attempt.score}%
                </b>
              </p>
            )
          })}
        </div>
      </section>
    </>
  )
}
