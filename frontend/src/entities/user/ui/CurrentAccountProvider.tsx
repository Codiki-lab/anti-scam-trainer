import type { PropsWithChildren } from 'react'
import { CurrentAccountContext, type CurrentAccountValue } from '../model/CurrentAccountContext'

export function CurrentAccountProvider({
  children,
  value,
}: PropsWithChildren<{ value: CurrentAccountValue }>) {
  return <CurrentAccountContext.Provider value={value}>{children}</CurrentAccountContext.Provider>
}
