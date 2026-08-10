import {
  BookOpen,
  Fire,
  ShieldCheck,
  ShoppingCart,
  Stack,
  Star,
  Storefront,
  Trophy,
  type Icon,
} from '@phosphor-icons/react'
import type { Achievement } from '@/entities/progress'
import styles from './Profile.module.scss'

const achievementPictures: Record<string, Icon> = {
  star: Star,
  first_training: Star,
  stack: Stack,
  five_trainings: Stack,
  shield: ShieldCheck,
  perfect_score: ShieldCheck,
  book: BookOpen,
  first_topic_completed: BookOpen,
  buyer: ShoppingCart,
  all_buyer_topics: ShoppingCart,
  seller: Storefront,
  all_seller_topics: Storefront,
  flame: Fire,
  streak_3: Fire,
  streak_7: Fire,
}

export function AchievementCard({ achievement }: { achievement: Achievement }) {
  const progress = Math.min(100, Math.round((achievement.current / achievement.target) * 100))
  const Picture =
    achievementPictures[achievement.icon] ?? achievementPictures[achievement.code] ?? Trophy

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
