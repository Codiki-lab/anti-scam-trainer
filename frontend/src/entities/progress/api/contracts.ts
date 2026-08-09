import type { TopicContract } from '@/entities/learning'
import type { StreakDto, UserRole } from '@/entities/user'
import { z } from 'zod'
import { streakDtoSchema, userRoleSchema } from '@/entities/user'
import { topicContractSchema } from '@/entities/learning'

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

export interface DashboardDto {
  profile: { id: number; username: string; training_role: UserRole }
  streak: StreakDto
  topics: TopicContract[]
  achievements: AchievementDto[]
  continue_action: ContinueActionDto | null
  daily_task: null
}

export interface ProgressDto {
  role: UserRole
  summary: {
    completed_topics: number
    total_topics: number
    completed_levels: number
    total_levels: number
    stars: number
    average_score: number
  }
  topics: TopicContract[]
  recent_attempts: Array<{
    attempt_id: number
    topic_id: number
    level: number
    score: number
    stars: number
    finished_at: string
  }>
}

export interface AchievementsDto {
  earned: AchievementDto[]
  available: AchievementDto[]
}

export const achievementDtoSchema = z.object({
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

export const dashboardDtoSchema = z.object({
  profile: z.object({ id: z.number().int(), username: z.string(), training_role: userRoleSchema }),
  streak: streakDtoSchema,
  topics: z.array(topicContractSchema).length(6),
  achievements: z.array(achievementDtoSchema).max(3),
  continue_action: z
    .object({
      type: z.enum([
        'resume_attempt',
        'read_theory',
        'take_quiz',
        'start_level',
        'start_free_play',
      ]),
      topic_id: z.number().int().optional(),
      level: z.number().int().optional(),
      attempt_id: z.number().int().optional(),
    })
    .nullable(),
  daily_task: z.null(),
})

export const progressDtoSchema = z.object({
  role: userRoleSchema,
  summary: z.object({
    completed_topics: z.number().int().nonnegative(),
    total_topics: z.literal(6),
    completed_levels: z.number().int().min(0).max(24),
    total_levels: z.literal(24),
    stars: z.number().int().nonnegative(),
    average_score: z.number().min(0).max(100),
  }),
  topics: z.array(topicContractSchema),
  recent_attempts: z
    .array(
      z.object({
        attempt_id: z.number().int(),
        topic_id: z.number().int(),
        level: z.number().int().min(1).max(4),
        score: z.number().int().min(0).max(100),
        stars: z.number().int().min(0).max(3),
        finished_at: z.string(),
      }),
    )
    .max(10),
})

export const achievementsDtoSchema = z.object({
  earned: z.array(achievementDtoSchema),
  available: z.array(achievementDtoSchema),
})
