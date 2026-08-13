import type { ContinueActionDto } from '../api/contracts'
import type { ContinueAction } from '../model/types'

export function mapContinueAction(
  dto: ContinueActionDto | null | undefined,
): ContinueAction | null {
  if (!dto) return null

  return {
    type: dto.type,
    topicId: dto.topic_id,
    level: dto.level,
    attemptId: dto.attempt_id,
  }
}
