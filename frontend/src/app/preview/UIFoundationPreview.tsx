import { useState } from 'react'
import { ErrorState } from '@/shared/error-state'
import { ConfirmDialog, EmptyState, LoadingState, StatusBadge, uiStyles } from '@/shared/ui-kit'
import styles from './UIFoundationPreview.module.scss'

export function UIFoundationPreview() {
  const [dialogOpen, setDialogOpen] = useState(false)

  return (
    <article>
      <section className={uiStyles.pageHeading}>
        <p className={uiStyles.eyebrow}>Визуальная документация</p>
        <h1>Состояния UI-основы</h1>
        <p className={uiStyles.muted}>Production-компоненты в основных интерактивных состояниях.</p>
      </section>
      <div className={styles.grid}>
        <section>
          <h2>Кнопки</h2>
          <div className={styles.row}>
            <button className={uiStyles.primaryButton}>Default</button>
            <button className={uiStyles.primaryButton} disabled>
              Disabled
            </button>
            <button className={uiStyles.primaryButton}>Loading…</button>
            <button className={uiStyles.secondaryButton} onClick={() => setDialogOpen(true)}>
              Открыть диалог
            </button>
          </div>
        </section>
        <section>
          <h2>Статусы</h2>
          <div className={styles.row}>
            <StatusBadge>Default</StatusBadge>
            <StatusBadge tone="progress">В процессе</StatusBadge>
            <StatusBadge tone="success">Успех</StatusBadge>
            <StatusBadge tone="locked">Закрыто</StatusBadge>
          </div>
        </section>
        <section>
          <h2>Skeleton</h2>
          <LoadingState />
        </section>
        <section>
          <h2>Empty</h2>
          <EmptyState
            title="Пока пусто"
            description="Данные появятся после первого учебного действия."
          />
        </section>
        <section>
          <h2>Error</h2>
          <ErrorState
            message="Проверьте соединение и повторите запрос."
            onRetry={() => undefined}
          />
        </section>
      </div>
      <ConfirmDialog
        open={dialogOpen}
        title="Подтвердить действие?"
        description="Диалог демонстрирует focus, cancel и pending состояния."
        confirmLabel="Подтвердить"
        onClose={() => setDialogOpen(false)}
        onConfirm={() => setDialogOpen(false)}
      />
    </article>
  )
}
