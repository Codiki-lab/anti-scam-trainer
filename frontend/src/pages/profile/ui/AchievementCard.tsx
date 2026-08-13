import {
  BookOpenText,
  Fire,
  ShieldCheck,
  ShoppingCart,
  Stack,
  Star,
  Storefront,
  Trophy,
} from '@phosphor-icons/react'
import type { Achievement } from '@/entities/progress'
import styles from './Profile.module.scss'

const achievementIcons = {
  star: Star,
  stack: Stack,
  shield: ShieldCheck,
  book: BookOpenText,
  buyer: ShoppingCart,
  seller: Storefront,
  flame: Fire,
} as const

const achievementTones: Record<string, string> = {
  first_training: styles.blue,
  five_trainings: styles.purple,
  perfect_score: styles.green,
  first_topic_completed: styles.blue,
  all_buyer_topics: styles.purple,
  all_seller_topics: styles.green,
  streak_3: styles.coral,
  streak_7: styles.coral,
}

function formatEarnedAt(value?: string) {
  if (!value) return null
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return null

  return new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
  }).format(date)
}

export function AchievementCard({ achievement }: { achievement: Achievement }) {
  const current = Math.min(achievement.current, achievement.target)
  const progress = Math.min(100, Math.round((current / achievement.target) * 100))
  const Icon = achievementIcons[achievement.icon as keyof typeof achievementIcons] ?? Trophy
  const earnedAt = formatEarnedAt(achievement.earnedAt)
  const tone = achievementTones[achievement.code] ?? styles.blue

  return (
    <article
      className={`${styles.achievement} ${tone} ${achievement.earned ? styles.earned : styles.locked}`}
    >
      <span className={styles.achievementIcon} aria-hidden="true">
        <Icon size={26} weight={achievement.earned ? 'fill' : 'regular'} />
      </span>
      <div className={styles.achievementBody}>
        <div className={styles.achievementTitleRow}>
          <h3>{achievement.title}</h3>
          {achievement.earned && <span className={styles.earnedBadge}>Получено</span>}
        </div>
        <p>{achievement.description}</p>
        {achievement.earned ? (
          <small className={styles.received}>
            {earnedAt ? `Получено ${earnedAt}` : 'Условие выполнено'}
          </small>
        ) : (
          <div className={styles.achievementProgress}>
            <div
              role="progressbar"
              aria-label={`Прогресс достижения «${achievement.title}»`}
              aria-valuemin={0}
              aria-valuemax={achievement.target}
              aria-valuenow={current}
            >
              <i style={{ width: `${progress}%` }} />
            </div>
            <small>
              {current} из {achievement.target}
            </small>
          </div>
        )}
      </div>
    </article>
  )
}
