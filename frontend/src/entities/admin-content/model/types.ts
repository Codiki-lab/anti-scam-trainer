import type { UserRole } from '@/entities/user'

export type ContentStatus = 'draft' | 'published' | 'archived'
export type TheoryKind = 'intro' | 'risk' | 'example' | 'safe_action' | 'summary'
export type ResponseType = 'multiple_choice' | 'similar_choice' | 'mixed' | 'free_text'
export type OptionPoints = 0 | 25 | 50 | 75 | 100

export interface AdminTheoryBlock {
  id?: number
  sortOrder: number
  kind: TheoryKind
  title: string
  body: string
}

export interface AdminQuizOption {
  id?: number
  sortOrder: number
  text: string
  isCorrect: boolean
}

export interface AdminQuizQuestion {
  id?: number
  sortOrder: number
  text: string
  explanation: string
  options: AdminQuizOption[]
}

export interface AdminTopic {
  id?: number
  slug: string
  role: UserRole
  title: string
  description: string
  sortOrder: number
  status?: ContentStatus
  theory: AdminTheoryBlock[]
  quiz: AdminQuizQuestion[]
}

export interface AdminScenarioOption {
  id?: number
  text: string
  explanation: string
  points: OptionPoints
  sortOrder: number
}

export interface AdminScenarioStep {
  id?: number
  number: number
  responseType: ResponseType
  goal: string
  counterpartyMessage: string
  maxPoints: number
  aiInstruction: string
  fallbackMessage: string
  options: AdminScenarioOption[]
}

export interface AdminScenario {
  id?: number
  title: string
  description: string
  levelId: number
  topicId: number
  role: UserRole
  status?: ContentStatus
  scamScheme: string
  productContext: Record<string, unknown>
  aiSystemPrompt: string
  finalRubric: Record<string, unknown>
  steps: AdminScenarioStep[]
}

export type AdminTopicDraft = Omit<AdminTopic, 'id' | 'status' | 'theory' | 'quiz'>
export type AdminScenarioDraft = Omit<AdminScenario, 'id' | 'status' | 'steps'>
