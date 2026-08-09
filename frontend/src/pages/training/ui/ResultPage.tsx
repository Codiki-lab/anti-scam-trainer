import { useParams } from 'react-router-dom'
import { useGetAttemptResultQuery, type AttemptResult } from '@/entities/training'
import { ErrorState, InvalidRouteState } from '@/shared/error-state'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'
import { uiStyles } from '@/shared/ui-kit'
import { parsePositiveInteger } from '@/shared/url'
import { ResultSummary } from '@/widgets/result-summary'
import { getNextActionPath } from '../lib/getNextActionPath'

export function ResultPage({ previewResult }: { previewResult?: AttemptResult }) {
  const { sessionId } = useParams()
  const isPreview = useIsPreview()
  const parsedAttemptId = parsePositiveInteger(sessionId)
  const attemptId = previewResult?.attemptId ?? parsedAttemptId ?? 0
  const query = useGetAttemptResultQuery(attemptId, { skip: isPreview || attemptId < 1 })
  const result = isPreview ? previewResult : query.data

  if (!isPreview && !parsedAttemptId) {
    return <InvalidRouteState backTo="/chats" backLabel="К тренировкам" />
  }
  if (query.isLoading) return <p className={uiStyles.muted}>Загружаем Result…</p>
  if (query.error) {
    return (
      <ErrorState message={getApiErrorMessage(query.error)} onRetry={() => void query.refetch()} />
    )
  }
  if (!result) return <p className={uiStyles.formError}>Result пока недоступен.</p>

  const basePath = isPreview ? '/preview' : ''
  return (
    <ResultSummary
      result={result}
      basePath={basePath}
      nextActionHref={getNextActionPath(result, basePath)}
    />
  )
}
