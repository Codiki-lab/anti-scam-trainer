import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { Topic } from '@/entities/learning'
import type { LevelState } from '@/entities/training'
import { TrainingList } from './TrainingList'

const topic: Topic = {
  id: 1,
  slug: 'phishing',
  role: 'buyer',
  title: 'Фишинговые ссылки',
  description: 'Распознавайте поддельные страницы.',
  order: 1,
  isTheoryRead: true,
  isQuizPassed: true,
  bestQuizScore: 80,
  isCompleted: false,
  levels: [
    { number: 1, isOpened: true, bestScore: 75, stars: 2, attempts: 1, lastAttemptId: 8 },
    { number: 2, isOpened: true, bestScore: 0, stars: 0, attempts: 0, lastAttemptId: null },
  ],
}

const levels: LevelState[] = [
  {
    number: 1,
    isOpened: true,
    scenarioId: 11,
    scenarioTitle: 'Безопасная ссылка',
    scenarioDescription: 'Выберите безопасный ответ.',
    responseMode: 'multiple_choice',
  },
  {
    number: 2,
    isOpened: true,
    scenarioId: 12,
    scenarioTitle: 'Похожие ответы',
    scenarioDescription: 'Продолжите начатое Прохождение.',
    responseMode: 'similar_choice',
    inProgressAttemptId: 44,
  },
]

describe('TrainingList', () => {
  it('renders backend scenario data and progress-dependent actions', () => {
    const onStart = vi.fn()
    render(<TrainingList topic={topic} levels={levels} isStarting={false} onStart={onStart} />)

    expect(screen.getByRole('heading', { name: 'Безопасная ссылка' })).toBeInTheDocument()
    expect(screen.getByText('Готовые варианты')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Пройти ещё раз/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Продолжить/ })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Продолжить/ }))
    expect(onStart).toHaveBeenCalledWith(2)
  })
})
