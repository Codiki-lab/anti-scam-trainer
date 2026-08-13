import { CaretDown, Check } from '@phosphor-icons/react'
import { useEffect, useId, useRef, useState } from 'react'
import type { KeyboardEvent } from 'react'
import type { Topic } from '@/entities/learning'
import styles from './TrainingPage.module.scss'

interface TopicSelectProps {
  topics: Topic[]
  value: number
  onChange: (topicId: number) => void
}

export function TopicSelect({ topics, value, onChange }: TopicSelectProps) {
  const [open, setOpen] = useState(false)
  const rootRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const optionRefs = useRef<Array<HTMLButtonElement | null>>([])
  const labelId = useId()
  const listboxId = useId()
  const selected = topics.find((topic) => topic.id === value) ?? topics[0]

  useEffect(() => {
    if (!open) return

    const closeOnOutsideClick = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false)
    }
    const closeOnEscape = (event: globalThis.KeyboardEvent) => {
      if (event.key !== 'Escape') return
      setOpen(false)
      triggerRef.current?.focus()
    }

    document.addEventListener('pointerdown', closeOnOutsideClick)
    document.addEventListener('keydown', closeOnEscape)
    return () => {
      document.removeEventListener('pointerdown', closeOnOutsideClick)
      document.removeEventListener('keydown', closeOnEscape)
    }
  }, [open])

  const openWithFocus = (index: number) => {
    setOpen(true)
    requestAnimationFrame(() => optionRefs.current[index]?.focus())
  }

  const handleTriggerKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    if (event.key !== 'ArrowDown' && event.key !== 'ArrowUp') return
    event.preventDefault()
    openWithFocus(event.key === 'ArrowDown' ? 0 : topics.length - 1)
  }

  const handleOptionKeyDown = (event: KeyboardEvent<HTMLButtonElement>, index: number) => {
    const focusIndex = (nextIndex: number) => {
      event.preventDefault()
      optionRefs.current[nextIndex]?.focus()
    }

    if (event.key === 'ArrowDown') focusIndex((index + 1) % topics.length)
    if (event.key === 'ArrowUp') focusIndex((index - 1 + topics.length) % topics.length)
    if (event.key === 'Home') focusIndex(0)
    if (event.key === 'End') focusIndex(topics.length - 1)
  }

  const selectTopic = (topicId: number) => {
    onChange(topicId)
    setOpen(false)
    triggerRef.current?.focus()
  }

  if (!selected) return null

  return (
    <div className={styles.topicSelect} ref={rootRef}>
      <span className={styles.topicSelectLabel} id={labelId}>
        Тема тренировки
      </span>
      <button
        ref={triggerRef}
        className={styles.topicSelectTrigger}
        type="button"
        aria-expanded={open}
        aria-haspopup="listbox"
        aria-labelledby={`${labelId} ${listboxId}-value`}
        onClick={() => setOpen((current) => !current)}
        onKeyDown={handleTriggerKeyDown}
      >
        <span className={styles.topicSelectNumber}>{String(selected.order).padStart(2, '0')}</span>
        <span className={styles.topicSelectValue} id={`${listboxId}-value`}>
          <small>Выбрана</small>
          <b>{selected.title}</b>
        </span>
        <CaretDown size={18} weight="bold" aria-hidden="true" />
      </button>

      {open && (
        <div
          className={styles.topicSelectMenu}
          id={listboxId}
          role="listbox"
          aria-labelledby={labelId}
        >
          <div className={styles.topicSelectMenuHeading}>
            <b>Выберите Тему</b>
            <span>{topics.length} вариантов</span>
          </div>
          {topics.map((topic, index) => {
            const isSelected = topic.id === selected.id
            const status = topic.isCompleted
              ? 'Тема завершена'
              : topic.isQuizPassed
                ? 'Доступны Уровни'
                : topic.isTheoryRead
                  ? 'Пройдите квиз'
                  : 'Начните с Теории'

            return (
              <button
                key={topic.id}
                ref={(element) => {
                  optionRefs.current[index] = element
                }}
                className={`${styles.topicSelectOption} ${isSelected ? styles.topicSelectOptionActive : ''}`}
                type="button"
                role="option"
                aria-selected={isSelected}
                onClick={() => selectTopic(topic.id)}
                onKeyDown={(event) => handleOptionKeyDown(event, index)}
              >
                <span>{String(topic.order).padStart(2, '0')}</span>
                <span>
                  <b>{topic.title}</b>
                  <small>{status}</small>
                </span>
                {isSelected && <Check size={18} weight="bold" aria-hidden="true" />}
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
