import {
  ArrowLeft,
  Bicycle,
  Camera,
  Check,
  DotsThree,
  MagnifyingGlass,
  Microphone,
  Paperclip,
  ShieldWarning,
} from '@phosphor-icons/react'
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useCurrentAccount } from '@/entities/user'
import { useAvitoChatRecommendation } from '@/features/avito-chat-recommendation'
import { getLearningActionPath } from '@/features/continue-learning'
import { Brand } from '@/shared/brand'
import { uiStyles } from '@/shared/ui-kit'
import styles from './AvitoChatIntegrationPage.module.scss'

const profileLinks = [
  'Мои объявления',
  'Заказы',
  'Мои отзывы',
  'Избранное',
  'Бонусы',
  'Портал призов',
  'Приглашайте друзей',
  'Моё резюме',
  'Подработка',
  'Портфолио',
  'Авито Аукцион',
  'Гараж',
]

function Message({
  children,
  isMine = false,
}: {
  children: React.ReactNode
  isMine?: boolean
}) {
  return <p className={isMine ? styles.outgoingMessage : styles.incomingMessage}>{children}</p>
}

export function AvitoChatIntegrationPage() {
  const { account } = useCurrentAccount()
  const [isConsentOpen, setConsentOpen] = useState(false)
  const state = useAvitoChatRecommendation(account.trainingRole)

  const openConsent = () => setConsentOpen(true)
  const closeConsent = () => {
    if (!state.isLoading) setConsentOpen(false)
  }

  return (
    <section className={styles.page} aria-label="Демо-интеграция с чатом Avito">
      <header className={styles.avitoHeader}>
        <div className={styles.utilityNavigation}>
          <div>
            <a href="#business">Для бизнеса</a>
            <a href="#career">Карьера в Авито</a>
            <a href="#help">Помощь</a>
            <a href="#catalogues">Каталоги</a>
            <a href="#helping">#ЯПомогаю</a>
          </div>
          <div>
            <a href="#post-ad">+ Разместить объявление</a>
            <a href="#my-ads">Мои объявления</a>
            <button type="button" aria-label="Открыть избранное">♡</button>
            <button type="button" aria-label="Открыть профиль">●</button>
          </div>
        </div>
        <div className={styles.marketNavigation}>
          <Brand />
          <nav aria-label="Разделы Avito">
            <a href="#business-360">Бизнес360</a>
            <a href="#auto">Авто</a>
            <a href="#real-estate">Недвижимость</a>
            <a href="#work">Работа</a>
            <a href="#services">Услуги</a>
            <a href="#more">Ещё</a>
          </nav>
        </div>
      </header>

      <div className={styles.messengerLayout}>
        <aside className={styles.profileSidebar} aria-label="Профиль пользователя Avito">
          <div className={styles.profileAvatar}>А<span /></div>
          <h2>Алексей</h2>
          <p className={styles.rating}>4,9 ★★★★★ <a href="#reviews">12 отзывов</a></p>
          <nav aria-label="Навигация профиля">
            {profileLinks.map((label) => (
              <a key={label} href={`#${label}`}>
                {label}
              </a>
            ))}
          </nav>
          <div className={styles.sidebarSection}>
            <b>Сообщения</b>
            <a className={styles.selectedLink} href="#messages">Сообщения</a>
            <a href="#notifications">Уведомления</a>
          </div>
          <div className={styles.sidebarSection}>
            <a href="#wallet">Кошелёк</a>
            <a href="#paid-services">Платные услуги</a>
            <a href="#pro">Для профессионалов</a>
            <a href="#offers">Спецпредложения</a>
            <a href="#mailings">Рассылки</a>
          </div>
        </aside>

        <article className={styles.chatPanel}>
          <header className={styles.chatHeader}>
            <button className={styles.iconButton} type="button" aria-label="Вернуться к списку диалогов">
              <ArrowLeft size={27} weight="bold" />
            </button>
            <div className={styles.counterpartyAvatar}>Н</div>
            <div className={styles.counterparty}>
              <b>Никита</b>
              <span>в сети в 13:05</span>
            </div>
            <div className={styles.productPreview} aria-label="Объявление о городском велосипеде">
              <span className={styles.productImage}><Bicycle size={28} weight="duotone" /></span>
              <div>
                <b>Городской велосипед</b>
                <span>18 000 ₽ · Москва</span>
              </div>
            </div>
            <div className={styles.chatActions}>
              <button className={styles.iconButton} type="button" aria-label="Искать в диалоге">
                <MagnifyingGlass size={22} weight="bold" />
              </button>
              <button className={styles.iconButton} type="button" aria-label="Другие действия">
                <DotsThree size={25} weight="bold" />
              </button>
            </div>
          </header>

          <div className={styles.conversation}>
            <time>Сегодня</time>
            <Message>{state.snapshot[0].text}</Message>
            <Message isMine>
              {state.snapshot[1].text}
              <small>12:04 <Check size={13} weight="bold" aria-label="Прочитано" /></small>
            </Message>
            <Message>
              {state.snapshot[2].text}
              <a className={styles.fakeLink} href="#phishing-example">avito-delivery.example</a>
              <small>12:05</small>
            </Message>
            <aside className={styles.safetyWarning} aria-label="Предупреждение безопасности">
              <ShieldWarning size={20} weight="fill" aria-hidden="true" />
              <div>
                <b>Собеседник кажется подозрительным</b>
                <p>В переписке обнаружена просьба перейти по внешней ссылке. Проверьте признаки риска перед продолжением сделки.</p>
                <button type="button" onClick={openConsent}>
                  Пройти антискам-тренажёр <span aria-hidden="true">→</span>
                </button>
              </div>
            </aside>
          </div>

          <form className={styles.composer} onSubmit={(event) => event.preventDefault()}>
            <button className={styles.iconButton} type="button" aria-label="Прикрепить файл">
              <Paperclip size={25} weight="regular" />
            </button>
            <label className={styles.messageInput}>
              <span className="sr-only">Сообщение</span>
              <input placeholder="Сообщение" disabled />
            </label>
            <button className={styles.iconButton} type="button" aria-label="Сделать фото">
              <Camera size={22} weight="regular" />
            </button>
            <button className={styles.iconButton} type="button" aria-label="Записать голосовое сообщение">
              <Microphone size={22} weight="regular" />
            </button>
          </form>
          <p className={styles.demoNote}>Демонстрационный макет интеграции антискам-тренажёра</p>
        </article>
      </div>

      {isConsentOpen && (
        <div className={styles.modalBackdrop} role="presentation" onMouseDown={closeConsent}>
          <section
            className={styles.consentDialog}
            role="dialog"
            aria-modal="true"
            aria-labelledby="consent-title"
            onMouseDown={(event) => event.stopPropagation()}
          >
            {!state.recommendation ? (
              <>
                <p className={uiStyles.eyebrow}>Контекстная тренировка</p>
                <h1 id="consent-title">Разобрать ситуацию в тренажёре?</h1>
                <p>
                  Передадим только обезличенный Снимок диалога из трёх Сообщений. История чата, контакты и данные объявления не сохраняются.
                </p>
                <div className={styles.dialogActions}>
                  <button className={uiStyles.primaryButton} type="button" disabled={state.isLoading} onClick={() => void state.submit()}>
                    {state.isLoading ? 'Подбираем Тему…' : 'Продолжить'}
                  </button>
                  <button className={uiStyles.secondaryButton} type="button" disabled={state.isLoading} onClick={closeConsent}>
                    Отмена
                  </button>
                </div>
                {state.error && <p className={uiStyles.formError}>{state.error}</p>}
              </>
            ) : (
              <>
                <p className={uiStyles.eyebrow}>Рекомендация готова</p>
                <h1 id="consent-title">{state.recommendation.topic.title}</h1>
                <p>{state.recommendation.explanation}</p>
                <div className={styles.dialogActions}>
                  <Link className={uiStyles.primaryButton} to={getLearningActionPath(state.recommendation.nextAction, '')}>
                    Начать тренировку
                  </Link>
                  <Link className={uiStyles.secondaryButton} to="/lessons">Все Темы</Link>
                </div>
              </>
            )}
          </section>
        </div>
      )}
    </section>
  )
}
