import { api } from '@/shared/http-client'
import { accountDtoSchema, type AccountDto, type UpdateTrainingRoleDto } from './contracts'
import { mapAccount, mapRegistration } from '../lib/mappers'
import type { Account, Credentials, Registration, UserRole } from '../model/types'

export const userApi = api.injectEndpoints({
  endpoints: (build) => ({
    getMe: build.query<Account, void>({
      query: () => '/auth/me',
      transformResponse: (response: AccountDto) => mapAccount(accountDtoSchema.parse(response)),
      providesTags: ['Account'],
    }),
    register: build.mutation<Account, Registration>({
      query: (body) => ({ url: '/auth/register', method: 'POST', body: mapRegistration(body) }),
      transformResponse: (response: AccountDto) => mapAccount(accountDtoSchema.parse(response)),
    }),
    login: build.mutation<void, Credentials>({
      query: (body) => ({ url: '/auth/login', method: 'POST', body, responseHandler: 'text' }),
      invalidatesTags: ['Account'],
    }),
    logout: build.mutation<void, void>({
      query: () => ({ url: '/auth/logout', method: 'POST', responseHandler: 'text' }),
      invalidatesTags: [
        'Account',
        'Topics',
        'Levels',
        'Attempt',
        'Dashboard',
        'Progress',
        'Achievements',
      ],
    }),
    updateTrainingRole: build.mutation<Account, UserRole>({
      query: (role) => {
        const body: UpdateTrainingRoleDto = { training_role: role }
        return { url: '/profile/preferences', method: 'PATCH', body }
      },
      transformResponse: (response: AccountDto) => mapAccount(accountDtoSchema.parse(response)),
      async onQueryStarted(_role, { dispatch, queryFulfilled }) {
        try {
          const { data: account } = await queryFulfilled
          dispatch(userApi.util.updateQueryData('getMe', undefined, () => account))
        } catch {
          // Mutation state exposes the transport error to the feature that initiated the change.
        }
      },
      invalidatesTags: ['Topics', 'Levels', 'Dashboard', 'Progress'],
    }),
  }),
})

export const {
  useGetMeQuery,
  useRegisterMutation,
  useLoginMutation,
  useLogoutMutation,
  useUpdateTrainingRoleMutation,
} = userApi
