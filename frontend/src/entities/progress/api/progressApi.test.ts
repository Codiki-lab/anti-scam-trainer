// @vitest-environment node

import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import { createTestStore } from '@/test/createTestStore'
import { server } from '@/test/server'
import { progressApi } from './progressApi'

const completedDailyTask = {
  date: '2026-08-09',
  role: 'buyer',
  messages: [
    { role: 'assistant', text: 'Перейдите по ссылке для получения оплаты.' },
    { role: 'user', text: 'Не буду открывать внешние ссылки.' },
  ],
  completed: true,
  completed_at: '2026-08-09T10:00:00Z',
  answer: true,
  correct: true,
  verdict: true,
  signals: ['Внешняя ссылка для оплаты'],
  safe_action: 'Оставайтесь внутри сервиса.',
}

describe('daily task API boundary', () => {
  it('posts the binary answer and exposes a validated Result to the UI', async () => {
    server.use(
      http.post('http://localhost/api/v1/daily-tasks/answer', async ({ request }) => {
        expect(await request.json()).toEqual({ answer: true })
        return HttpResponse.json({
          daily_task: completedDailyTask,
          streak: { current: 4, longest: 4, active_today: true, last_activity_date: '2026-08-09' },
        })
      }),
    )

    const store = createTestStore()
    const result = await store
      .dispatch(progressApi.endpoints.answerDailyTask.initiate(true))
      .unwrap()

    expect(result).toMatchObject({
      dailyTask: {
        isCompleted: true,
        answer: true,
        isCorrect: true,
        verdict: true,
        safeAction: 'Оставайтесь внутри сервиса.',
      },
      streak: { current: 4, isActiveToday: true },
    })
  })

  it('keeps the backend conflict available to the caller', async () => {
    server.use(
      http.post('http://localhost/api/v1/daily-tasks/answer', () =>
        HttpResponse.json(
          {
            error: {
              code: 'STATE_CONFLICT',
              message: 'daily task is already answered',
              details: {},
              request_id: 'request-1',
            },
          },
          { status: 409 },
        ),
      ),
    )

    const store = createTestStore()
    await expect(
      store.dispatch(progressApi.endpoints.answerDailyTask.initiate(false)).unwrap(),
    ).rejects.toMatchObject({ status: 409 })
  })
})
