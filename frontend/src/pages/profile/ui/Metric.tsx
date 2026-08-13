import type { ReactNode } from 'react'
import styles from './Profile.module.scss'

export function Metric({
  value,
  label,
  icon,
  tone,
}: {
  value: string
  label: string
  icon: ReactNode
  tone: 'blue' | 'purple' | 'green'
}) {
  return (
    <div className={`${styles.metric} ${styles[`metric${tone}`]}`}>
      <span className={styles.metricIcon} aria-hidden="true">
        {icon}
      </span>
      <div>
        <b>{value}</b>
        <span>{label}</span>
      </div>
    </div>
  )
}
