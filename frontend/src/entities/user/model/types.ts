export type UserRole = 'buyer' | 'seller'

export interface Streak {
  current: number
  longest: number
  isActiveToday: boolean
  lastActivityDate?: string
}

export interface Account {
  id: number
  username: string
  accessRole: 'user' | 'admin'
  trainingRole: UserRole
  streak: Streak
}

export interface Credentials {
  username: string
  password: string
}

export interface Registration extends Credentials {
  trainingRole: UserRole
}
