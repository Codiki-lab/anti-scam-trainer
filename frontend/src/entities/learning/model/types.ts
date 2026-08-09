import type { Streak, UserRole } from '../../user'

export interface LevelProgress {
  number: number
  isOpened: boolean
  bestScore: number
  stars: number
  attempts: number
  lastAttemptId: number | null
}

export interface Topic {
  id: number
  slug: string
  role: UserRole
  title: string
  description: string
  order: number
  isTheoryRead: boolean
  isQuizPassed: boolean
  bestQuizScore: number
  isCompleted: boolean
  levels: LevelProgress[]
}

export interface TheorySection {
  id: number
  order: number
  kind: string
  title: string
  body: string
}

export interface Theory {
  topic: Topic
  sections: TheorySection[]
}

export interface QuizChoice {
  id: number
  text: string
}

export interface QuizQuestion {
  id: number
  order: number
  text: string
  choices: QuizChoice[]
}

export interface Quiz {
  questions: QuizQuestion[]
  passThreshold: number
}

export interface QuizSubmission {
  answers: Array<{ questionId: number; choiceId: number }>
}

export interface QuizOutcome {
  score: number
  isPassed: boolean
  bestScore: number
  isFirstPass: boolean
  streak: Streak
}
