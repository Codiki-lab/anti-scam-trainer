import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { TrainingSession } from '@/entities/training'
import { TrainingChat } from './TrainingChat'

const session: TrainingSession = {
  attemptId: 7,
  status: 'IN_PROGRESS',
  scenarioId: 4,
  scenarioTitle: 'Проверка оплаты смартфона',
  scenarioDescription: 'Убедитесь, что деньги действительно поступили.',
  topicId: 2,
  topicTitle: 'Поддельная оплата',
  level: 2,
  userRole: 'seller',
  counterpartyRole: 'buyer',
  productContext: {
    itemTitle: 'Смартфон',
    category: 'Электроника',
    dealMethod: 'delivery',
    price: 42000,
    currency: 'RUB',
  },
  mode: 'multiple_choice',
  progress: { currentStep: 1, answeredSteps: 0, totalSteps: 2 },
  step: {
    id: 10,
    number: 1,
    counterpartyMessage: 'Я уже оплатил заказ.',
    options: [
      { id: 1, text: 'Проверю поступление самостоятельно' },
      { id: 2, text: 'Доверюсь скриншоту' },
    ],
  },
  answers: [],
  messages: [{ role: 'assistant', text: 'Я уже оплатил заказ.' }],
  canFinishEarly: false,
}

describe('TrainingChat', () => {
  it('separates option selection from confirmation', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn().mockResolvedValue(true)
    render(
      <TrainingChat
        session={session}
        isSubmitting={false}
        error=""
        cooldown={0}
        onSubmit={onSubmit}
        onAbandon={vi.fn()}
      />,
    )

    await user.click(screen.getByRole('button', { name: 'Проверю поступление самостоятельно' }))
    expect(onSubmit).not.toHaveBeenCalled()
    expect(
      screen.getByRole('button', { name: 'Проверю поступление самостоятельно' }),
    ).toHaveAttribute('aria-pressed', 'true')

    await user.click(screen.getByRole('button', { name: 'Подтвердить ответ' }))
    expect(onSubmit).toHaveBeenCalledWith({ type: 'option', stepId: 10, optionId: 1 })
  })

  it('keeps free text after a failed request', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn().mockResolvedValue(false)
    render(
      <TrainingChat
        session={{ ...session, mode: 'free_text', step: { ...session.step, options: [] } }}
        isSubmitting={false}
        error="AI временно недоступен"
        cooldown={0}
        onSubmit={onSubmit}
        onAbandon={vi.fn()}
      />,
    )

    const input = screen.getByPlaceholderText('Напишите безопасный ответ…')
    await user.type(input, 'Продолжим только внутри приложения')
    await user.click(screen.getByRole('button', { name: 'Отправить' }))
    expect(input).toHaveValue('Продолжим только внутри приложения')
  })

  it('restores option and free-text answers from structured history', () => {
    render(
      <TrainingChat
        session={{
          ...session,
          productContext: { ...session.productContext, imageKey: 'smartphone' },
          answers: [
            { stepId: 8, answerType: 'option', optionText: 'Проверю оплату в приложении', points: 100 },
            { stepId: 9, answerType: 'free_text', freeText: 'Не буду открывать ссылку', points: 100 },
          ],
          messages: [
            { role: 'assistant', text: 'Я уже оплатил заказ.' },
            { role: 'user', text: 'legacy text that must not be parsed' },
            { role: 'assistant', text: 'Тогда откройте ссылку.' },
          ],
        }}
        isSubmitting={false}
        error=""
        cooldown={0}
        onSubmit={vi.fn()}
        onAbandon={vi.fn()}
      />,
    )

    expect(screen.getByText('Проверю оплату в приложении')).toBeVisible()
    expect(screen.getByText('Не буду открывать ссылку')).toBeVisible()
    expect(screen.queryByText('legacy text that must not be parsed')).not.toBeInTheDocument()
  })
})
