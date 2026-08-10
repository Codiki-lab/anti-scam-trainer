import {
  BookOpen,
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

const achievementPictures = {
  star: Star,
  stack: Stack,
  shield: ShieldCheck,
  book: BookOpen,
  buyer: ShoppingCart,
  seller: Storefront,
  flame: Fire,
} as const

const achievementPicturesByCode = {
  first_training: Star,
  five_trainings: Stack,
  perfect_score: ShieldCheck,
  first_topic_completed: BookOpen,
  all_buyer_topics: ShoppingCart,
  all_seller_topics: Storefront,
  streak_3: Fire,
  streak_7: Fire,
} as const

export function AchievementCard({ achievement }: { achievement: Achievement }) {
  const progress = Math.min(100, Math.round((achievement.current / achievement.target) * 100))
  const Picture =
    achievementPictures[achievement.icon as keyof typeof achievementPictures] ??
    achievementPicturesByCode[achievement.code as keyof typeof achievementPicturesByCode] ??
    Trophy

  return (
    <div className={`${styles.achievement} ${achievement.earned ? styles.earned : ''}`}>
      <span className={styles.achievementPicture} aria-hidden="true">
        <Picture size={38} weight={achievement.earned ? 'fill' : 'duotone'} />
      </span>
      <div className={styles.achievementBody}>
        <h3>{achievement.title}</h3>
        <p>{achievement.description}</p>
        {achievement.earned ? (
          <small className={styles.received}>Получено</small>
        ) : (
          <div className={styles.achievementProgress}>
            <div>
              <i style={{ width: `${progress}%` }} />
            </div>
            <small>
              {achievement.current} из {achievement.target}
            </small>
          </div>
        )}
      </div>
    </div>
  )
}
