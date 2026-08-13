import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { TrainingSession } from '@/entities/training'
import { TrainingChat } from './TrainingChat'

const freePlaySession: TrainingSession = {
  attemptId: 1,
  status: 'IN_PROGRESS',
  scenarioId: 0,
  scenarioTitle: 'Свободная игра',
  scenarioDescription: 'Диалог без подсказок',
  topicId: 0,
  topicTitle: 'Свободная игра',
  level: 4,
  userRole: 'buyer',
  counterpartyRole: 'seller',
  productContext: {
    itemTitle: 'Велосипед',
    category: 'Спорт',
    dealMethod: 'delivery',
  },
  mode: 'free_text',
  progress: { currentStep: 1, answeredSteps: 0, totalSteps: 4 },
  step: { id: 1, number: 1, counterpartyMessage: 'Товар актуален?', options: [] },
  answers: [],
  messages: [{ role: 'assistant', text: 'Товар актуален?' }],
  canFinishEarly: false,
}

describe('TrainingChat', () => {
  it('keeps free text after a rejected server submission', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn().mockResolvedValue(false)
    render(
      <TrainingChat session={freePlaySession} isSubmitting={false} error="" onSubmit={onSubmit} />,
    )

    const field = screen.getByPlaceholderText('Напишите безопасный ответ…')
    await user.type(field, 'Оформим всё внутри сервиса')
    await user.click(screen.getByRole('button', { name: 'Отправить' }))

    expect(field).toHaveValue('Оформим всё внутри сервиса')
  })

  it('clears free text only after the server accepts it', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn().mockResolvedValue(true)
    render(
      <TrainingChat session={freePlaySession} isSubmitting={false} error="" onSubmit={onSubmit} />,
    )

    const field = screen.getByPlaceholderText('Напишите безопасный ответ…')
    await user.type(field, 'Проверю заказ в приложении')
    await user.click(screen.getByRole('button', { name: 'Отправить' }))

    expect(field).toHaveValue('')
  })
})
