import { useEffect, useRef } from 'react'
import styles from './ConfirmDialog.module.scss'

interface ConfirmDialogProps {
  open: boolean
  title: string
  description: string
  confirmLabel: string
  isPending?: boolean
  onConfirm: () => void
  onClose: () => void
}

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel,
  isPending = false,
  onConfirm,
  onClose,
}: ConfirmDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return
    if (open && !dialog.open) dialog.showModal()
    if (!open && dialog.open) dialog.close()
  }, [open])

  return (
    <dialog
      ref={dialogRef}
      className={styles.dialog}
      aria-labelledby="confirm-dialog-title"
      onCancel={(event) => {
        event.preventDefault()
        onClose()
      }}
      onClose={onClose}
    >
      <h2 id="confirm-dialog-title">{title}</h2>
      <p>{description}</p>
      <div className={styles.actions}>
        <button type="button" onClick={onClose} disabled={isPending} autoFocus>
          Отмена
        </button>
        <button className={styles.danger} type="button" onClick={onConfirm} disabled={isPending}>
          {isPending ? 'Подождите…' : confirmLabel}
        </button>
      </div>
    </dialog>
  )
}
