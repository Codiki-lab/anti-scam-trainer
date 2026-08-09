import { useNavigate, useParams } from 'react-router-dom'
import type { Quiz, QuizOutcome } from '@/entities/learning'
import { useQuiz, useSubmitQuizAttempt } from '@/features/learning-content'
import { ErrorState, InvalidRouteState } from '@/shared/error-state'
import { useIsPreview } from '@/shared/runtime-mode'
import { parsePositiveInteger } from '@/shared/url'
import { uiStyles } from '@/shared/ui-kit'
import { QuizPanel } from '@/widgets/quiz-panel'

interface QuizPageProps {
  previewQuiz?: Quiz
  previewOutcome?: QuizOutcome
  previewTopicId?: number
}

export function QuizPage({ previewQuiz, previewOutcome, previewTopicId }: QuizPageProps) {
  const { lessonId } = useParams()
  const navigate = useNavigate()
  const isPreview = useIsPreview()
  const parsedTopicId = parsePositiveInteger(lessonId)
  const topicId = previewTopicId ?? parsedTopicId ?? 0
  const { quiz, isLoading, error, retry } = useQuiz(topicId, previewQuiz)
  const { submit, isSubmitting } = useSubmitQuizAttempt(topicId, previewOutcome)
  const basePath = isPreview ? '/preview/lessons' : '/lessons'

  if (!isPreview && !parsedTopicId) {
    return <InvalidRouteState backTo={basePath} backLabel="К темам" />
  }
  if (isLoading) return <p className={uiStyles.muted}>Загружаем Quiz…</p>
  if (error) return <ErrorState message={error} onRetry={() => void retry()} />
  if (!quiz) return <p className={uiStyles.formError}>Quiz недоступен.</p>

  return (
    <QuizPanel
      quiz={quiz}
      isSubmitting={isSubmitting}
      onSubmit={submit}
      onPassed={() => navigate(isPreview ? '/preview/chats' : '/chats')}
    />
  )
}
