import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'
import type { Topic } from '@/entities/learning'
import type { LevelState } from '@/entities/training'
import { TrainingList } from './TrainingList'

const topic: Topic = {
  id: 1,
  slug: 'phishing',
  role: 'buyer',
  title: 'Фишинговые ссылки',
  description: 'Проверяйте страницы оплаты.',
  order: 1,
  isTheoryRead: true,
  isQuizPassed: true,
  bestQuizScore: 100,
  isCompleted: false,
  levels: [
    { number: 1, isOpened: true, bestScore: 0, stars: 0, attempts: 0, lastAttemptId: null },
    { number: 2, isOpened: true, bestScore: 25, stars: 0, attempts: 1, lastAttemptId: 12 },
    { number: 3, isOpened: true, bestScore: 100, stars: 3, attempts: 2, lastAttemptId: 13 },
    { number: 4, isOpened: false, bestScore: 0, stars: 0, attempts: 0, lastAttemptId: null },
  ],
}

const levels: LevelState[] = [
  {
    number: 1,
    isOpened: true,
    scenarioId: 1,
    scenarioTitle: 'Новый сценарий',
    scenarioDescription: 'Первый запуск.',
    responseType: 'multiple_choice',
  },
  {
    number: 2,
    isOpened: true,
    scenarioId: 2,
    scenarioTitle: 'Начатый сценарий',
    scenarioDescription: 'Продолжите попытку.',
    responseType: 'similar_choice',
    inProgressAttemptId: 22,
  },
  {
    number: 3,
    isOpened: true,
    scenarioId: 3,
    scenarioTitle: 'Пройденный сценарий',
    scenarioDescription: 'Улучшите результат.',
    responseType: 'mixed',
  },
  {
    number: 4,
    isOpened: false,
    scenarioId: 4,
    scenarioTitle: 'Закрытый сценарий',
    scenarioDescription: 'Сначала пройдите предыдущий.',
    responseType: 'free_text',
  },
]

describe('TrainingList', () => {
  it('distinguishes new, in-progress, completed and locked actions', async () => {
    const user = userEvent.setup()
    const onStart = vi.fn()
    render(
      <MemoryRouter>
        <TrainingList
          topic={topic}
          levels={levels}
          isStarting={false}
          onStart={onStart}
          basePath=""
        />
      </MemoryRouter>,
    )

    expect(screen.getByRole('button', { name: 'Начать' })).toBeEnabled()
    expect(screen.getByRole('button', { name: 'Продолжить' })).toBeEnabled()
    expect(screen.getByRole('button', { name: 'Попробовать снова' })).toBeEnabled()
    expect(screen.getByRole('button', { name: /Закрыто/ })).toBeDisabled()
    expect(screen.getByText(/Получите хотя бы одну Звезду за Уровень 3/)).toBeVisible()
    expect(screen.getAllByRole('link', { name: 'Последний Result →' })[0]).toHaveAttribute(
      'href',
      '/sessions/12/result',
    )

    await user.click(screen.getByRole('button', { name: 'Продолжить' }))
    expect(onStart).toHaveBeenCalledWith(2)
  })
})
