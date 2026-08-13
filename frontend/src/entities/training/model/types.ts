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
  scenarioTitle: string
  scenarioDescription: string
  topicId: number
  topicTitle: string
  level: number
  userRole: 'buyer' | 'seller'
  counterpartyRole: 'buyer' | 'seller'
  productContext: {
    itemTitle: string
    category: string
    dealMethod: 'delivery' | 'meetup' | 'pickup'
    price?: number
    currency?: 'RUB'
    location?: string
    imageKey?: string
  }
  mode: ResponseMode
  progress: { currentStep: number; answeredSteps: number; totalSteps: number }
  step: {
    id: number
    number: number
    counterpartyMessage: string
    options: TrainingOption[]
  }
  answers: Array<{
    stepId: number
    answerType: 'option' | 'free_text'
    optionId?: number
    optionText?: string
    freeText?: string
    points: number
  }>
  messages: TrainingMessage[]
  canFinishEarly: boolean
}

export type TrainingAnswer =
  | { type: 'option'; stepId: number; optionId: number }
  | { type: 'text'; stepId: number; text: string; finish?: boolean }

export interface DecisionReview {
  stepId: number
  stepNumber: number
  answerType: 'option' | 'free_text'
  optionId?: number
  optionText?: string
  freeText?: string
  points: number
  assessment: 'unsafe' | 'risky' | 'mostly_safe' | 'safe'
  explanation: string
  safeAction: string
  riskSignals: RiskSignal[]
}

export interface RiskSignal {
  code: string
  label: string
}

export interface ResultFeedback {
  reason: string
  riskSignals: RiskSignal[]
  safeAlternative: string
}

export interface MicroQuestion {
  patternCode: string
  question: string
  options: [string, string]
}

export interface MicroQuestionAnswer {
  isCorrect: boolean
  safeAction: string
}

export interface AttemptResult {
  attemptId: number
  score: number
  stars: number
  decisionReview: DecisionReview[]
  riskSignals: RiskSignal[]
  safeActions: string[]
  feedback: ResultFeedback
  microQuestion?: MicroQuestion
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
