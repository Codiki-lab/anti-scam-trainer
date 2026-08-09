import styles from './Stars.module.scss'

export function Stars({ value }: { value: number }) {
  return (
    <span className={styles.stars} aria-label={`${value} из 3 звёзд`}>
      {[1, 2, 3].map((star) => (
        <b key={star} className={star <= value ? styles.active : undefined}>
          ★
        </b>
      ))}
    </span>
  )
}
