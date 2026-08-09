import { Link, useNavigate, useParams } from 'react-router-dom'
import type { Theory } from '@/entities/learning'
import { useFinishTheory, useTheory } from '@/features/learning-content'
import { ErrorState, InvalidRouteState } from '@/shared/error-state'
import { useIsPreview } from '@/shared/runtime-mode'
import { parsePositiveInteger } from '@/shared/url'
import { uiStyles } from '@/shared/ui-kit'
import { TheoryContent } from '@/widgets/theory-content'

export function TheoryPage({ previewTheory }: { previewTheory?: Theory }) {
  const { lessonId } = useParams()
  const navigate = useNavigate()
  const isPreview = useIsPreview()
  const parsedTopicId = parsePositiveInteger(lessonId)
  const topicId = previewTheory?.topic.id ?? parsedTopicId ?? 0
  const { theory, isLoading, error, retry } = useTheory(topicId, previewTheory)
  const { finishTheory, isSaving, error: finishError } = useFinishTheory(topicId)
  const basePath = isPreview ? '/preview/lessons' : '/lessons'

  if (!isPreview && !parsedTopicId) {
    return <InvalidRouteState backTo={basePath} backLabel="К темам" />
  }
  if (isLoading) return <p className={uiStyles.muted}>Загружаем Теорию…</p>
  if (error) return <ErrorState message={error} onRetry={() => void retry()} />
  if (!theory) return <p className={uiStyles.formError}>Теория недоступна.</p>

  return (
    <>
      <div className={uiStyles.breadcrumbs}>
        <Link to={basePath}>Обучение</Link> / {theory.topic.title}
      </div>
      <TheoryContent
        theory={theory}
        isSaving={isSaving}
        onFinish={async () => {
          if (await finishTheory()) navigate(`${basePath}/${topicId}/quiz`)
        }}
      />
      {finishError && (
        <p className={uiStyles.formError} role="alert">
          {finishError}
        </p>
      )}
    </>
  )
}
