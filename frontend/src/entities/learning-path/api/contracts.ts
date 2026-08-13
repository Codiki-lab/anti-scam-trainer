import { z } from 'zod'

export const continueActionDtoSchema = z.object({
  type: z.enum(['resume_attempt', 'read_theory', 'take_quiz', 'start_level', 'start_free_play']),
  topic_id: z.number().int().positive().optional(),
  level: z.number().int().min(1).max(4).optional(),
  attempt_id: z.number().int().positive().optional(),
})

export type ContinueActionDto = z.infer<typeof continueActionDtoSchema>
