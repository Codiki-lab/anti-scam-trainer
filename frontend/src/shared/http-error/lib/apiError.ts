import type { FetchBaseQueryError } from '@reduxjs/toolkit/query'
import { isApiErrorEnvelope, type ApiErrorCode } from '../model/contracts'

export function getApiErrorCode(error: unknown): ApiErrorCode | undefined {
  if (!error || typeof error !== 'object' || !('data' in error)) return undefined

  const data = (error as FetchBaseQueryError).data

  return isApiErrorEnvelope(data) ? data.error.code : undefined
}

export function getApiErrorMessage(error: unknown): string {
  if (error && typeof error === 'object' && 'data' in error) {
    const data = (error as FetchBaseQueryError).data
    if (isApiErrorEnvelope(data)) return data.error.message
  }

  if (error instanceof Error) return error.message

  return 'Не удалось выполнить запрос. Попробуйте ещё раз.'
}
