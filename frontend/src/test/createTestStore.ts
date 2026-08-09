import { configureStore } from '@reduxjs/toolkit'
import { api } from '@/shared/http-client'

export function createTestStore() {
  return configureStore({
    reducer: { [api.reducerPath]: api.reducer },
    middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(api.middleware),
  })
}
