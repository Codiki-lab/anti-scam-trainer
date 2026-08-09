import type { UserRole } from '../model/types'
import styles from './RoleSelector.module.scss'

interface RoleSelectorProps {
  value: UserRole
  variant?: 'cards' | 'select' | 'segmented'
  disabled?: boolean
  onChange: (role: UserRole) => void
}

export function RoleSelector({ value, variant = 'cards', disabled, onChange }: RoleSelectorProps) {
  if (variant === 'select') {
    return (
      <select
        className={styles.select}
        disabled={disabled}
        value={value}
        onChange={(event) => onChange(event.target.value as UserRole)}
      >
        <option value="buyer">Я в роли: Покупатель</option>
        <option value="seller">Я в роли: Продавец</option>
      </select>
    )
  }

  if (variant === 'segmented') {
    return (
      <div className={styles.segmented} aria-label="Роль в тренировках">
        <button
          className={value === 'buyer' ? styles.segmentActive : undefined}
          disabled={disabled}
          type="button"
          onClick={() => onChange('buyer')}
        >
          Покупатель
        </button>
        <button
          className={value === 'seller' ? styles.segmentActive : undefined}
          disabled={disabled}
          type="button"
          onClick={() => onChange('seller')}
        >
          Продавец
        </button>
      </div>
    )
  }

  return (
    <div className={styles.picker}>
      <button
        className={`${styles.button} ${value === 'buyer' ? styles.active : ''}`}
        disabled={disabled}
        type="button"
        onClick={() => onChange('buyer')}
      >
        <span>🛍️</span>Покупатель<small>Покупаю товары</small>
      </button>
      <button
        className={`${styles.button} ${value === 'seller' ? styles.active : ''}`}
        disabled={disabled}
        type="button"
        onClick={() => onChange('seller')}
      >
        <span>🏷️</span>Продавец<small>Продаю товары</small>
      </button>
    </div>
  )
}
