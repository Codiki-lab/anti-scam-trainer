import { useEffect, useState } from 'react'
import { useGetAttemptQuery, useSubmitAnswerMutation, isAttemptResult } from '@/entities/training'
import type { AttemptResult, TrainingAnswer, TrainingSession } from '@/entities/training'
import { getApiErrorCode, getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

interface TrainingSessionState {
  session?: TrainingSession
  result?: AttemptResult
  error: string
  isLoading: boolean
  isSubmitting: boolean
  submit: (answer: TrainingAnswer) => Promise<boolean>
}

export function useTrainingSession(
  attemptId: number,
  preview?: { session: TrainingSession; result: AttemptResult },
): TrainingSessionState {
  const isPreview = useIsPreview()
  const query = useGetAttemptQuery(attemptId, { skip: isPreview || attemptId < 1 })
  const [submitAnswer, submitState] = useSubmitAnswerMutation()
  const [session, setSession] = useState<TrainingSession | undefined>(
    isPreview ? preview?.session : undefined,
  )
  const [result, setResult] = useState<AttemptResult | undefined>()
  const [error, setError] = useState('')

  useEffect(() => {
    if (query.data) setSession(query.data)
  }, [query.data])

  const submit = async (answer: TrainingAnswer) => {
    setError('')

    if (isPreview) {
      if (!preview) throw new Error('Training preview data is missing')
      setResult(preview.result)
      return true
    }

    try {
      const response = await submitAnswer({ attemptId, answer }).unwrap()
      if (isAttemptResult(response)) setResult(response)
      else setSession(response)
      return true
    } catch (requestError) {
      const code = getApiErrorCode(requestError)
      if (code === 'STALE_STEP') {
        await query.refetch()
        setError('Состояние тренировки обновилось. Проверьте новый шаг и ответьте ещё раз.')
        return false
      }
      if (code === 'RATE_LIMITED') {
        setError('Модель обрабатывает много запросов. Подождите немного и повторите отправку.')
      } else if (code === 'AI_UNAVAILABLE' || code === 'AI_INVALID_RESPONSE') {
        setError('Не удалось проверить ответ. Он сохранён в поле — повторите отправку.')
      } else {
        setError(getApiErrorMessage(requestError))
      }
      return false
    }
  }

  return {
    session,
    result,
    error: error || (!isPreview && query.error ? getApiErrorMessage(query.error) : ''),
    isLoading: isPreview ? false : query.isLoading,
    isSubmitting: submitState.isLoading,
    submit,
  }
}
