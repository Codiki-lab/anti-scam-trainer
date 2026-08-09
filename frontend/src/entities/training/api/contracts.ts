import type { StreakDto } from '../../user'
import type { TopicLevelProgress } from '@/entities/learning'
import { z } from 'zod'
import { streakDtoSchema } from '@/entities/user'
import { topicLevelProgressSchema } from '@/entities/learning'

export type ResponseModeDto = 'multiple_choice' | 'similar_choice' | 'mixed' | 'free_text'

export interface LevelStateDto {
  number: number
  opened: boolean
  scenario_id: number
}

export const levelStateDtoSchema = z.object({
  number: z.number().int().min(1).max(4),
  opened: z.boolean(),
  scenario_id: z.number().int(),
})

export interface GameStateDto {
  attempt_id: number
  status: 'IN_PROGRESS' | 'COMPLETED' | 'ABANDONED'
  scenario_id: number
  topic_id: number
  product_context: Record<string, unknown>
  mode: ResponseModeDto
  step_progress: { current: number; answered: number }
  step: {
    id: number
    number: number
    counterparty_message: string
    options: Array<{ id: number; text: string }>
  }
  answers: Array<{ step_id: number; option_id: number }>
  messages: Array<{ role: 'user' | 'assistant'; text: string }>
  can_finish_early: boolean
}

export interface AnswerCommandDto {
  step_id: number
  option_id?: number
  free_text?: string
  finish?: boolean
}

export interface AnswerBreakdownDto {
  step_id: number
  option_id?: number
  option_text?: string
  free_text?: string
  points: number
  explanation: string
  risk_signals?: string[]
}

export interface AchievementDto {
  code: string
  title: string
  description: string
  icon: string
  earned: boolean
  earned_at?: string
  progress: { current: number; target: number }
}

export interface ContinueActionDto {
  type: 'resume_attempt' | 'read_theory' | 'take_quiz' | 'start_level' | 'start_free_play'
  topic_id?: number
  level?: number
  attempt_id?: number
}

export interface AttemptResultDto {
  attempt_id: number
  score: number
  stars: number
  decision_review: AnswerBreakdownDto[]
  risk_signals: string[]
  safe_actions: string[]
  level_progress: TopicLevelProgress
  topic_id: number
  topic_completed: boolean
  next_action?: ContinueActionDto | null
  new_achievements: AchievementDto[]
  streak: StreakDto
  is_scam?: boolean
}

const gameOptionSchema = z.object({ id: z.number().int(), text: z.string().max(140) })
const gameStepSchema = z.object({
  id: z.number().int(),
  number: z.number().int(),
  counterparty_message: z.string().max(280),
  options: z.array(gameOptionSchema),
})

export const gameStateDtoSchema = z.object({
  attempt_id: z.number().int(),
  status: z.enum(['IN_PROGRESS', 'COMPLETED', 'ABANDONED']),
  scenario_id: z.number().int().nonnegative(),
  topic_id: z.number().int().nonnegative(),
  product_context: z.record(z.unknown()),
  mode: z.enum(['multiple_choice', 'similar_choice', 'mixed', 'free_text']),
  step_progress: z.object({
    current: z.number().int().nonnegative(),
    answered: z.number().int().nonnegative(),
  }),
  step: gameStepSchema,
  answers: z.array(
    z.object({
      step_id: z.number().int().nonnegative(),
      option_id: z.number().int().nonnegative(),
    }),
  ),
  messages: z.array(z.object({ role: z.enum(['user', 'assistant']), text: z.string().max(400) })),
  can_finish_early: z.boolean(),
})

const achievementDtoSchema = z.object({
  code: z.string(),
  title: z.string(),
  description: z.string(),
  icon: z.string(),
  earned: z.boolean(),
  earned_at: z.string().optional(),
  progress: z.object({
    current: z.number().int().nonnegative(),
    target: z.number().int().positive(),
  }),
})

const continueActionDtoSchema = z.object({
  type: z.enum(['resume_attempt', 'read_theory', 'take_quiz', 'start_level', 'start_free_play']),
  topic_id: z.number().int().optional(),
  level: z.number().int().optional(),
  attempt_id: z.number().int().optional(),
})

export const attemptResultDtoSchema = z.object({
  attempt_id: z.number().int(),
  score: z.number().int().min(0).max(100),
  stars: z.number().int().min(0).max(3),
  decision_review: z.array(
    z.object({
      step_id: z.number().int(),
      option_id: z.number().int().optional(),
      option_text: z.string().optional(),
      free_text: z.string().optional(),
      points: z.union([z.literal(0), z.literal(25), z.literal(50), z.literal(75), z.literal(100)]),
      explanation: z.string(),
      risk_signals: z.array(z.string()).optional(),
    }),
  ),
  risk_signals: z.array(z.string()),
  safe_actions: z.array(z.string()),
  level_progress: topicLevelProgressSchema,
  topic_id: z.number().int(),
  topic_completed: z.boolean(),
  next_action: continueActionDtoSchema.nullable().optional(),
  new_achievements: z.array(achievementDtoSchema),
  streak: streakDtoSchema,
  is_scam: z.boolean().optional(),
})
