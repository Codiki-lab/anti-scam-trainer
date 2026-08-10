import { useState } from 'react'
import { Link } from 'react-router-dom'
import type { Dashboard } from '@/entities/progress'
import { useCurrentAccount } from '@/entities/user'
import { TrainingRoleSelector } from '@/features/change-training-role'
import { useStartFreePlay } from '@/features/start-training'
import { useDashboardData } from '@/features/view-progress'
import { ErrorState } from '@/shared/error-state'
import { useIsPreview } from '@/shared/runtime-mode'
import { LoadingState, uiStyles } from '@/shared/ui-kit'
import { DailyTaskModal } from '@/widgets/daily-task'
import { TopicCard } from './TopicCard'
import { TrainingRow } from './TrainingRow'
import { getContinuePath } from '../lib/getContinuePath'
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

  if (isLoading) return <LoadingState label="Загружаем главную…" />
  if (error) return <ErrorState message={error} onRetry={() => void retry()} />
  if (!dashboard) return <p className={uiStyles.formError}>Не удалось загрузить главную.</p>

  const completedTopics = dashboard.topics.filter((topic) => topic.isCompleted).length
  const progress = Math.round((completedTopics / Math.max(dashboard.topics.length, 1)) * 100)
  const action = dashboard.continueAction
  const actionTopic = dashboard.topics.find((topic) => topic.id === action?.topicId)
  const actionCopy = action
    ? {
        resume_attempt: ['Продолжить Прохождение', 'Вернитесь к сохранённому Шагу сделки.'],
        read_theory: ['Читать Теорию', 'Начните с пяти коротких учебных блоков.'],
        take_quiz: ['Пройти Quiz', 'Наберите не менее 80%, чтобы открыть практику.'],
        start_level: [
          `Начать Уровень ${action.level ?? 1}`,
          'Примените знания в реалистичной сделке.',
        ],
        start_free_play: [
          'Начать Свободную игру',
          'Проверьте весь навык без заранее известного типа собеседника.',
        ],
      }[action.type]
    : ['Выбрать следующую Тему', 'Продолжите независимый Прогресс этой ролевой ветки.']

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
            <aside className={styles.recommendation} aria-label="Рекомендованное действие">
              <div>
                <span>Рекомендуем продолжить</span>
                <h2>{actionCopy[0]}</h2>
                <p>
                  {actionTopic ? `${actionTopic.title}. ` : ''}
                  {actionCopy[1]}
                </p>
                {action?.level && (
                  <small>
                    Уровень {action.level} ·{' '}
                    {actionTopic?.levels.find((level) => level.number === action.level)?.stars ?? 0}{' '}
                    из 3 Звёзд
                  </small>
                )}
              </div>
              {action?.type === 'start_free_play' ? (
                <button
                  className={uiStyles.primaryButton}
                  type="button"
                  disabled={isFreePlayStarting}
                  onClick={() =>
                    void startFreePlay().then((message) => setFreePlayError(message ?? ''))
                  }
                >
                  {isFreePlayStarting ? 'Запускаем…' : actionCopy[0]}
                </button>
              ) : (
                <Link className={uiStyles.primaryButton} to={getContinuePath(action, basePath)}>
                  {actionCopy[0]}
                </Link>
              )}
            </aside>
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
              <small>
                {completedTopics === 6
                  ? 'Открыта: завершены 6 из 6 Тем.'
                  : `Вы ещё не завершили теорию и Уровни этой ролевой ветки (${completedTopics}/6 Тем). Игра доступна для тестирования Ollama.`}
              </small>
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
