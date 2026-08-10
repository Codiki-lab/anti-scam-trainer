import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { Theory } from '@/entities/learning'
import { TheoryContent } from './TheoryContent'

const theory: Theory = {
  topic: {
    id: 1,
    slug: 'phishing',
    role: 'buyer',
    title: 'Фишинговые ссылки',
    description: 'Как распознать поддельную страницу.',
    order: 1,
    isTheoryRead: false,
    isQuizPassed: false,
    bestQuizScore: 0,
    isCompleted: false,
    levels: [],
  },
  sections: [
    { id: 1, order: 1, kind: 'intro', title: 'Введение', body: 'Проверяйте адрес страницы.' },
    { id: 2, order: 2, kind: 'risk', title: 'Риски', body: 'Чужой домен. Срочность.' },
    { id: 3, order: 3, kind: 'example', title: 'Пример', body: 'Оплатите по этой ссылке.' },
    {
      id: 4,
      order: 4,
      kind: 'safe_action',
      title: 'Что делать',
      body: 'Закройте страницу. Откройте приложение.',
    },
    { id: 5, order: 5, kind: 'summary', title: 'Итог', body: 'Оставайтесь внутри сервиса.' },
  ],
}

describe('TheoryContent', () => {
  it('renders all learning block kinds and completes explicitly', async () => {
    const user = userEvent.setup()
    const onFinish = vi.fn().mockResolvedValue(undefined)
    render(<TheoryContent theory={theory} onFinish={onFinish} />)

    expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(5)
    expect(screen.getByText('Чужой домен.')).toBeVisible()
    expect(screen.getByText('Оплатите по этой ссылке.').closest('blockquote')).not.toBeNull()
    await user.click(screen.getByRole('button', { name: 'Проверить знания' }))
    expect(onFinish).toHaveBeenCalledOnce()
  })

  it('shows a pending completion state', () => {
    render(<TheoryContent theory={theory} isSaving onFinish={vi.fn()} />)
    expect(screen.getByRole('button', { name: 'Сохраняем…' })).toBeDisabled()
  })
})
