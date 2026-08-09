import { RoleSelector } from '@/entities/user'
import { uiStyles } from '@/shared/ui-kit'
import { useChangeTrainingRole } from '../model/useChangeTrainingRole'
import styles from './TrainingRoleSelector.module.scss'

export function TrainingRoleSelector() {
  const { role, changeRole, isChanging, error } = useChangeTrainingRole()

  return (
    <div className={styles.root}>
      <span>Тренироваться как</span>
      <RoleSelector
        value={role}
        variant="segmented"
        disabled={isChanging}
        onChange={(nextRole) => void changeRole(nextRole)}
      />
      {error && (
        <p className={uiStyles.fieldError} role="alert">
          {error}
        </p>
      )}
    </div>
  )
}
