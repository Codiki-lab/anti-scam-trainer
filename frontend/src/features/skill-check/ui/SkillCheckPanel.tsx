import { Link } from 'react-router-dom'
import type { SkillCheck } from '@/entities/learning'
import { uiStyles } from '@/shared/ui-kit'
import { useSkillCheck } from '../model/useSkillCheck'
import styles from './SkillCheckPanel.module.scss'

interface SkillCheckPanelProps {
  topicId: number
  basePath: string
  preview?: SkillCheck
}

export function SkillCheckPanel({ topicId, basePath, preview }: SkillCheckPanelProps) {
  const state = useSkillCheck(topicId)
  const check = preview ?? state.check

  if (!check) {
    return (
      <section className={styles.panel}>
        <p className={uiStyles.eyebrow}>Проверка навыка</p>
        <h1>Сначала оцените ситуацию без подсказки</h1>
        <p>После обучения вы увидите тот же Снимок и сможете сравнить решение.</p>
        <button
          className={uiStyles.primaryButton}
          disabled={state.isLoading}
          type="button"
          onClick={() => void state.start()}
        >
          {state.isLoading ? 'Готовим ситуацию…' : 'Начать проверку'}
        </button>
        {state.error && <p className={uiStyles.formError}>{state.error}</p>}
      </section>
    )
  }

  if (check.phase === 'after_locked') {
    return (
      <section className={styles.panel}>
        <p className={uiStyles.eyebrow}>Первый Снимок сохранён</p>
        <h1>Пройдите Тему, чтобы сравнить решение</h1>
        <p>Проверка не начисляет Баллы и не меняет Прогресс.</p>
        <div className={uiStyles.buttonRow}>
          <Link className={uiStyles.primaryButton} to={`${basePath}/${topicId}`}>
            Перейти к Теории
          </Link>
          {!preview && (
            <button
              className={uiStyles.secondaryButton}
              disabled={state.isLoading}
              type="button"
              onClick={() => void state.refresh()}
            >
              Обновить состояние
            </button>
          )}
        </div>
      </section>
    )
  }

  if (check.phase === 'completed') {
    return (
      <section className={styles.panel}>
        <p className={uiStyles.eyebrow}>Сравнение завершено</p>
        <h1>{check.isImproved ? 'Решение стало безопаснее' : 'Закрепите безопасный паттерн'}</h1>
        <div className={styles.comparison}>
          <div>
            <b>До обучения</b>
            <p>{check.beforePattern ?? 'Первый ответ сохранён'}</p>
          </div>
          <div>
            <b>После обучения</b>
            <p>{check.afterPattern ?? 'Повторный ответ сохранён'}</p>
          </div>
        </div>
        <Link className={uiStyles.primaryButton} to={`${basePath}/${topicId}`}>
          Вернуться к Теме
        </Link>
      </section>
    )
  }

  return (
    <section className={styles.panel}>
      <p className={uiStyles.eyebrow}>
        {check.phase === 'before' ? 'До обучения' : 'После обучения'}
      </p>
      <h1>Кажется ли эта ситуация мошеннической?</h1>
      <div className={styles.dialogue}>
        {check.snapshot?.map((message, index) => (
          <p
            key={`${message.role}-${index}`}
            className={message.role === 'user' ? styles.mine : undefined}
          >
            {message.text}
          </p>
        ))}
      </div>
      <div className={uiStyles.buttonRow}>
        <button
          disabled={state.isLoading || Boolean(preview)}
          type="button"
          onClick={() => void state.answer(true)}
        >
          Да, здесь есть риск
        </button>
        <button
          disabled={state.isLoading || Boolean(preview)}
          type="button"
          onClick={() => void state.answer(false)}
        >
          Нет, ситуация безопасна
        </button>
      </div>
      {state.error && <p className={uiStyles.formError}>{state.error}</p>}
    </section>
  )
}
