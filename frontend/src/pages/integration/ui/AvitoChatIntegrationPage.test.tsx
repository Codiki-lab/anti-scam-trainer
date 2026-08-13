import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AvitoChatIntegrationPage } from './AvitoChatIntegrationPage'

const submit = vi.fn()

vi.mock('@/entities/user', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/user')>()

  return {
    ...actual,
    useCurrentAccount: () => ({
      account: {
        id: 1,
        username: 'alexey',
        accessRole: 'user' as const,
        trainingRole: 'buyer' as const,
        streak: { current: 2, longest: 2, isActiveToday: true },
      },
    }),
  }
})

vi.mock('@/features/avito-chat-recommendation', () => ({
  useAvitoChatRecommendation: () => ({
    snapshot: [
      { role: 'assistant', text: 'Велосипед ещё продаётся?' },
      { role: 'user', text: 'Да, ещё актуален.' },
      { role: 'assistant', text: 'Подтвердите получение на внешней странице.' },
    ],
    submit,
    isLoading: false,
    error: '',
    recommendation: undefined,
  }),
}))

describe('AvitoChatIntegrationPage', () => {
  beforeEach(() => submit.mockReset())

  it('opens consent from the system warning and sends only after confirmation', async () => {
    const user = userEvent.setup()

    render(
      <MemoryRouter>
        <AvitoChatIntegrationPage />
      </MemoryRouter>,
    )

    expect(screen.getByText('Собеседник кажется подозрительным')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: /Пройти антискам-тренажёр/ }))

    expect(
      screen.getByRole('dialog', { name: 'Разобрать ситуацию в тренажёре?' }),
    ).toBeInTheDocument()
    expect(submit).not.toHaveBeenCalled()

    await user.click(screen.getByRole('button', { name: 'Продолжить' }))
    expect(submit).toHaveBeenCalledOnce()
  })
})
