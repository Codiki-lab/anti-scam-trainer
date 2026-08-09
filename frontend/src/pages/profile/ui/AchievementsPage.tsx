import { Link } from 'react-router-dom'
import type { Achievements } from '@/entities/progress'
import { useCurrentAccount } from '@/entities/user'
import { useAchievementData } from '@/features/view-progress'
import { ErrorState } from '@/shared/error-state'
import { useIsPreview } from '@/shared/runtime-mode'
import { uiStyles } from '@/shared/ui-kit'
import { AchievementCard } from './AchievementCard'
import styles from './Profile.module.scss'

export function AchievementsPage({ previewAchievements }: { previewAchievements?: Achievements }) {
  const isPreview = useIsPreview()
  const { account } = useCurrentAccount()
  const { achievements, isLoading, error, retry } = useAchievementData(previewAchievements)
  const basePath = isPreview ? '/preview' : ''

  return (
    <>
      <section className={styles.profileHero}>
        <div className={styles.profileAvatar}>{account.username.slice(0, 1).toUpperCase()}</div>
        <div>
          <p className={uiStyles.eyebrow}>Профиль</p>
          <h1>{account.username}</h1>
          <p className={`${uiStyles.muted} ${styles.profileSummary}`}>
            Роль: {account.trainingRole === 'buyer' ? 'Покупатель' : 'Продавец'} · Серия обучения:{' '}
            <b>🔥 {account.streak.current} дня</b>
          </p>
        </div>
        <Link
          className={`${uiStyles.secondaryButton} ${styles.roleLink}`}
          to={`${basePath}/dashboard`}
        >
          Сменить роль
        </Link>
      </section>

      <section>
        <div className={uiStyles.sectionHeading}>
          <div>
            <p className={uiStyles.eyebrow}>Достижения</p>
            <h2>Ваши награды</h2>
          </div>
          {achievements && (
            <span className={styles.achievementCount}>
              Получено {achievements.earned.length} из{' '}
              {achievements.earned.length + achievements.available.length}
            </span>
          )}
        </div>
        {isLoading && <p className={uiStyles.muted}>Загружаем достижения…</p>}
        {!isLoading && error && <ErrorState message={error} onRetry={() => void retry()} />}
        {!isLoading && !error && !achievements && (
          <p className={uiStyles.formError}>Не удалось загрузить достижения.</p>
        )}
        {achievements && (
          <div className={styles.achievementSections}>
            <section>
              <div className={styles.groupHeading}>
                <h3>Полученные</h3>
                <span>{achievements.earned.length}</span>
              </div>
              <div className={styles.achievementGrid}>
                {achievements.earned.map((achievement) => (
                  <AchievementCard key={achievement.code} achievement={achievement} />
                ))}
              </div>
            </section>
            <section>
              <div className={styles.groupHeading}>
                <h3>В процессе</h3>
                <span>{achievements.available.length}</span>
              </div>
              <div className={styles.achievementGrid}>
                {achievements.available.map((achievement) => (
                  <AchievementCard key={achievement.code} achievement={achievement} />
                ))}
              </div>
            </section>
          </div>
        )}
      </section>
    </>
  )
}
