import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import type { Topic } from '@/entities/learning'
import { TopicList } from './TopicList'

const topic: Topic = {
  id: 7,
  slug: 'phishing',
  role: 'buyer',
  title: 'Фишинговые ссылки',
  description: 'Распознавайте поддельные страницы оплаты и доставки.',
  order: 1,
  isTheoryRead: true,
  isQuizPassed: true,
  bestQuizScore: 80,
  isCompleted: false,
  levels: [
    { number: 1, isOpened: true, bestScore: 70, stars: 2, attempts: 1, lastAttemptId: 12 },
    { number: 2, isOpened: true, bestScore: 0, stars: 0, attempts: 0, lastAttemptId: null },
    { number: 3, isOpened: false, bestScore: 0, stars: 0, attempts: 0, lastAttemptId: null },
    { number: 4, isOpened: false, bestScore: 0, stars: 0, attempts: 0, lastAttemptId: null },
  ],
}

describe('TopicList', () => {
  it('shows backend progress and links to the selected Topic', () => {
    render(
      <MemoryRouter>
        <TopicList topics={[topic]} basePath="/preview/lessons" />
      </MemoryRouter>,
    )

    expect(screen.getByRole('link', { name: /Фишинговые ссылки/ })).toHaveAttribute(
      'href',
      '/preview/lessons/7',
    )
    expect(screen.getByText('Доступны Уровни')).toBeInTheDocument()
    expect(screen.getByText('3 из 6')).toBeInTheDocument()
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '3')
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuemax', '6')
  })
})
