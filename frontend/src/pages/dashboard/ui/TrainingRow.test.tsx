import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import type { Topic } from '@/entities/learning'
import { TrainingRow } from './TrainingRow'

const topic: Topic = {
  id: 1,
  slug: 'phishing-links',
  role: 'buyer',
  title: 'Фишинговые ссылки',
  description: 'Проверяйте ссылки.',
  order: 1,
  isTheoryRead: true,
  isQuizPassed: true,
  bestQuizScore: 100,
  isCompleted: false,
  levels: [
    { number: 1, isOpened: true, bestScore: 100, stars: 3, attempts: 1, lastAttemptId: 1 },
    { number: 2, isOpened: true, bestScore: 80, stars: 1, attempts: 1, lastAttemptId: 2 },
    { number: 3, isOpened: true, bestScore: 0, stars: 0, attempts: 0, lastAttemptId: null },
    { number: 4, isOpened: false, bestScore: 0, stars: 0, attempts: 0, lastAttemptId: null },
  ],
}

describe('TrainingRow', () => {
  it('shows a dot for every level and marks completed levels', () => {
    render(
      <MemoryRouter>
        <TrainingRow topic={topic} basePath="" />
      </MemoryRouter>,
    )

    expect(screen.getByLabelText('Пройдено 2 из 4 уровней')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Продолжить' })).toBeInTheDocument()
  })
})
