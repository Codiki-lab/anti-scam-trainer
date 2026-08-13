import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query'
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
  adminTopics: 'AdminTopics',
  adminScenarios: 'AdminScenarios',
} as const

const configuredApiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? '/api'
const apiBaseUrl =
  import.meta.env.MODE === 'test'
    ? new URL(configuredApiBaseUrl, 'http://localhost').toString().replace(/\/$/, '')
    : configuredApiBaseUrl

const baseQuery = fetchBaseQuery({
  baseUrl: `${apiBaseUrl}/v1`,
  credentials: 'include',
})

const baseQueryWithAuthenticationRedirect: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  const result = await baseQuery(args, api, extraOptions)
  const pathname = typeof window === 'undefined' ? '' : window.location.pathname

  if (
    result.error?.status === 401 &&
    !pathname.startsWith('/login') &&
    !pathname.startsWith('/register') &&
    !pathname.startsWith('/preview')
  ) {
    window.location.assign('/login')
  }

  return result
}

export const api = createApi({
  reducerPath: 'api',
  baseQuery: baseQueryWithAuthenticationRedirect,
  tagTypes: Object.values(apiTags),
  endpoints: () => ({}),
})
