import { useGetDashboardQuery, type Dashboard } from '@/entities/progress'
import type { UserRole } from '@/entities/user'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

export function useDashboardData(role: UserRole, previewDashboard?: Dashboard) {
  const isPreview = useIsPreview()
  const query = useGetDashboardQuery(role, { skip: isPreview })

  return {
    dashboard: isPreview ? previewDashboard : query.data,
    isLoading: !isPreview && query.isLoading,
    error: isPreview || !query.error ? '' : getApiErrorMessage(query.error),
    retry: query.refetch,
  }
}
