import type { StreakDto, UserRole } from '@/entities/user'
import { z } from 'zod'
import { streakDtoSchema, userRoleSchema } from '@/entities/user'

export const topicLevelProgressSchema = z.object({
  number: z.number().int(),
  opened: z.boolean(),
  best_score: z.number().int(),
  stars: z.number().int().min(0).max(3),
  attempts: z.number().int().nonnegative(),
  last_attempt_id: z.number().int(),
})

export const topicContractSchema = z.object({
  id: z.number().int(),
  slug: z.string(),
  role: userRoleSchema,
  title: z.string(),
  description: z.string(),
  sort_order: z.number().int(),
  theory_read: z.boolean(),
  quiz_passed: z.boolean(),
  quiz_best_score: z.number().int(),
  completed: z.boolean(),
  levels: z.array(topicLevelProgressSchema),
})

/** Types mirror backend/openapi/v1/openapi.yaml. No UI-only fields belong here. */
export interface TopicLevelProgress {
  number: number
  opened: boolean
  best_score: number
  stars: number
  attempts: number
  last_attempt_id: number
}

export interface TopicContract {
  id: number
  slug: string
  role: UserRole
  title: string
  description: string
  sort_order: number
  theory_read: boolean
  quiz_passed: boolean
  quiz_best_score: number
  completed: boolean
  levels: TopicLevelProgress[]
}

export interface TheoryBlock {
  id: number
  sort_order: number
  kind: string
  title: string
  body: string
}

export interface TheoryResponse {
  topic: TopicContract
  blocks: TheoryBlock[]
}

export interface QuizOption {
  id: number
  text: string
}

export interface QuizQuestion {
  id: number
  sort_order: number
  text: string
  options: QuizOption[]
}

export interface Quiz {
  questions: QuizQuestion[]
  pass_threshold: 80
}

export interface QuizAnswer {
  question_id: number
  option_id: number
}

export interface QuizResult {
  score: number
  passed: boolean
  best_score: number
  newly_passed: boolean
  streak: StreakDto
}

export interface MarkTheoryReadResponse {
  theory_read: true
  newly_read: boolean
  streak: StreakDto
}

export const theoryResponseSchema = z.object({
  topic: topicContractSchema,
  blocks: z
    .array(
      z.object({
        id: z.number().int(),
        sort_order: z.number().int().min(1).max(5),
        kind: z.string(),
        title: z.string(),
        body: z.string(),
      }),
    )
    .length(5),
})

export const quizSchema = z.object({
  questions: z
    .array(
      z.object({
        id: z.number().int(),
        sort_order: z.number().int(),
        text: z.string(),
        options: z.array(z.object({ id: z.number().int(), text: z.string() })).length(4),
      }),
    )
    .length(5),
  pass_threshold: z.literal(80),
})

export const quizResultSchema = z.object({
  score: z.number().int().min(0).max(100),
  passed: z.boolean(),
  best_score: z.number().int(),
  newly_passed: z.boolean(),
  streak: streakDtoSchema,
})

export const markTheoryReadResponseSchema = z.object({
  theory_read: z.literal(true),
  newly_read: z.boolean(),
  streak: streakDtoSchema,
})
