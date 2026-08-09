import { useGetAchievementsQuery, type Achievements } from '@/entities/progress'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

export function useAchievementData(previewAchievements?: Achievements) {
  const isPreview = useIsPreview()
  const query = useGetAchievementsQuery(undefined, { skip: isPreview })

  return {
    achievements: isPreview ? previewAchievements : query.data,
    isLoading: !isPreview && query.isLoading,
    error: isPreview || !query.error ? '' : getApiErrorMessage(query.error),
    retry: query.refetch,
  }
}
