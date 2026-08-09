import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export const apiTags = {
  account: 'Account',
  topics: 'Topics',
  quiz: 'Quiz',
  levels: 'Levels',
  attempt: 'Attempt',
  dashboard: 'Dashboard',
  progress: 'Progress',
  achievements: 'Achievements',
} as const

export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: `${import.meta.env.VITE_API_BASE_URL ?? '/api'}/v1`,
    credentials: 'include',
  }),
  tagTypes: Object.values(apiTags),
  endpoints: () => ({}),
})
