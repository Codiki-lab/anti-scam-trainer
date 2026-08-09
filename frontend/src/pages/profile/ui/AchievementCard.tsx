import type { Achievement } from '@/entities/progress'
import styles from './Profile.module.scss'

export function AchievementCard({ achievement }: { achievement: Achievement }) {
  const progress = Math.min(100, Math.round((achievement.current / achievement.target) * 100))

  return (
    <div className={`${styles.achievement} ${achievement.earned ? styles.earned : ''}`}>
      <span>{achievement.icon || (achievement.earned ? '✦' : '🔒')}</span>
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
