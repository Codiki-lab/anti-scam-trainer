import { createContext, useContext } from 'react'

export type RuntimeMode = 'production' | 'preview'

export const RuntimeModeContext = createContext<RuntimeMode>('production')

export function useIsPreview() {
  return useContext(RuntimeModeContext) === 'preview'
}
