import { api } from '@/shared/http-client'
import type { UserRole } from '@/entities/user'
import type { AchievementsDto, DashboardDto, ProgressDto } from './contracts'
import { achievementsDtoSchema, dashboardDtoSchema, progressDtoSchema } from './contracts'
import { mapAchievements, mapDashboard, mapProgress } from '../lib/mappers'
import type { Achievements, Dashboard, Progress } from '../model/types'

export const progressApi = api.injectEndpoints({
  endpoints: (build) => ({
    getDashboard: build.query<Dashboard, UserRole>({
      query: (role: UserRole) => ({ url: '/dashboard', params: { role } }),
      transformResponse: (response: DashboardDto) =>
        mapDashboard(dashboardDtoSchema.parse(response)),
      providesTags: ['Dashboard'],
    }),
    getProgress: build.query<Progress, UserRole>({
      query: (role: UserRole) => ({ url: '/progress', params: { role } }),
      transformResponse: (response: ProgressDto) => mapProgress(progressDtoSchema.parse(response)),
      providesTags: ['Progress'],
    }),
    getAchievements: build.query<Achievements, void>({
      query: () => '/achievements',
      transformResponse: (response: AchievementsDto) =>
        mapAchievements(achievementsDtoSchema.parse(response)),
      providesTags: ['Achievements'],
    }),
  }),
})

export const { useGetDashboardQuery, useGetProgressQuery, useGetAchievementsQuery } = progressApi
