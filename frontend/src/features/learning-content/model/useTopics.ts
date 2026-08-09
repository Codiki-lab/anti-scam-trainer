import { useGetTopicsQuery, type Topic } from '@/entities/learning'
import type { UserRole } from '@/entities/user'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

export function useTopics(role: UserRole, previewTopics: Topic[] = []) {
  const isPreview = useIsPreview()
  const query = useGetTopicsQuery(role, { skip: isPreview })

  return {
    topics: isPreview ? previewTopics : (query.data ?? []),
    isLoading: isPreview ? false : query.isLoading,
    error: isPreview || !query.error ? '' : getApiErrorMessage(query.error),
    retry: query.refetch,
  }
}
