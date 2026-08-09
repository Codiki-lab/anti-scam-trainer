import styles from './Profile.module.scss'

export function Metric({ value, label }: { value: string; label: string }) {
  return (
    <div className={styles.metric}>
      <b>{value}</b>
      <span>{label}</span>
    </div>
  )
}
