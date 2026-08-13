import type { Topic } from '@/entities/learning'
import type { Streak, UserRole } from '@/entities/user'

export type AchievementCode =
  | 'first_training'
  | 'five_trainings'
  | 'perfect_score'
  | 'first_topic_completed'
  | 'all_buyer_topics'
  | 'all_seller_topics'
  | 'streak_3'
  | 'streak_7'

export interface Achievement {
  code: AchievementCode
  title: string
  description: string
  icon: string
  earned: boolean
  earnedAt?: string
  current: number
  target: number
}

export interface ContinueAction {
  type: 'resume_attempt' | 'read_theory' | 'take_quiz' | 'start_level' | 'start_free_play'
  topicId?: number
  level?: number
  attemptId?: number
}

export interface DailyTask {
  date: string
  role: UserRole
  messages: Array<{ role: 'user' | 'assistant'; text: string }>
  isCompleted: boolean
  completedAt?: string
  answer?: boolean
  isCorrect?: boolean
  verdict?: boolean
  signals: string[]
  safeAction?: string
}

export interface DailyTaskAnswer {
  dailyTask: DailyTask
  streak: Streak
}

export interface Dashboard {
  profile: { id: number; username: string; trainingRole: UserRole }
  streak: Streak
  topics: Topic[]
  achievements: Achievement[]
  continueAction: ContinueAction | null
  dailyTask: DailyTask
}

export interface Progress {
  role: UserRole
  summary: {
    completedTopics: number
    totalTopics: number
    completedLevels: number
    totalLevels: number
    stars: number
    averageScore: number
  }
  topics: Topic[]
  recentAttempts: Array<{
    attemptId: number
    topicId: number
    level: number
    score: number
    stars: number
    finishedAt: string
  }>
}

export interface Achievements {
  earned: Achievement[]
  available: Achievement[]
}
