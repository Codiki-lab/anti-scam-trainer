import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { Achievement } from '@/entities/progress'
import { AchievementCard } from './AchievementCard'

describe('AchievementCard', () => {
  it('renders backend copy and exposes progress to assistive technology', () => {
    const achievement: Achievement = {
      code: 'five_trainings',
      title: 'Пять прохождений',
      description: 'Завершить пять Прохождений.',
      icon: 'stack',
      earned: false,
      current: 3,
      target: 5,
    }

    render(<AchievementCard achievement={achievement} />)

    expect(screen.getByRole('heading', { name: achievement.title })).toBeInTheDocument()
    expect(screen.getByText(achievement.description)).toBeInTheDocument()
    expect(screen.getByText('3 из 5')).toBeInTheDocument()
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '3')
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuemax', '5')
  })

  it('uses earned_at from the API when the achievement is earned', () => {
    const achievement: Achievement = {
      code: 'first_training',
      title: 'Первое прохождение',
      description: 'Завершить первое Прохождение.',
      icon: 'star',
      earned: true,
      earnedAt: '2026-08-08T10:00:00Z',
      current: 1,
      target: 1,
    }

    render(<AchievementCard achievement={achievement} />)

    expect(screen.getByText('Получено 8 августа')).toBeInTheDocument()
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
  })
})
