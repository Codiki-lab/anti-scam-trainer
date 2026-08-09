import { z } from 'zod'

export type ApiErrorCode =
  | 'VALIDATION_ERROR'
  | 'AUTHENTICATION_REQUIRED'
  | 'FORBIDDEN'
  | 'CONTENT_UNAVAILABLE'
  | 'NOT_FOUND'
  | 'STATE_CONFLICT'
  | 'STALE_STEP'
  | 'AI_INVALID_RESPONSE'
  | 'AI_UNAVAILABLE'
  | 'METHOD_NOT_ALLOWED'
  | 'INTERNAL_ERROR'

export interface ApiErrorEnvelope {
  error: {
    code: ApiErrorCode
    message: string
    details: Record<string, unknown>
    request_id: string
  }
}

export const apiErrorEnvelopeSchema = z.object({
  error: z.object({
    code: z.enum([
      'VALIDATION_ERROR',
      'AUTHENTICATION_REQUIRED',
      'FORBIDDEN',
      'CONTENT_UNAVAILABLE',
      'NOT_FOUND',
      'STATE_CONFLICT',
      'STALE_STEP',
      'AI_INVALID_RESPONSE',
      'AI_UNAVAILABLE',
      'METHOD_NOT_ALLOWED',
      'INTERNAL_ERROR',
    ]),
    message: z.string(),
    details: z.record(z.unknown()),
    request_id: z.string(),
  }),
})

export function isApiErrorEnvelope(value: unknown): value is ApiErrorEnvelope {
  return apiErrorEnvelopeSchema.safeParse(value).success
}
