import { useParams } from 'react-router-dom'
import { SkillCheckPanel } from '@/features/skill-check'
import type { SkillCheck } from '@/entities/learning'
import { InvalidRouteState } from '@/shared/error-state'
import { useIsPreview } from '@/shared/runtime-mode'
import { parsePositiveInteger } from '@/shared/url'

export function SkillCheckPage({ preview }: { preview?: SkillCheck }) {
  const { lessonId } = useParams()
  const isPreview = useIsPreview()
  const topicId = preview?.topicId ?? parsePositiveInteger(lessonId)
  const basePath = isPreview ? '/preview/lessons' : '/lessons'

  if (!topicId) return <InvalidRouteState backTo={basePath} backLabel="К Темам" />

  return <SkillCheckPanel topicId={topicId} basePath={basePath} preview={preview} />
}
