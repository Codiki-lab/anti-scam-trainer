import { useState } from 'react'
import { AdminScenarios } from './AdminScenarios'
import { AdminTopics } from './AdminTopics'
import styles from './AdminPage.module.scss'

export function AdminPage() {
  const [section, setSection] = useState<'topics' | 'scenarios'>('topics')

  return (
    <section className={styles.adminPage}>
      <header className={styles.pageHeader}>
        <div>
          <span className={styles.eyebrow}>Управление контентом</span>
          <h1>Админ-панель</h1>
          <p>Создавайте и публикуйте Темы, Теорию, квизы и Сценарии тренировок.</p>
        </div>
        <div className={styles.sectionTabs} role="tablist" aria-label="Раздел админ-панели">
          <button
            className={section === 'topics' ? styles.activeTab : undefined}
            type="button"
            onClick={() => setSection('topics')}
          >
            Темы и обучение
          </button>
          <button
            className={section === 'scenarios' ? styles.activeTab : undefined}
            type="button"
            onClick={() => setSection('scenarios')}
          >
            Сценарии
          </button>
        </div>
      </header>

      {section === 'topics' ? <AdminTopics /> : <AdminScenarios />}
    </section>
  )
}
