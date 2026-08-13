import { useEffect } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useTrainingSession } from '@/features/training-session'
import { InvalidRouteState } from '@/shared/error-state'
import { useIsPreview } from '@/shared/runtime-mode'
import { uiStyles } from '@/shared/ui-kit'
import { parsePositiveInteger } from '@/shared/url'
import { TrainingChat } from '@/widgets/training-chat'
import type { TrainingPreview } from '../model/types'

interface ChatTrainingPageProps {
  preview?: Pick<TrainingPreview, 'session' | 'result'>
}

export function ChatTrainingPage({ preview }: ChatTrainingPageProps) {
  const { sessionId } = useParams()
  const navigate = useNavigate()
  const isPreview = useIsPreview()
  const parsedAttemptId = parsePositiveInteger(sessionId)
  const attemptId = preview?.session.attemptId ?? parsedAttemptId ?? 0
  const state = useTrainingSession(attemptId, preview)

  useEffect(() => {
    if (state.result) navigate(`${isPreview ? '/preview' : ''}/sessions/${attemptId}/result`)
  }, [attemptId, isPreview, navigate, state.result])

  if (!isPreview && !parsedAttemptId) {
    return <InvalidRouteState backTo="/chats" backLabel="К тренировкам" />
  }
  if (state.isLoading) return <p className={uiStyles.muted}>Восстанавливаем Прохождение…</p>
  if (!state.session) {
    return (
      <section className={uiStyles.pageHeading}>
        <h1>Прохождение не найдено</h1>
        <p className={uiStyles.muted}>Вернитесь к списку Уровней и начните тренировку.</p>
        <Link className={uiStyles.primaryButton} to={`${isPreview ? '/preview' : ''}/chats`}>
          К тренировкам
        </Link>
      </section>
    )
  }

  return (
    <TrainingChat
      session={state.session}
      isSubmitting={state.isSubmitting}
      error={state.error}
      onSubmit={state.submit}
    />
  )
}
