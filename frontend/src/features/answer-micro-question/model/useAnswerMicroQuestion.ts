import { useState } from 'react'
import { useAnswerMicroQuestionMutation, type MicroQuestionAnswer } from '@/entities/training'
import { getApiErrorCode, getApiErrorMessage } from '@/shared/http-error'

interface AnswerMicroQuestionState {
  answer?: MicroQuestionAnswer
  error: string
  isSubmitting: boolean
  submit: (answerIndex: 0 | 1) => Promise<void>
}

export function useAnswerMicroQuestion(
  attemptId: number,
  onConflict: () => void,
): AnswerMicroQuestionState {
  const [answerMicroQuestion, request] = useAnswerMicroQuestionMutation()
  const [answer, setAnswer] = useState<MicroQuestionAnswer>()
  const [error, setError] = useState('')

  const submit = async (answerIndex: 0 | 1) => {
    setError('')
    try {
      setAnswer(await answerMicroQuestion({ attemptId, answerIndex }).unwrap())
    } catch (requestError) {
      if (getApiErrorCode(requestError) === 'STATE_CONFLICT') {
        onConflict()
        return
      }
      setError(getApiErrorMessage(requestError))
    }
  }

  return { answer, error, isSubmitting: request.isLoading, submit }
}
