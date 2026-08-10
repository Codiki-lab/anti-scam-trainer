import { z } from 'zod'
import { userRoleSchema } from '@/entities/user'

export const contentStatusSchema = z.enum(['draft', 'published', 'archived'])

export const adminTheoryBlockDtoSchema = z.object({
  id: z.number().int().optional(),
  sort_order: z.number().int().min(1).max(5),
  kind: z.enum(['intro', 'risk', 'example', 'safe_action', 'summary']),
  title: z.string().min(1),
  body: z.string().min(1),
})

export const adminQuizOptionDtoSchema = z.object({
  id: z.number().int().optional(),
  sort_order: z.number().int().min(1).max(4),
  text: z.string().min(1),
  is_correct: z.boolean(),
})

export const adminQuizQuestionDtoSchema = z.object({
  id: z.number().int().optional(),
  sort_order: z.number().int().min(1).max(5),
  text: z.string().min(1),
  explanation: z.string().min(1),
  options: z.array(adminQuizOptionDtoSchema).default([]),
})

export const adminTopicDtoSchema = z.object({
  id: z.number().int().optional(),
  slug: z.string().min(1),
  role: userRoleSchema,
  title: z.string().min(1),
  description: z.string().min(1),
  sort_order: z.number().int().min(1).max(6),
  status: contentStatusSchema.optional(),
  theory: z.array(adminTheoryBlockDtoSchema).default([]),
  quiz: z.array(adminQuizQuestionDtoSchema).default([]),
})

export const adminScenarioOptionDtoSchema = z.object({
  id: z.number().int().optional(),
  text: z.string().min(1).max(140),
  explanation: z.string(),
  points: z.union([z.literal(0), z.literal(25), z.literal(50), z.literal(75), z.literal(100)]),
  sort_order: z.number().int().min(1),
})

export const adminScenarioStepDtoSchema = z.object({
  id: z.number().int().optional(),
  number: z.number().int().min(1),
  response_type: z.enum(['multiple_choice', 'similar_choice', 'mixed', 'free_text']),
  goal: z.string(),
  counterparty_message: z.string().min(1).max(280),
  max_points: z.number().int().nonnegative(),
  ai_instruction: z.string(),
  fallback_message: z.string(),
  options: z.array(adminScenarioOptionDtoSchema).default([]),
})

export const adminScenarioDtoSchema = z.object({
  id: z.number().int().optional(),
  title: z.string().min(1),
  description: z.string(),
  level_id: z.number().int().positive(),
  topic_id: z.number().int().positive(),
  role: userRoleSchema,
  status: contentStatusSchema.optional(),
  scam_scheme: z.string(),
  product_context: z.record(z.unknown()),
  ai_system_prompt: z.string(),
  final_rubric: z.record(z.unknown()),
  steps: z.array(adminScenarioStepDtoSchema).default([]),
})

export type AdminTheoryBlockDto = z.infer<typeof adminTheoryBlockDtoSchema>
export type AdminQuizOptionDto = z.infer<typeof adminQuizOptionDtoSchema>
export type AdminQuizQuestionDto = z.infer<typeof adminQuizQuestionDtoSchema>
export type AdminTopicDto = z.infer<typeof adminTopicDtoSchema>
export type AdminScenarioOptionDto = z.infer<typeof adminScenarioOptionDtoSchema>
export type AdminScenarioStepDto = z.infer<typeof adminScenarioStepDtoSchema>
export type AdminScenarioDto = z.infer<typeof adminScenarioDtoSchema>

export type AdminTopicWriteDto = Omit<AdminTopicDto, 'id' | 'status' | 'theory' | 'quiz'>
export type AdminQuizQuestionWriteDto = Omit<AdminQuizQuestionDto, 'id' | 'options'>
export type AdminScenarioStepWriteDto = Omit<AdminScenarioStepDto, 'id' | 'options'>
export type AdminScenarioWriteDto = Omit<AdminScenarioDto, 'id' | 'status' | 'steps'>
