import {
  useGetQuizQuery,
  useSubmitQuizMutation,
  type Quiz,
  type QuizOutcome,
  type QuizSubmission,
} from '@/entities/learning'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

export function useQuiz(topicId: number, previewQuiz?: Quiz) {
  const isPreview = useIsPreview()
  const query = useGetQuizQuery(topicId, { skip: isPreview || topicId < 1 })

  return {
    quiz: isPreview ? previewQuiz : query.data,
    isLoading: isPreview ? false : query.isLoading,
    error: isPreview || !query.error ? '' : getApiErrorMessage(query.error),
    retry: query.refetch,
  }
}

export function useSubmitQuizAttempt(topicId: number, previewOutcome?: QuizOutcome) {
  const isPreview = useIsPreview()
  const [submitQuiz, state] = useSubmitQuizMutation()

  const submit = async (submission: QuizSubmission): Promise<QuizOutcome> => {
    if (isPreview) {
      if (!previewOutcome) throw new Error('Quiz preview outcome is missing')
      return previewOutcome
    }

    return submitQuiz({ topicId, submission }).unwrap()
  }

  return { submit, isSubmitting: !isPreview && state.isLoading }
}
