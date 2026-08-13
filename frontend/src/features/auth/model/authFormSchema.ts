import { z } from 'zod'
import { userRoleSchema } from '@/entities/user'

export const authFormSchema = z.object({
  username: z.string().trim().min(1, 'Введите логин.'),
  password: z.string().min(1, 'Введите пароль.'),
  trainingRole: userRoleSchema,
})

export type AuthFormValues = z.infer<typeof authFormSchema>
