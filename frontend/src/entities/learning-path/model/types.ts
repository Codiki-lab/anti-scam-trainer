export type ContinueActionType =
  'resume_attempt' | 'read_theory' | 'take_quiz' | 'start_level' | 'start_free_play'

export interface ContinueAction {
  type: ContinueActionType
  topicId?: number
  level?: number
  attemptId?: number
}
