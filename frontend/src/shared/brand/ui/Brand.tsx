import avitoLogo from '../assets/avito-logo.svg'
import styles from './Brand.module.scss'

export function Brand() {
  return (
    <span className={styles.brand} aria-label="Avito" role="img">
      <img src={avitoLogo} alt="" aria-hidden="true" />
    </span>
  )
}
