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
  submit: (answer: TrainingAnswer) => Promise<void>
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
      return
    }

    try {
      const response = await submitAnswer({ attemptId, answer }).unwrap()
      if (isAttemptResult(response)) setResult(response)
      else setSession(response)
    } catch (requestError) {
      if (getApiErrorCode(requestError) === 'STALE_STEP') {
        await query.refetch()
        setError('Состояние тренировки обновилось. Проверьте новый шаг и ответьте ещё раз.')
        return
      }
      setError(getApiErrorMessage(requestError))
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
