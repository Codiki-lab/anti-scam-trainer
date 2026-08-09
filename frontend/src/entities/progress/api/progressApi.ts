import { api } from '@/shared/http-client'
import type { UserRole } from '@/entities/user'
import type { AchievementsDto, DailyTaskAnswerDto, DashboardDto, ProgressDto } from './contracts'
import {
  achievementsDtoSchema,
  dailyTaskAnswerDtoSchema,
  dashboardDtoSchema,
  progressDtoSchema,
} from './contracts'
import { mapAchievements, mapDailyTaskAnswer, mapDashboard, mapProgress } from '../lib/mappers'
import type { Achievements, DailyTaskAnswer, Dashboard, Progress } from '../model/types'

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
    answerDailyTask: build.mutation<DailyTaskAnswer, boolean>({
      query: (answer) => ({ url: '/daily-tasks/answer', method: 'POST', body: { answer } }),
      transformResponse: (response: DailyTaskAnswerDto) =>
        mapDailyTaskAnswer(dailyTaskAnswerDtoSchema.parse(response)),
      invalidatesTags: ['Account', 'Dashboard'],
    }),
  }),
})

export const {
  useGetDashboardQuery,
  useGetProgressQuery,
  useGetAchievementsQuery,
  useAnswerDailyTaskMutation,
} = progressApi
