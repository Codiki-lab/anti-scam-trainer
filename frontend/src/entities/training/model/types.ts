import type { LevelProgress } from '../../learning'
import type { Streak } from '../../user'

export type ResponseMode = 'multiple_choice' | 'similar_choice' | 'mixed' | 'free_text'
export type AttemptStatus = 'IN_PROGRESS' | 'COMPLETED' | 'ABANDONED'

export interface LevelState {
  number: number
  isOpened: boolean
  scenarioId: number
  scenarioTitle: string
  scenarioDescription: string
  responseMode: ResponseMode
  inProgressAttemptId?: number
}

export interface TrainingMessage {
  role: 'user' | 'assistant'
  text: string
}

export interface TrainingOption {
  id: number
  text: string
}

export interface TrainingSession {
  attemptId: number
  status: AttemptStatus
  scenarioId: number
  topicId: number
  productContext: Record<string, unknown>
  mode: ResponseMode
  progress: { currentStep: number; answeredSteps: number }
  step: {
    id: number
    number: number
    counterpartyMessage: string
    options: TrainingOption[]
  }
  answers: Array<{ stepId: number; optionId: number }>
  messages: TrainingMessage[]
  canFinishEarly: boolean
}

export type TrainingAnswer =
  | { type: 'option'; stepId: number; optionId: number }
  | { type: 'text'; stepId: number; text: string; finish?: boolean }

export interface DecisionReview {
  stepId: number
  optionId?: number
  optionText?: string
  freeText?: string
  points: number
  explanation: string
  riskSignals: string[]
}

export interface AttemptResult {
  attemptId: number
  score: number
  stars: number
  decisionReview: DecisionReview[]
  riskSignals: string[]
  safeActions: string[]
  levelProgress: LevelProgress
  topicId: number
  isTopicCompleted: boolean
  nextAction: {
    type: 'resume_attempt' | 'read_theory' | 'take_quiz' | 'start_level' | 'start_free_play'
    topicId?: number
    level?: number
    attemptId?: number
  } | null
  newAchievements: Array<{
    code: string
    title: string
    description: string
    icon: string
  }>
  streak: Streak
  isScam?: boolean
}

export type AnswerResult = TrainingSession | AttemptResult

export function isAttemptResult(value: AnswerResult): value is AttemptResult {
  return 'score' in value
}
