import {
  ArrowRight,
  ChatCircleText,
  CheckCircle,
  Clock,
  ShieldCheck,
  WarningCircle,
} from '@phosphor-icons/react'
import type { Theory, TheorySection } from '@/entities/learning'
import { uiStyles } from '@/shared/ui-kit'
import styles from './TheoryContent.module.scss'

interface TheoryContentProps {
  theory: Theory
  isSaving?: boolean
  onFinish: () => Promise<void>
}

type SectionRole = 'intro' | 'risk' | 'example' | 'safe_action' | 'summary'

const sectionAliases: Record<string, SectionRole> = {
  intro: 'intro',
  rule: 'intro',
  risk: 'risk',
  signs: 'risk',
  example: 'example',
  safe_action: 'safe_action',
  action: 'safe_action',
  summary: 'summary',
}

const sectionOrder: Record<SectionRole, number> = {
  intro: 1,
  risk: 2,
  example: 3,
  safe_action: 4,
  summary: 5,
}

function findSection(sections: TheorySection[], role: SectionRole) {
  return (
    sections.find((section) => sectionAliases[section.kind] === role) ||
    sections.find((section) => section.order === sectionOrder[role])
  )
}

function SectionText({ section, fallback }: { section?: TheorySection; fallback: string }) {
  return <>{section?.body || fallback}</>
}

export function TheoryContent({ theory, isSaving, onFinish }: TheoryContentProps) {
  const intro = findSection(theory.sections, 'intro')
  const risk = findSection(theory.sections, 'risk')
  const example = findSection(theory.sections, 'example')
  const safeAction = findSection(theory.sections, 'safe_action')
  const summary = findSection(theory.sections, 'summary')

  return (
    <article className={styles.theory}>
      <header className={styles.hero}>
        <div>
          <div className={styles.meta}>
            <span>Тема {theory.topic.order}</span>
            <span>
              <Clock size={16} weight="bold" aria-hidden="true" />
              {theory.sections.length} блоков
            </span>
          </div>
          <h1>{theory.topic.title}</h1>
          <p className={styles.lead}>{theory.topic.description}</p>
        </div>
        <div className={styles.heroIcon} aria-hidden="true">
          <ShieldCheck size={42} weight="duotone" />
        </div>
      </header>

      <div className={styles.lessonProgress} aria-label="Пять учебных блоков">
        {['Правило', 'Сигналы', 'Пример', 'Действие', 'Итог'].map((label, index) => (
          <div key={label}>
            <span>{index + 1}</span>
            <small>{label}</small>
          </div>
        ))}
      </div>

      <section className={styles.ruleBox}>
        <div className={styles.sectionIcon} aria-hidden="true">
          <ShieldCheck size={28} weight="fill" />
        </div>
        <div>
          <span className={styles.kicker}>Главное правило</span>
          <h2>{intro?.title || 'Оставайтесь внутри сервиса'}</h2>
          <p>
            <SectionText section={intro} fallback="Изучите, как устроен риск в этой Теме." />
          </p>
        </div>
      </section>

      <div className={styles.contentGrid}>
        <section className={styles.riskCard}>
          <div className={styles.cardHeading}>
            <WarningCircle size={26} weight="fill" aria-hidden="true" />
            <div>
              <span className={styles.kicker}>Красные флаги</span>
              <h2>{risk?.title || 'Как распознать риск'}</h2>
            </div>
          </div>
          <p>
            <SectionText section={risk} fallback="Изучите признаки риска для этой Темы." />
          </p>
        </section>

        <section className={styles.exampleCard}>
          <div className={styles.cardHeading}>
            <ChatCircleText size={26} weight="fill" aria-hidden="true" />
            <div>
              <span className={styles.kicker}>Ситуация в чате</span>
              <h2>{example?.title || 'Посмотрите на сообщение'}</h2>
            </div>
          </div>
          <div className={styles.chatPreview}>
            <div className={styles.chatHeader}>
              <span className={styles.avatar}>П</span>
              <div>
                <b>Собеседник</b>
                <small>по объявлению</small>
              </div>
            </div>
            <p>
              <SectionText section={example} fallback="Разберите пример ситуации из этой Темы." />
            </p>
          </div>
        </section>
      </div>

      <section className={styles.actionCard}>
        <div className={styles.actionIcon} aria-hidden="true">
          <CheckCircle size={30} weight="fill" />
        </div>
        <div>
          <span className={styles.kicker}>Безопасное действие</span>
          <h2>{safeAction?.title || 'Остановитесь и проверьте'}</h2>
          <p>
            <SectionText
              section={safeAction}
              fallback="Выполните безопасное действие, указанное в этой Теме."
            />
          </p>
        </div>
      </section>

      <section className={styles.summaryCard}>
        <CheckCircle size={30} weight="fill" aria-hidden="true" />
        <div>
          <span className={styles.kicker}>Запомните</span>
          <h2>{summary?.title || 'Главное за минуту'}</h2>
          <p>
            <SectionText
              section={summary}
              fallback="Повторите главное правило этой Темы перед квизом."
            />
          </p>
        </div>
      </section>

      <footer className={styles.finishPanel}>
        <div>
          <span>Финальный шаг</span>
          <h2>Теперь проверьте себя на практике</h2>
          <p>Пять коротких вопросов помогут закрепить безопасное действие.</p>
        </div>
        <button
          type="button"
          disabled={isSaving}
          className={`${uiStyles.primaryButton} ${styles.finishButton}`}
          onClick={() => void onFinish()}
        >
          {isSaving ? 'Сохраняем…' : 'Перейти к квизу'}
          {!isSaving && <ArrowRight size={18} weight="bold" aria-hidden="true" />}
        </button>
      </footer>
    </article>
  )
}
