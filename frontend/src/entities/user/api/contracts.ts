import type { UserRole } from '../model/types'
import { z } from 'zod'

export const userRoleSchema = z.enum(['buyer', 'seller'])

export const streakDtoSchema = z.object({
  current: z.number().int().nonnegative(),
  longest: z.number().int().nonnegative(),
  active_today: z.boolean(),
  last_activity_date: z.string().optional(),
})

export const accountDtoSchema = z.object({
  id: z.number().int(),
  username: z.string(),
  access_role: z.enum(['user', 'admin']),
  training_role: userRoleSchema,
  streak: streakDtoSchema,
})

export interface StreakDto {
  current: number
  longest: number
  active_today: boolean
  last_activity_date?: string
}

export interface AccountDto {
  id: number
  username: string
  access_role: 'user' | 'admin'
  training_role: UserRole
  streak: StreakDto
}

export interface RegistrationDto {
  username: string
  password: string
  training_role: UserRole
}

export interface UpdateTrainingRoleDto {
  training_role: UserRole
}
