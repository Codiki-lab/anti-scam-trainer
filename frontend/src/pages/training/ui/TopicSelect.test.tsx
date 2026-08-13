import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { Topic } from '@/entities/learning'
import { TopicSelect } from './TopicSelect'

const topics: Topic[] = [
  {
    id: 1,
    slug: 'phishing-links',
    role: 'buyer',
    title: 'Фишинговые ссылки',
    description: 'Описание',
    order: 1,
    isTheoryRead: true,
    isQuizPassed: true,
    bestQuizScore: 80,
    isCompleted: false,
    levels: [],
  },
  {
    id: 2,
    slug: 'prepayment',
    role: 'buyer',
    title: 'Предоплата',
    description: 'Описание',
    order: 2,
    isTheoryRead: false,
    isQuizPassed: false,
    bestQuizScore: 0,
    isCompleted: false,
    levels: [],
  },
]

describe('TopicSelect', () => {
  it('selects a topic and closes with Escape or an outside press', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<TopicSelect topics={topics} value={1} onChange={onChange} />)

    const trigger = screen.getByRole('button', { name: /тема тренировки.*фишинговые ссылки/i })

    await user.click(trigger)
    expect(screen.getByRole('listbox')).toBeInTheDocument()
    expect(screen.getByRole('option', { name: /фишинговые ссылки/i })).toHaveAttribute(
      'aria-selected',
      'true',
    )

    fireEvent.keyDown(document, { key: 'Escape' })
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()

    await user.click(trigger)
    await user.click(screen.getByRole('option', { name: /предоплата/i }))
    expect(onChange).toHaveBeenCalledWith(2)
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()

    await user.click(trigger)
    fireEvent.pointerDown(document.body)
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })
})
