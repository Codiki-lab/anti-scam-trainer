import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { DailyTask } from '@/entities/progress'
import { DailyTaskModal } from './DailyTaskModal'

const answerDailyTask = vi.fn()

vi.mock('@/entities/progress', () => ({
  useAnswerDailyTaskMutation: () => [answerDailyTask, { isLoading: false }],
}))

const pendingTask: DailyTask = {
  date: '2026-08-09',
  role: 'seller',
  messages: [
    { role: 'assistant', text: 'Подтвердите получение перевода по ссылке.' },
    { role: 'user', text: 'Ссылка похожа на сервис объявлений.' },
  ],
  isCompleted: false,
  signals: [],
}

function renderModal(task = pendingTask) {
  return render(<DailyTaskModal isOpen onClose={vi.fn()} onConflict={vi.fn()} task={task} />)
}

describe('DailyTaskModal', () => {
  it('submits one answer and shows the saved review', async () => {
    const user = userEvent.setup()
    answerDailyTask.mockReturnValue({
      unwrap: () =>
        Promise.resolve({
          dailyTask: {
            ...pendingTask,
            isCompleted: true,
            completedAt: '2026-08-09T10:00:00Z',
            answer: true,
            isCorrect: true,
            verdict: true,
            signals: ['Внешняя ссылка для подтверждения'],
            safeAction: 'Не переходите по ссылкам из сообщений.',
          },
          streak: { current: 2, longest: 2, isActiveToday: true },
        }),
    })

    renderModal()
    await user.click(screen.getByRole('button', { name: 'Это обман' }))

    expect(await screen.findByText(/Не переходите по ссылкам/)).toBeInTheDocument()
    expect(answerDailyTask).toHaveBeenCalledOnce()
    expect(answerDailyTask).toHaveBeenCalledWith(true)
    expect(screen.queryByRole('button', { name: 'Ситуация безопасна' })).not.toBeInTheDocument()
  })

  it('shows a stored review without allowing another answer', () => {
    renderModal({
      ...pendingTask,
      isCompleted: true,
      answer: false,
      isCorrect: true,
      verdict: false,
      signals: ['Нет просьбы уйти из сервиса'],
      safeAction: 'Продолжайте использовать защищённые инструменты сервиса.',
    })

    expect(screen.getByText('Выполнено')).toBeInTheDocument()
    expect(screen.getByText('Ситуация выглядит безопасной')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Это обман' })).not.toBeInTheDocument()
  })
})
