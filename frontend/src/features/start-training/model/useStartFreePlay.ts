import { useNavigate } from 'react-router-dom'
import { useStartFreePlayMutation } from '@/entities/training'
import type { UserRole } from '@/entities/user'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

export function useStartFreePlay(role: UserRole) {
  const navigate = useNavigate()
  const isPreview = useIsPreview()
  const [startFreePlay, state] = useStartFreePlayMutation()

  const start = async (): Promise<string | undefined> => {
    if (isPreview) {
      navigate('/preview/sessions/free-play')
      return undefined
    }

    try {
      const session = await startFreePlay(role).unwrap()
      navigate(`/sessions/${session.attemptId}`)
      return undefined
    } catch (error) {
      return getApiErrorMessage(error)
    }
  }

  return { start, isStarting: state.isLoading }
}
