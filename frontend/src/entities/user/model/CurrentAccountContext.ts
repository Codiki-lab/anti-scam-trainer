import { createContext, useContext } from 'react'
import type { Account, UserRole } from './types'

export interface CurrentAccountValue {
  account: Account
  changeTrainingRole: (role: UserRole) => Promise<void>
}

export const CurrentAccountContext = createContext<CurrentAccountValue | null>(null)

export function useCurrentAccount(): CurrentAccountValue {
  const value = useContext(CurrentAccountContext)
  if (!value) throw new Error('CurrentAccountProvider is missing')
  return value
}
