import { useState } from 'react'
import { Link } from 'react-router-dom'
import type { Dashboard } from '@/entities/progress'
import { useCurrentAccount } from '@/entities/user'
import { TrainingRoleSelector } from '@/features/change-training-role'
import { LearningRecommendationCard } from '@/features/continue-learning'
import { useStartFreePlay } from '@/features/start-training'
import { useDashboardData } from '@/features/view-progress'
import { ErrorState } from '@/shared/error-state'
import { useIsPreview } from '@/shared/runtime-mode'
import { uiStyles } from '@/shared/ui-kit'
import { DailyTaskModal } from '@/widgets/daily-task'
import { TopicCard } from './TopicCard'
import { TrainingRow } from './TrainingRow'
import styles from './Dashboard.module.scss'

export function DashboardPage({ previewDashboard }: { previewDashboard?: Dashboard }) {
  const isPreview = useIsPreview()
  const { account } = useCurrentAccount()
  const { dashboard, isLoading, error, retry } = useDashboardData(
    account.trainingRole,
    previewDashboard,
  )
  const basePath = isPreview ? '/preview' : ''
  const { start: startFreePlay, isStarting: isFreePlayStarting } = useStartFreePlay(
    account.trainingRole,
  )
  const [freePlayError, setFreePlayError] = useState('')
  const [isDailyTaskOpen, setDailyTaskOpen] = useState(false)

  if (isLoading) return <p className={uiStyles.muted}>Загружаем главную…</p>
  if (error) return <ErrorState message={error} onRetry={() => void retry()} />
  if (!dashboard) return <p className={uiStyles.formError}>Не удалось загрузить главную.</p>

  const completedTopics = dashboard.topics.filter((topic) => topic.isCompleted).length
  const progress = Math.round((completedTopics / Math.max(dashboard.topics.length, 1)) * 100)

  return (
    <>
      <section className={styles.dashboardOverview}>
        <div className={styles.mainColumn}>
          <div className={styles.heroContent}>
            <p className={uiStyles.eyebrow}>Тренажёр безопасности</p>
            <h1>Учитесь защищать себя и свои сделки</h1>
            <p className={uiStyles.muted}>
              Небольшие задания помогут вовремя заметить опасный сигнал.
            </p>
            <TrainingRoleSelector />
          </div>
          <div className={styles.topicsSection}>
            <div className={uiStyles.sectionHeading}>
              <div>
                <p className={uiStyles.eyebrow}>Обучение</p>
                <h2>Темы безопасности</h2>
              </div>
              <Link to={`${basePath}/lessons`}>Все темы →</Link>
            </div>
            <div className={styles.topicGrid}>
              {dashboard.topics.slice(0, 3).map((topic) => (
                <TopicCard key={topic.id} topic={topic} basePath={basePath} />
              ))}
            </div>
          </div>
        </div>

        <div className={styles.sideColumn}>
          <LearningRecommendationCard
            role={account.trainingRole}
            basePath={basePath}
            disabled={isPreview}
          />
          <div className={styles.quickActions}>
            <aside className={styles.dailyCard}>
              <span>Задание дня</span>
              <h3>{dashboard.dailyTask.isCompleted ? 'Задание выполнено' : 'Мошенник или нет?'}</h3>
              <p>
                {dashboard.dailyTask.isCompleted
                  ? 'Откройте сохранённый разбор ситуации.'
                  : 'Разберите короткую ситуацию и определите, безопасна ли сделка.'}
              </p>
              <button onClick={() => setDailyTaskOpen(true)} type="button">
                {dashboard.dailyTask.isCompleted ? 'Открыть разбор →' : 'Проверить сделку →'}
              </button>
            </aside>
            <aside className={styles.freePlayCard}>
              <span>Свободная игра</span>
              <h3>Диалог без подсказок</h3>
              <p>Сложность подстроится под ваши знания и результаты тренировок.</p>
              <button
                disabled={isFreePlayStarting}
                type="button"
                onClick={() =>
                  void startFreePlay().then((message) => setFreePlayError(message ?? ''))
                }
              >
                {isFreePlayStarting ? 'Запускаем…' : 'Начать игру →'}
              </button>
              {freePlayError && <small className={styles.cardError}>{freePlayError}</small>}
            </aside>
          </div>

          <div className={styles.dashboardSide}>
            <h3>Ваш прогресс</h3>
            <div className={styles.progressValue}>
              <b>{progress}%</b>
              <span>Пройдено тем</span>
            </div>
            <div className={styles.progressTrack}>
              <i style={{ width: `${progress}%` }} />
            </div>
            <div className={styles.miniStat}>
              <span>Завершено тем</span>
              <b>{completedTopics}</b>
            </div>
            <div className={styles.miniStat}>
              <span>Серия дней</span>
              <b>{dashboard.streak.current}</b>
            </div>
          </div>
        </div>
      </section>

      <section>
        <div className={uiStyles.sectionHeading}>
          <div>
            <p className={uiStyles.eyebrow}>Практика</p>
            <h2>Тренировки</h2>
          </div>
          <Link to={`${basePath}/chats`}>Все тренировки →</Link>
        </div>
        <div className={styles.trainingList}>
          {dashboard.topics.slice(0, 3).map((topic) => (
            <TrainingRow key={topic.id} topic={topic} basePath={basePath} />
          ))}
        </div>
      </section>
      <DailyTaskModal
        isOpen={isDailyTaskOpen}
        onClose={() => setDailyTaskOpen(false)}
        onConflict={retry}
        task={dashboard.dailyTask}
      />
    </>
  )
}
