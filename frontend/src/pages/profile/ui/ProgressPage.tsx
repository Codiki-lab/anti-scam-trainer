import {
  ArrowRight,
  ChartLineUp,
  CheckCircle,
  ClockCounterClockwise,
  Star,
} from '@phosphor-icons/react'
import { Link } from 'react-router-dom'
import type { Progress } from '@/entities/progress'
import { useCurrentAccount } from '@/entities/user'
import { useProgressData } from '@/features/view-progress'
import { ErrorState } from '@/shared/error-state'
import { useIsPreview } from '@/shared/runtime-mode'
import { uiStyles } from '@/shared/ui-kit'
import { Metric } from './Metric'
import styles from './Profile.module.scss'

export function ProgressPage({ previewProgress }: { previewProgress?: Progress }) {
  const isPreview = useIsPreview()
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
  const basePath = isPreview ? '/preview' : ''
  const roleName = progress.role === 'buyer' ? 'покупателя' : 'продавца'

  return (
    <>
      <section className={`${uiStyles.pageHeading} ${styles.progressHeading}`}>
        <p className={uiStyles.eyebrow}>Ваш путь</p>
        <h1>Прогресс</h1>
        <p className={uiStyles.muted}>
          Следите за результатами в ролевой ветке {roleName} и возвращайтесь к темам, которые
          хочется закрепить.
        </p>
      </section>

      <section className={styles.progressOverview}>
        <div className={styles.overallProgress}>
          <div className={styles.progressCopy}>
            <span className={styles.progressIcon}>
              <ChartLineUp size={24} weight="duotone" aria-hidden="true" />
            </span>
            <div>
              <p>Общий прогресс</p>
              <h2>{completion}%</h2>
            </div>
          </div>
          <p>
            {progress.summary.completedTopics} из {progress.summary.totalTopics} Тем завершено
          </p>
          <div
            className={styles.wideTrack}
            role="progressbar"
            aria-label="Общий прогресс по Темам"
            aria-valuemin={0}
            aria-valuemax={progress.summary.totalTopics}
            aria-valuenow={progress.summary.completedTopics}
          >
            <i style={{ width: `${completion}%` }} />
          </div>
          <Link className={styles.continueLink} to={`${basePath}/lessons`}>
            Продолжить обучение <ArrowRight size={18} aria-hidden="true" />
          </Link>
        </div>

        <div className={styles.metrics}>
          <Metric
            value={`${progress.summary.completedLevels}/${progress.summary.totalLevels}`}
            label="Уровней пройдено"
            icon={<CheckCircle size={23} weight="duotone" />}
            tone="blue"
          />
          <Metric
            value={`${Math.round(progress.summary.averageScore)}%`}
            label="Средний Балл"
            icon={<ChartLineUp size={23} weight="duotone" />}
            tone="green"
          />
          <Metric
            value={`${progress.summary.stars}`}
            label="Звёзд получено"
            icon={<Star size={23} weight="duotone" />}
            tone="purple"
          />
        </div>
      </section>

      <section className={styles.topicProgressSection}>
        <div className={styles.profileSectionHeading}>
          <div>
            <p className={uiStyles.eyebrow}>По темам</p>
            <h2>Путь обучения</h2>
          </div>
          <span>{progress.summary.completedTopics} завершено</span>
        </div>
        <div className={styles.topicProgressGrid}>
          {progress.topics.map((topic) => {
            const completedLevels = topic.levels.filter((level) => level.stars > 0).length
            const topicCompletion = Math.round(
              (completedLevels / Math.max(topic.levels.length, 1)) * 100,
            )

            return (
              <Link
                className={styles.topicProgressCard}
                key={topic.id}
                to={`${basePath}/lessons/${topic.id}`}
              >
                <span className={styles.topicNumber}>{String(topic.order).padStart(2, '0')}</span>
                <div>
                  <h3>{topic.title}</h3>
                  <p>
                    {topic.isCompleted
                      ? 'Тема завершена'
                      : `${completedLevels} из ${topic.levels.length} Уровней пройдено`}
                  </p>
                </div>
                <span className={styles.topicPercent}>
                  {topic.isCompleted ? 'Готово' : `${topicCompletion}%`}
                </span>
                <div
                  className={styles.topicTrack}
                  role="progressbar"
                  aria-label={`Прогресс Темы «${topic.title}»`}
                  aria-valuemin={0}
                  aria-valuemax={topic.levels.length}
                  aria-valuenow={completedLevels}
                >
                  <i style={{ width: `${topicCompletion}%` }} />
                </div>
              </Link>
            )
          })}
        </div>
      </section>

      <section className={styles.recentSection}>
        <div className={styles.profileSectionHeading}>
          <div>
            <p className={uiStyles.eyebrow}>История</p>
            <h2>Последние Прохождения</h2>
          </div>
          <ClockCounterClockwise size={25} color="#00aaff" aria-hidden="true" />
        </div>
        {progress.recentAttempts.length > 0 ? (
          <div className={styles.history}>
            {progress.recentAttempts.map((attempt) => {
              const topic = progress.topics.find((item) => item.id === attempt.topicId)
              return (
                <article key={attempt.attemptId}>
                  <div className={styles.attemptIcon} aria-hidden="true">
                    {String(topic?.order ?? attempt.level).padStart(2, '0')}
                  </div>
                  <div className={styles.attemptCopy}>
                    <h3>{topic?.title ?? `Тема ${attempt.topicId}`}</h3>
                    <p>
                      Уровень {attempt.level} · {formatAttemptDate(attempt.finishedAt)}
                    </p>
                  </div>
                  <div className={styles.attemptStars} aria-label={`${attempt.stars} из 3 Звёзд`}>
                    {Array.from({ length: 3 }, (_, index) => (
                      <Star
                        key={index}
                        size={16}
                        weight={index < attempt.stars ? 'fill' : 'regular'}
                        aria-hidden="true"
                      />
                    ))}
                  </div>
                  <b className={attempt.score === 100 ? styles.perfectScore : undefined}>
                    {attempt.score}%
                  </b>
                </article>
              )
            })}
          </div>
        ) : (
          <div className={styles.emptyAchievements}>
            <ClockCounterClockwise size={27} aria-hidden="true" />
            Завершённые Прохождения появятся здесь.
          </div>
        )}
      </section>
    </>
  )
}

function formatAttemptDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'дата не указана'

  return new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
  }).format(date)
}
