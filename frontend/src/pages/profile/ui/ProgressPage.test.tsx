import { render, screen } from '@testing-library/react'
import { Provider } from 'react-redux'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import type { Progress } from '@/entities/progress'
import { CurrentAccountProvider, type Account } from '@/entities/user'
import { PreviewModeProvider } from '@/shared/runtime-mode'
import { createTestStore } from '@/test/createTestStore'
import { ProgressPage } from './ProgressPage'

const account: Account = {
  id: 1,
  username: 'demo',
  accessRole: 'user',
  trainingRole: 'buyer',
  streak: { current: 3, longest: 3, isActiveToday: true },
}

const progress: Progress = {
  role: 'buyer',
  summary: {
    completedTopics: 1,
    totalTopics: 6,
    completedLevels: 4,
    totalLevels: 24,
    stars: 8,
    averageScore: 76,
  },
  topics: [
    {
      id: 11,
      slug: 'phishing-links',
      role: 'buyer',
      title: 'Фишинговые ссылки',
      description: 'Описание Темы',
      order: 1,
      isTheoryRead: true,
      isQuizPassed: true,
      bestQuizScore: 80,
      isCompleted: true,
      levels: [1, 2, 3, 4].map((number) => ({
        number,
        isOpened: true,
        bestScore: 75,
        stars: 2,
        attempts: 1,
        lastAttemptId: 100 + number,
      })),
    },
  ],
  recentAttempts: [
    {
      attemptId: 101,
      topicId: 11,
      level: 2,
      score: 75,
      stars: 2,
      finishedAt: '2026-08-09T09:00:00Z',
    },
  ],
}

describe('ProgressPage', () => {
  it('renders summary, per-topic progress and recent attempts from the API model', () => {
    render(
      <Provider store={createTestStore()}>
        <CurrentAccountProvider value={{ account, changeTrainingRole: async () => undefined }}>
          <PreviewModeProvider>
            <MemoryRouter>
              <ProgressPage previewProgress={progress} />
            </MemoryRouter>
          </PreviewModeProvider>
        </CurrentAccountProvider>
      </Provider>,
    )

    expect(screen.getByRole('heading', { name: '17%' })).toBeInTheDocument()
    expect(screen.getByText('4/24')).toBeInTheDocument()
    expect(screen.getAllByText('Фишинговые ссылки')).toHaveLength(2)
    expect(screen.getByRole('link', { name: /фишинговые ссылки/i })).toHaveAttribute(
      'href',
      '/preview/lessons/11',
    )
    expect(screen.getByLabelText('2 из 3 Звёзд')).toBeInTheDocument()
    expect(screen.getByText('75%')).toBeInTheDocument()
  })
})
