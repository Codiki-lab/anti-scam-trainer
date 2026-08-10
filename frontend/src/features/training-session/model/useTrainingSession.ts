import { useEffect, useState } from 'react'
import {
  useAbandonAttemptMutation,
  useGetAttemptQuery,
  useSubmitAnswerMutation,
  isAttemptResult,
} from '@/entities/training'
import type { AttemptResult, TrainingAnswer, TrainingSession } from '@/entities/training'
import { getApiErrorCode, getApiErrorDetails, getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

interface TrainingSessionState {
  session?: TrainingSession
  result?: AttemptResult
  error: string
  isLoading: boolean
  isSubmitting: boolean
  cooldown: number
  submit: (answer: TrainingAnswer) => Promise<boolean>
  abandon: () => Promise<boolean>
}

export function useTrainingSession(
  attemptId: number,
  preview?: { session: TrainingSession; result: AttemptResult },
): TrainingSessionState {
  const isPreview = useIsPreview()
  const query = useGetAttemptQuery(attemptId, { skip: isPreview || attemptId < 1 })
  const [submitAnswer, submitState] = useSubmitAnswerMutation()
  const [abandonAttempt, abandonState] = useAbandonAttemptMutation()
  const [session, setSession] = useState<TrainingSession | undefined>(
    isPreview ? preview?.session : undefined,
  )
  const [result, setResult] = useState<AttemptResult | undefined>()
  const [error, setError] = useState('')
  const [cooldown, setCooldown] = useState(0)

  useEffect(() => {
    if (query.data) setSession(query.data)
  }, [query.data])

  useEffect(() => {
    if (cooldown < 1) return
    const timer = window.setTimeout(() => setCooldown((value) => Math.max(0, value - 1)), 1000)
    return () => window.clearTimeout(timer)
  }, [cooldown])

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
      if (getApiErrorCode(requestError) === 'STALE_STEP') {
        await query.refetch()
        setError('Состояние тренировки обновилось. Проверьте новый шаг и ответьте ещё раз.')
        return false
      }
      if (getApiErrorCode(requestError) === 'RATE_LIMITED') {
        const retry = getApiErrorDetails(requestError).retry_after_seconds
        const seconds = typeof retry === 'number' ? retry : 1
        setCooldown(seconds)
        setError(`Лимит запросов. Повторите через ${seconds} сек.`)
      } else {
        setError(getApiErrorMessage(requestError))
      }
      return false
    }
  }

  const abandon = async () => {
    if (isPreview) return true
    try {
      await abandonAttempt(attemptId).unwrap()
      return true
    } catch (requestError) {
      setError(getApiErrorMessage(requestError))
      return false
    }
  }

  return {
    session,
    result,
    error: error || (!isPreview && query.error ? getApiErrorMessage(query.error) : ''),
    isLoading: isPreview ? false : query.isLoading,
    isSubmitting: submitState.isLoading || abandonState.isLoading,
    cooldown,
    submit,
    abandon,
  }
}
