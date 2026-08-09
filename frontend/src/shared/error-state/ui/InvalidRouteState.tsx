import { Link } from 'react-router-dom'
import { uiStyles } from '@/shared/ui-kit'
import styles from './ErrorState.module.scss'

interface InvalidRouteStateProps {
  backTo: string
  backLabel: string
}

export function InvalidRouteState({ backTo, backLabel }: InvalidRouteStateProps) {
  return (
    <section className={styles.state} role="alert">
      <h1>Некорректная ссылка</h1>
      <p>В адресе отсутствует правильный идентификатор. Выберите нужный раздел из списка.</p>
      <Link className={uiStyles.primaryButton} to={backTo}>
        {backLabel}
      </Link>
    </section>
  )
}
