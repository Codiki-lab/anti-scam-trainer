import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useCurrentAccount } from '@/entities/user'
import { useAvitoChatRecommendation } from '@/features/avito-chat-recommendation'
import { getLearningActionPath } from '@/features/continue-learning'
import { uiStyles } from '@/shared/ui-kit'
import styles from './AvitoChatIntegrationPage.module.scss'

export function AvitoChatIntegrationPage() {
  const { account } = useCurrentAccount()
  const [hasConsent, setConsent] = useState(false)
  const state = useAvitoChatRecommendation(account.trainingRole)

  return (
    <section className={styles.page}>
      <header>
        <p className={uiStyles.eyebrow}>Демо-интеграция с чатом</p>
        <h1>Разберите подозрительную ситуацию</h1>
        <p>Это подготовленный макет. Он не подключён к настоящим чатам Avito.</p>
      </header>
      <div className={styles.chat}>
        <b>Городской велосипед · 18 000 ₽</b>
        {state.snapshot.map((message, index) => (
          <p
            key={`${message.role}-${index}`}
            className={message.role === 'user' ? styles.mine : undefined}
          >
            {message.text}
          </p>
        ))}
        <aside>
          <b>Ситуация может быть небезопасной</b>
          <span>В переписке предлагают подтвердить сделку на внешней странице.</span>
        </aside>
      </div>
      {!state.recommendation ? (
        <div className={styles.consent}>
          <label>
            <input
              type="checkbox"
              checked={hasConsent}
              onChange={(event) => setConsent(event.target.checked)}
            />
            Передать в тренажёр только этот обезличенный Снимок диалога
          </label>
          <button
            className={uiStyles.primaryButton}
            disabled={!hasConsent || state.isLoading}
            type="button"
            onClick={() => void state.submit()}
          >
            {state.isLoading ? 'Подбираем Тему…' : 'Разобрать ситуацию'}
          </button>
          {state.error && <p className={uiStyles.formError}>{state.error}</p>}
        </div>
      ) : (
        <aside className={styles.recommendation}>
          <p className={uiStyles.eyebrow}>Рекомендация</p>
          <h2>{state.recommendation.topic.title}</h2>
          <p>{state.recommendation.explanation}</p>
          <div className={uiStyles.buttonRow}>
            <Link
              className={uiStyles.primaryButton}
              to={getLearningActionPath(state.recommendation.nextAction, '')}
            >
              Начать рекомендованный шаг
            </Link>
            <Link className={uiStyles.secondaryButton} to="/lessons">
              Все Темы
            </Link>
          </div>
        </aside>
      )}
    </section>
  )
}
