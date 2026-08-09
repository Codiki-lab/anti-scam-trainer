import type { Topic } from '@/entities/learning'
import { useCurrentAccount } from '@/entities/user'
import { useTopics } from '@/features/learning-content'
import { ErrorState } from '@/shared/error-state'
import { useIsPreview } from '@/shared/runtime-mode'
import { uiStyles } from '@/shared/ui-kit'
import { TopicList } from '@/widgets/topic-list'

export function LessonsPage({ previewTopics }: { previewTopics?: Topic[] }) {
  const { account } = useCurrentAccount()
  const role = account.trainingRole
  const isPreview = useIsPreview()
  const { topics, isLoading, error, retry } = useTopics(role, previewTopics)

  if (isLoading) return <p className={uiStyles.muted}>Загружаем Темы…</p>
  if (error) return <ErrorState message={error} onRetry={() => void retry()} />

  return (
    <>
      <section className={uiStyles.pageHeading}>
        <p className={uiStyles.eyebrow}>Обучение</p>
        <h1>Темы для {role === 'buyer' ? 'покупателя' : 'продавца'}</h1>
        <p className={uiStyles.muted}>Прочитайте Теорию и закрепите её коротким Quiz.</p>
      </section>
      <TopicList topics={topics} basePath={isPreview ? '/preview/lessons' : '/lessons'} />
    </>
  )
}
