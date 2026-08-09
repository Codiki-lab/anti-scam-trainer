import type { Theory } from '@/entities/learning'
import { uiStyles } from '@/shared/ui-kit'
import styles from './TheoryContent.module.scss'

interface TheoryContentProps {
  theory: Theory
  isSaving?: boolean
  onFinish: () => Promise<void>
}

export function TheoryContent({ theory, isSaving, onFinish }: TheoryContentProps) {
  return (
    <article className={styles.theory}>
      <p className={uiStyles.eyebrow}>Тема {theory.topic.order}</p>
      <h1>{theory.topic.title}</h1>
      <p className={styles.lead}>{theory.topic.description}</p>
      {theory.sections.map((section) => (
        <section key={section.id} className={section.kind === 'rule' ? styles.ruleBox : undefined}>
          <h2>{section.title}</h2>
          <p>{section.body}</p>
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
