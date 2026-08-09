import { AuthForm } from '@/features/auth'
import { Brand } from '@/shared/brand'
import { uiStyles } from '@/shared/ui-kit'
import styles from './AuthPage.module.scss'

export function AuthPage({ mode }: { mode: 'login' | 'register' }) {
  return (
    <main className={styles.page}>
      <section className={styles.hero} aria-labelledby="auth-title">
        <Brand />

        <div>
          <p className={uiStyles.eyebrow}>Безопасность на Avito</p>
          <h1 id="auth-title">Учитесь замечать обман до того, как он случится.</h1>
          <p className={styles.intro}>
            Пройдите короткие тренировки в роли покупателя или продавца. Ошибаться здесь безопасно.
          </p>
        </div>

        <div className={styles.rule}>
          <span aria-hidden="true">✓</span>
          <div>
            <b>Главное правило</b>
            <br />
            Оплата и доставка — только внутри Avito.
          </div>
        </div>
      </section>

      <section className={styles.cardWrap} aria-label={mode === 'login' ? 'Вход' : 'Регистрация'}>
        <AuthForm mode={mode} />
      </section>
    </main>
  )
}
