import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { Quiz, QuizOutcome } from '@/entities/learning'
import { QuizPanel } from './QuizPanel'

const quiz: Quiz = {
  passThreshold: 80,
  questions: [
    {
      id: 1,
      order: 1,
      text: 'Где безопасно общаться?',
      choices: [
        { id: 11, text: 'Внутри сервиса' },
        { id: 12, text: 'В мессенджере' },
      ],
    },
    {
      id: 2,
      order: 2,
      text: 'Кому сообщать код?',
      choices: [
        { id: 21, text: 'Никому' },
        { id: 22, text: 'Покупателю' },
      ],
    },
  ],
}

const passed: QuizOutcome = {
  score: 100,
  isPassed: true,
  bestScore: 100,
  isFirstPass: true,
  streak: { current: 2, longest: 2, isActiveToday: true },
}

describe('QuizPanel', () => {
  it('collects structured answers and renders a passing result', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn().mockResolvedValue(passed)
    render(<QuizPanel quiz={quiz} isSubmitting={false} onSubmit={onSubmit} onPassed={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: /Внутри сервиса/ }))
    await user.click(screen.getByRole('button', { name: 'Ответить' }))
    await user.click(screen.getByRole('button', { name: /Никому/ }))
    await user.click(screen.getByRole('button', { name: 'Завершить' }))

    expect(onSubmit).toHaveBeenCalledWith({
      answers: [
        { questionId: 1, choiceId: 11 },
        { questionId: 2, choiceId: 21 },
      ],
    })
    expect(await screen.findByRole('heading', { name: '100%' })).toBeVisible()
    expect(screen.getByRole('button', { name: 'К тренировкам' })).toBeVisible()
  })

  it('keeps the final question recoverable after a submission failure', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn().mockRejectedValue(new Error('offline'))
    render(
      <QuizPanel
        quiz={{ ...quiz, questions: [quiz.questions[0]] }}
        isSubmitting={false}
        onSubmit={onSubmit}
        onPassed={vi.fn()}
      />,
    )

    await user.click(screen.getByRole('button', { name: /Внутри сервиса/ }))
    await user.click(screen.getByRole('button', { name: 'Завершить' }))
    expect(
      await screen.findByText('Не удалось отправить ответы. Попробуйте ещё раз.'),
    ).toBeVisible()
    expect(screen.getByRole('button', { name: /Внутри сервиса/ })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
  })
})
