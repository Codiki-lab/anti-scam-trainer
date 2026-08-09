import type { AccountDto, RegistrationDto, StreakDto } from '../api/contracts'
import type { Account, Registration, Streak } from '../model/types'

export function mapStreak(dto: StreakDto): Streak {
  return {
    current: dto.current,
    longest: dto.longest,
    isActiveToday: dto.active_today,
    lastActivityDate: dto.last_activity_date,
  }
}

export function mapAccount(dto: AccountDto): Account {
  return {
    id: dto.id,
    username: dto.username,
    accessRole: dto.access_role,
    trainingRole: dto.training_role,
    streak: mapStreak(dto.streak),
  }
}

export function mapRegistration(model: Registration): RegistrationDto {
  return {
    username: model.username,
    password: model.password,
    training_role: model.trainingRole,
  }
}
