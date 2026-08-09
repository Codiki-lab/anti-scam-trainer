interface ErrorStateProps {
  message: string
  onRetry?: () => void
}

export function ErrorState({ message, onRetry }: ErrorStateProps) {
  return (
    <section className={styles.state} role="alert">
      <h2>Не удалось загрузить данные</h2>
      <p>{message}</p>
      {onRetry && (
        <button className={uiStyles.secondaryButton} type="button" onClick={onRetry}>
          Попробовать снова
        </button>
      )}
    </section>
  )
}
import { uiStyles } from '@/shared/ui-kit'
import styles from './ErrorState.module.scss'
