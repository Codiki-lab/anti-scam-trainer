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
  scenario_title: string
  scenario_description: string
  response_type: ResponseModeDto
  in_progress_attempt_id?: number
}

export const levelStateDtoSchema = z.object({
  number: z.number().int().min(1).max(4),
  opened: z.boolean(),
  scenario_id: z.number().int(),
  scenario_title: z.string(),
  scenario_description: z.string(),
  response_type: z.enum(['multiple_choice', 'similar_choice', 'mixed', 'free_text']),
  in_progress_attempt_id: z.number().int().positive().optional(),
})

export interface GameStateDto {
  attempt_id: number
  status: 'IN_PROGRESS' | 'COMPLETED' | 'ABANDONED'
  scenario_id: number
  scenario_title: string
  scenario_description: string
  topic_id: number
  topic_title: string
  level: number
  user_role: 'buyer' | 'seller'
  counterparty_role: 'buyer' | 'seller'
  product_context: {
    item_title: string
    category: string
    deal_method: 'delivery' | 'meetup' | 'pickup'
    price?: number
    currency?: 'RUB'
    location?: string
    image_key?:
      | 'smartphone'
      | 'electronics'
      | 'appliance'
      | 'camera'
      | 'bicycle'
      | 'laptop'
      | 'headphones'
      | 'console'
  }
  mode: ResponseModeDto
  step_progress: { current: number; answered: number; total: number }
  step: {
    id: number
    number: number
    counterparty_message: string
    options: Array<{ id: number; text: string }>
  }
  answers: Array<{
    step_id: number
    answer_type: 'option' | 'free_text'
    option_id?: number
    option_text?: string
    free_text?: string
    points: 0 | 25 | 50 | 75 | 100
  }>
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
  step_number: number
  answer_type: 'option' | 'free_text'
  option_id?: number
  option_text?: string
  free_text?: string
  points: number
  assessment: 'unsafe' | 'risky' | 'mostly_safe' | 'safe'
  explanation: string
  safe_action: string
  risk_signals: RiskSignalDto[]
}

export interface RiskSignalDto {
  code: string
  label: string
}

export interface ResultFeedbackDto {
  reason: string
  risk_signals: RiskSignalDto[]
  safe_alternative: string
}

export interface MicroQuestionDto {
  pattern_code: string
  question: string
  options: [string, string]
}

export interface MicroQuestionAnswerDto {
  correct: boolean
  safe_action: string
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
  risk_signals: RiskSignalDto[]
  safe_actions: string[]
  feedback: ResultFeedbackDto
  micro_question?: MicroQuestionDto
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

const productContextSchema = z.object({
  item_title: z.string().min(1),
  category: z.string().min(1),
  deal_method: z.enum(['delivery', 'meetup', 'pickup']),
  price: z.number().int().nonnegative().optional(),
  currency: z.literal('RUB').optional(),
  location: z.string().optional(),
  image_key: z
    .enum([
      'smartphone',
      'electronics',
      'appliance',
      'camera',
      'bicycle',
      'laptop',
      'headphones',
      'console',
    ])
    .optional(),
})

export const gameStateDtoSchema = z.object({
  attempt_id: z.number().int(),
  status: z.enum(['IN_PROGRESS', 'COMPLETED', 'ABANDONED']),
  scenario_id: z.number().int().nonnegative(),
  scenario_title: z.string(),
  scenario_description: z.string(),
  topic_id: z.number().int().nonnegative(),
  topic_title: z.string(),
  level: z.number().int().min(0).max(4),
  user_role: z.enum(['buyer', 'seller']),
  counterparty_role: z.enum(['buyer', 'seller']),
  product_context: productContextSchema,
  mode: z.enum(['multiple_choice', 'similar_choice', 'mixed', 'free_text']),
  step_progress: z.object({
    current: z.number().int().nonnegative(),
    answered: z.number().int().nonnegative(),
    total: z.number().int().positive(),
  }),
  step: gameStepSchema,
  answers: z.array(
    z.object({
      step_id: z.number().int().nonnegative(),
      answer_type: z.enum(['option', 'free_text']),
      option_id: z.number().int().positive().optional(),
      option_text: z.string().optional(),
      free_text: z.string().optional(),
      points: z.union([z.literal(0), z.literal(25), z.literal(50), z.literal(75), z.literal(100)]),
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

const riskSignalDtoSchema = z.object({
  code: z.string(),
  label: z.string(),
})

const resultFeedbackDtoSchema = z.object({
  reason: z.string(),
  risk_signals: z.array(riskSignalDtoSchema).max(2),
  safe_alternative: z.string(),
})

const microQuestionDtoSchema = z.object({
  pattern_code: z.string(),
  question: z.string(),
  options: z.tuple([z.string(), z.string()]),
})

export const microQuestionAnswerDtoSchema = z.object({
  correct: z.boolean(),
  safe_action: z.string(),
})

export const attemptResultDtoSchema = z.object({
  attempt_id: z.number().int(),
  score: z.number().int().min(0).max(100),
  stars: z.number().int().min(0).max(3),
  decision_review: z.array(
    z.object({
      step_id: z.number().int(),
      step_number: z.number().int().positive(),
      answer_type: z.enum(['option', 'free_text']),
      option_id: z.number().int().optional(),
      option_text: z.string().optional(),
      free_text: z.string().optional(),
      points: z.union([z.literal(0), z.literal(25), z.literal(50), z.literal(75), z.literal(100)]),
      assessment: z.enum(['unsafe', 'risky', 'mostly_safe', 'safe']),
      explanation: z.string(),
      safe_action: z.string(),
      risk_signals: z.array(riskSignalDtoSchema),
    }),
  ),
  risk_signals: z.array(riskSignalDtoSchema),
  safe_actions: z.array(z.string()),
  feedback: resultFeedbackDtoSchema,
  micro_question: microQuestionDtoSchema.optional(),
  level_progress: topicLevelProgressSchema,
  topic_id: z.number().int(),
  topic_completed: z.boolean(),
  next_action: continueActionDtoSchema.nullable().optional(),
  new_achievements: z.array(achievementDtoSchema),
  streak: streakDtoSchema,
  is_scam: z.boolean().optional(),
})
