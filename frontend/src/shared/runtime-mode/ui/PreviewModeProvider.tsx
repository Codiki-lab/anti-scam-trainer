import type { PropsWithChildren } from 'react'
import { RuntimeModeContext } from '../model/runtimeMode'

export function PreviewModeProvider({ children }: PropsWithChildren) {
  return <RuntimeModeContext.Provider value="preview">{children}</RuntimeModeContext.Provider>
}
