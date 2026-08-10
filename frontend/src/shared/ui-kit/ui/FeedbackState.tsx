import type { ReactNode } from 'react'
import styles from './FeedbackState.module.scss'

export function LoadingState({ label = 'Загружаем данные…' }: { label?: string }) {
  return (
    <section className={styles.loading} role="status" aria-busy="true" aria-label={label}>
      <span />
      <span />
      <span />
    </section>
  )
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string
  description: string
  action?: ReactNode
}) {
  return (
    <section className={styles.empty} aria-label={title}>
      <h2>{title}</h2>
      <p>{description}</p>
      {action}
    </section>
  )
}

export function StatusBadge({
  children,
  tone = 'default',
}: {
  children: ReactNode
  tone?: 'default' | 'success' | 'progress' | 'locked'
}) {
  return <span className={`${styles.status} ${styles[tone]}`}>{children}</span>
}
