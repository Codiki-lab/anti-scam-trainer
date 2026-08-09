import type { Topic } from '@/entities/learning'
import type { AttemptResult, LevelState, TrainingSession } from '@/entities/training'

export interface TrainingPreview {
  topics: Topic[]
  levels: LevelState[]
  session: TrainingSession
  result: AttemptResult
}
