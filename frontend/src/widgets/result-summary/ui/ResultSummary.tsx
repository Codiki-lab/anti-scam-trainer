import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import type { AttemptResult } from '@/entities/training'
import { Stars } from '@/shared/stars'
import { uiStyles } from '@/shared/ui-kit'
import styles from './ResultSummary.module.scss'

interface ResultSummaryProps {
  result: AttemptResult
  basePath?: string
  nextActionHref?: string
  microQuestion?: ReactNode
}

const assessmentLabel = {
  unsafe: 'опасно',
  risky: 'риск',
  mostly_safe: 'почти безопасно',
  safe: 'безопасно',
} as const

export function ResultSummary({
  result,
  basePath = '',
  nextActionHref,
  microQuestion,
}: ResultSummaryProps) {
  return (
    <section className={styles.result}>
      <p className={uiStyles.eyebrow}>Прохождение завершено</p>
      <div className={styles.scoreCard}>
        <b>
          {result.score}
          <small>/100</small>
        </b>
        <div>
          <h1>{result.score >= 75 ? 'Хороший результат' : 'Есть что улучшить'}</h1>
          <p>Разберите ответы и используйте эти правила в настоящих сделках.</p>
          <Stars value={result.stars} />
        </div>
      </div>
      <section className={styles.feedback}>
        <p className={uiStyles.eyebrow}>Почему такой Result</p>
        <h2>{result.feedback.reason}</h2>
        <p>
          <b>Безопасная альтернатива:</b> {result.feedback.safeAlternative}
        </p>
      </section>
      <h2>Разбор сделки</h2>
      <div className={styles.checkList}>
        {result.decisionReview.map((answer, index) => (
          <p key={answer.stepId} className={answer.assessment === 'safe' ? undefined : styles.risk}>
            {answer.assessment === 'safe' ? '✓' : '!'}
            <span>
              <b>
                Шаг {answer.stepNumber || index + 1} — {assessmentLabel[answer.assessment]}
              </b>
              {answer.explanation} <strong>{answer.safeAction}</strong>
            </span>
          </p>
        ))}
      </div>
      {(result.riskSignals.length > 0 || result.safeActions.length > 0) && (
        <div className={styles.guidance}>
          <div>
            <h2>Сигналы риска</h2>
            <ul>
              {result.riskSignals.map((signal) => (
                <li key={signal.code}>{signal.label}</li>
              ))}
            </ul>
          </div>
          <div>
            <h2>Безопасные действия</h2>
            <ul>
              {result.safeActions.map((action) => (
                <li key={action}>{action}</li>
              ))}
            </ul>
          </div>
        </div>
      )}
      {result.newAchievements.length > 0 && (
        <section className={styles.achievements}>
          <h2>Новые достижения</h2>
          {result.newAchievements.map((achievement) => (
            <p key={achievement.code}>
              <span>{achievement.icon || '✦'}</span>
              <b>{achievement.title}</b> — {achievement.description}
            </p>
          ))}
        </section>
      )}
      {microQuestion}
      <div className={uiStyles.buttonRow}>
        <Link className={uiStyles.primaryButton} to={nextActionHref ?? `${basePath}/chats`}>
          Продолжить обучение
        </Link>
        <Link className={uiStyles.secondaryButton} to={`${basePath}/progress`}>
          Мой прогресс
        </Link>
      </div>
    </section>
  )
}
