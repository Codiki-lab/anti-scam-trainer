import { useNavigate } from 'react-router-dom'
import { useStartLevelMutation } from '@/entities/training'
import type { UserRole } from '@/entities/user'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

export function useStartTraining(topicId: number, role: UserRole) {
  const navigate = useNavigate()
  const isPreview = useIsPreview()
  const [startLevel, state] = useStartLevelMutation()

  const start = async (level: number): Promise<string | undefined> => {
    if (isPreview) {
      navigate('/preview/sessions/demo')
      return undefined
    }

    try {
      const session = await startLevel({ role, topicId, level }).unwrap()
      navigate(`/sessions/${session.attemptId}`)
      return undefined
    } catch (error) {
      return getApiErrorMessage(error)
    }
  }

  return { start, isStarting: state.isLoading }
}
