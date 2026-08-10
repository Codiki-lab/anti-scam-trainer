import type { Theory } from '@/entities/learning'
import { uiStyles } from '@/shared/ui-kit'
import styles from './TheoryContent.module.scss'

interface TheoryContentProps {
  theory: Theory
  isSaving?: boolean
  onFinish: () => Promise<void>
}

export function TheoryContent({ theory, isSaving, onFinish }: TheoryContentProps) {
  const words = theory.sections.reduce(
    (total, section) => total + section.body.split(/\s+/).length,
    0,
  )
  const readingMinutes = Math.max(3, Math.ceil(words / 180))

  const renderBody = (kind: (typeof theory.sections)[number]['kind'], body: string) => {
    const fragments = body
      .split(/(?:\n+|(?<=[.!?])\s+)/)
      .map((item) => item.trim())
      .filter(Boolean)
    if (kind === 'risk' || kind === 'summary') {
      return (
        <ul>
          {fragments.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      )
    }
    if (kind === 'safe_action') {
      return (
        <ol>
          {fragments.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ol>
      )
    }
    if (kind === 'example') return <blockquote>{body}</blockquote>
    return <p>{body}</p>
  }

  return (
    <article className={styles.theory}>
      <p className={uiStyles.eyebrow}>Тема {theory.topic.order}</p>
      <h1>{theory.topic.title}</h1>
      <p className={styles.lead}>{theory.topic.description}</p>
      <div className={styles.meta} aria-label="Структура Теории">
        <span>≈ {readingMinutes} минут чтения</span>
        <span>5 учебных блоков</span>
      </div>
      {theory.sections.map((section) => (
        <section key={section.id} className={`${styles.section} ${styles[section.kind]}`}>
          <span className={styles.sectionNumber}>{section.order}/5</span>
          <h2>{section.title}</h2>
          {renderBody(section.kind, section.body)}
        </section>
      ))}
      <button
        type="button"
        disabled={isSaving}
        className={`${uiStyles.primaryButton} ${styles.finishButton}`}
        onClick={() => void onFinish()}
      >
        {isSaving ? 'Сохраняем…' : 'Проверить знания'}
      </button>
    </article>
  )
}
