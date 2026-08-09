import { useGetProgressQuery, type Progress } from '@/entities/progress'
import type { UserRole } from '@/entities/user'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

export function useProgressData(role: UserRole, previewProgress?: Progress) {
  const isPreview = useIsPreview()
  const query = useGetProgressQuery(role, { skip: isPreview })

  return {
    progress: isPreview ? previewProgress : query.data,
    isLoading: !isPreview && query.isLoading,
    error: isPreview || !query.error ? '' : getApiErrorMessage(query.error),
    retry: query.refetch,
  }
}
