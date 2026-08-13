import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { Theory } from '@/entities/learning'
import { TheoryContent } from './TheoryContent'

const theory: Theory = {
  topic: {
    id: 5,
    slug: 'sms-codes',
    role: 'seller',
    title: 'Коды из SMS',
    description: 'Не передавайте секретные коды собеседникам.',
    order: 5,
    isTheoryRead: false,
    isQuizPassed: false,
    bestQuizScore: 0,
    isCompleted: false,
    levels: [],
  },
  sections: [
    { id: 1, order: 1, kind: 'intro', title: 'Как устроена схема', body: 'Вводный текст API' },
    { id: 2, order: 2, kind: 'risk', title: 'Сигналы риска', body: 'Риск из API' },
    { id: 3, order: 3, kind: 'example', title: 'Ситуация', body: 'Пример из API' },
    {
      id: 4,
      order: 4,
      kind: 'safe_action',
      title: 'Безопасное действие',
      body: 'Действие из API',
    },
    { id: 5, order: 5, kind: 'summary', title: 'Памятка', body: 'Итог из API' },
  ],
}

describe('TheoryContent', () => {
  it('renders all five backend theory blocks without phishing-specific copy', () => {
    render(<TheoryContent theory={theory} onFinish={async () => undefined} />)

    for (const body of theory.sections.map((section) => section.body)) {
      expect(screen.getByText(body)).toBeInTheDocument()
    }
    expect(screen.queryByText('avito-pay.example/order')).not.toBeInTheDocument()
    expect(screen.queryByText('Внешняя ссылка')).not.toBeInTheDocument()
  })
})
