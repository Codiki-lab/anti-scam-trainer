import { act, renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useTrainingSession } from './useTrainingSession'

const mocks = vi.hoisted(() => ({
  refetch: vi.fn(),
  submit: vi.fn(),
  abandon: vi.fn(),
}))

vi.mock('@/shared/runtime-mode', () => ({ useIsPreview: () => false }))
vi.mock('@/entities/training', () => ({
  useGetAttemptQuery: () => ({
    data: undefined,
    error: undefined,
    isLoading: false,
    refetch: mocks.refetch,
  }),
  useSubmitAnswerMutation: () => [mocks.submit, { isLoading: false }],
  useAbandonAttemptMutation: () => [mocks.abandon, { isLoading: false }],
  isAttemptResult: (value: unknown) =>
    Boolean(value && typeof value === 'object' && 'score' in value),
}))

const apiError = (code: 'STALE_STEP' | 'RATE_LIMITED', details: Record<string, unknown> = {}) => ({
  status: code === 'STALE_STEP' ? 409 : 429,
  data: { error: { code, message: code, details, request_id: 'test-request' } },
})

describe('useTrainingSession recovery', () => {
  beforeEach(() => {
    mocks.refetch.mockReset().mockResolvedValue(undefined)
    mocks.submit.mockReset()
    mocks.abandon.mockReset()
  })

  it('refetches the authoritative step after a stale submission', async () => {
    mocks.submit.mockReturnValue({ unwrap: () => Promise.reject(apiError('STALE_STEP')) })
    const { result } = renderHook(() => useTrainingSession(7))

    await act(async () => {
      expect(await result.current.submit({ type: 'option', stepId: 4, optionId: 11 })).toBe(false)
    })

    expect(mocks.refetch).toHaveBeenCalledOnce()
    expect(result.current.error).toMatch(/Состояние тренировки обновилось/)
  })

  it('exposes the server retry window after rate limiting', async () => {
    mocks.submit.mockReturnValue({
      unwrap: () => Promise.reject(apiError('RATE_LIMITED', { retry_after_seconds: 5 })),
    })
    const { result } = renderHook(() => useTrainingSession(7))

    await act(async () => {
      await result.current.submit({ type: 'text', stepId: 4, text: 'Останусь в сервисе' })
    })

    expect(result.current.cooldown).toBe(5)
    expect(result.current.error).toBe('Лимит запросов. Повторите через 5 сек.')
  })
})
