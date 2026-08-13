import { ArrowRight, Fire, Medal, UserCircle } from '@phosphor-icons/react'
import type { CSSProperties } from 'react'
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
  const earnedCount = achievements?.earned.length ?? 0
  const total = achievements ? achievements.earned.length + achievements.available.length : 0
  const completion = total > 0 ? Math.round((earnedCount / total) * 100) : 0
  const roleName = account.trainingRole === 'buyer' ? 'Покупатель' : 'Продавец'
  const streakDayLabel = getDayLabel(account.streak.current)

  return (
    <>
      <section className={styles.achievementsHero}>
        <div className={styles.achievementsIntro}>
          <p className={uiStyles.eyebrow}>Ваш результат</p>
          <h1>Достижения</h1>
          <p className={uiStyles.muted}>
            Награды отмечают важные шаги в обучении и помогают видеть, что уже получилось.
          </p>

          <div className={styles.profileLine}>
            <span className={styles.profileAvatar}>
              {account.username.slice(0, 1).toUpperCase()}
            </span>
            <div>
              <b>{account.username}</b>
              <span>{roleName}</span>
            </div>
            <span className={styles.streakBadge}>
              <Fire size={18} weight="fill" aria-hidden="true" />
              Серия {account.streak.current} {streakDayLabel}
            </span>
            <Link className={styles.roleLink} to={`${basePath}/dashboard`}>
              Сменить роль <ArrowRight size={17} aria-hidden="true" />
            </Link>
          </div>
        </div>

        <div className={styles.achievementSummary}>
          <div
            className={styles.summaryRing}
            style={{ '--achievement-completion': `${completion * 3.6}deg` } as CSSProperties}
            aria-hidden="true"
          >
            <span>
              <b>{earnedCount}</b>
              <small>из {total || '—'}</small>
            </span>
          </div>
          <div>
            <Medal size={24} weight="duotone" aria-hidden="true" />
            <b>Собрано наград</b>
            <span>{completion}% коллекции</span>
          </div>
        </div>
      </section>

      <section>
        <div className={uiStyles.sectionHeading}>
          <div>
            <p className={uiStyles.eyebrow}>Коллекция</p>
            <h2>Ваши награды</h2>
          </div>
          <span className={styles.achievementHint}>Все условия проверяет тренажёр</span>
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
                <span className={`${styles.groupIcon} ${styles.groupIconEarned}`}>
                  <Medal size={19} weight="fill" aria-hidden="true" />
                </span>
                <div>
                  <h3>Полученные</h3>
                  <p>Уже в вашей коллекции</p>
                </div>
                <span>{achievements.earned.length}</span>
              </div>
              {achievements.earned.length > 0 ? (
                <div className={styles.achievementGrid}>
                  {achievements.earned.map((achievement) => (
                    <AchievementCard key={achievement.code} achievement={achievement} />
                  ))}
                </div>
              ) : (
                <div className={styles.emptyAchievements}>
                  <UserCircle size={28} aria-hidden="true" />
                  Завершите первое Прохождение — здесь появится первая награда.
                </div>
              )}
            </section>
            <section>
              <div className={styles.groupHeading}>
                <span className={styles.groupIcon}>
                  <Medal size={19} weight="duotone" aria-hidden="true" />
                </span>
                <div>
                  <h3>Следующие цели</h3>
                  <p>Прогресс обновляется после учебных действий</p>
                </div>
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

function getDayLabel(value: number) {
  const mod100 = value % 100
  const mod10 = value % 10
  if (mod100 >= 11 && mod100 <= 14) return 'дней'
  if (mod10 === 1) return 'день'
  if (mod10 >= 2 && mod10 <= 4) return 'дня'
  return 'дней'
}
