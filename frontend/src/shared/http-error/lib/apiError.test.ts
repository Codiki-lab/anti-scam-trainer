import { describe, expect, it } from 'vitest'
import { getApiErrorCode, getApiErrorMessage } from './apiError'

const backendError = {
  status: 409,
  data: {
    error: {
      code: 'STALE_STEP',
      message: 'Шаг уже изменился',
      details: {},
      request_id: 'request-1',
    },
  },
}

describe('API error adapter', () => {
  it('extracts a validated backend error', () => {
    expect(getApiErrorCode(backendError)).toBe('STALE_STEP')
    expect(getApiErrorMessage(backendError)).toBe('Шаг уже изменился')
  })

  it('does not trust an incomplete envelope', () => {
    expect(getApiErrorMessage({ status: 500, data: { error: { message: 'unsafe' } } })).toBe(
      'Не удалось выполнить запрос. Попробуйте ещё раз.',
    )
  })
})
