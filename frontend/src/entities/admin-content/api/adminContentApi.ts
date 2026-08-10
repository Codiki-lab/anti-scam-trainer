import { api, apiTags } from '@/shared/http-client'
import {
  adminQuizOptionDtoSchema,
  adminQuizQuestionDtoSchema,
  adminScenarioDtoSchema,
  adminScenarioOptionDtoSchema,
  adminScenarioStepDtoSchema,
  adminTheoryBlockDtoSchema,
  adminTopicDtoSchema,
  type AdminQuizOptionDto,
  type AdminQuizQuestionDto,
  type AdminScenarioDto,
  type AdminScenarioOptionDto,
  type AdminScenarioStepDto,
  type AdminTheoryBlockDto,
  type AdminTopicDto,
} from './contracts'
import {
  mapAdminScenario,
  mapAdminScenarioDto,
  mapAdminTopic,
  mapAdminTopicDto,
  mapQuizOption,
  mapQuizOptionDto,
  mapQuizQuestion,
  mapQuizQuestionDto,
  mapScenarioOption,
  mapScenarioOptionDto,
  mapScenarioStep,
  mapScenarioStepDto,
  mapTheoryBlock,
  mapTheoryBlockDto,
} from '../lib/mappers'
import type {
  AdminQuizOption,
  AdminQuizQuestion,
  AdminScenario,
  AdminScenarioDraft,
  AdminScenarioOption,
  AdminScenarioStep,
  AdminTheoryBlock,
  AdminTopic,
  AdminTopicDraft,
  ContentStatus,
} from '../model/types'

interface TopicItem<T> {
  topicId: number
  value: T
}

interface TopicNestedItem<T> extends TopicItem<T> {
  itemId: number
}

interface QuizOptionItem extends TopicItem<AdminQuizOption> {
  questionId: number
}

interface QuizOptionNestedItem extends QuizOptionItem {
  optionId: number
}

interface ScenarioItem<T> {
  scenarioId: number
  value: T
}

interface ScenarioNestedItem<T> extends ScenarioItem<T> {
  itemId: number
}

interface StepOptionItem {
  stepId: number
  value: AdminScenarioOption
}

interface StepOptionNestedItem extends StepOptionItem {
  optionId: number
}

const emptyResponse = { responseHandler: 'text' as const }

export const adminContentApi = api.injectEndpoints({
  endpoints: (build) => ({
    getAdminTopics: build.query<AdminTopic[], void>({
      query: () => '/admin/topics',
      transformResponse: (response: AdminTopicDto[]) =>
        adminTopicDtoSchema.array().parse(response).map(mapAdminTopic),
      providesTags: [apiTags.adminTopics],
    }),
    getAdminTopic: build.query<AdminTopic, number>({
      query: (id) => `/admin/topics/${id}`,
      transformResponse: (response: AdminTopicDto) =>
        mapAdminTopic(adminTopicDtoSchema.parse(response)),
      providesTags: (_result, _error, id) => [{ type: apiTags.adminTopics, id }],
    }),
    createAdminTopic: build.mutation<AdminTopic, AdminTopicDraft>({
      query: (value) => ({ url: '/admin/topics', method: 'POST', body: mapAdminTopicDto(value) }),
      transformResponse: (response: AdminTopicDto) =>
        mapAdminTopic(adminTopicDtoSchema.parse(response)),
      invalidatesTags: [apiTags.adminTopics],
    }),
    updateAdminTopic: build.mutation<void, TopicItem<AdminTopicDraft>>({
      query: ({ topicId, value }) => ({
        url: `/admin/topics/${topicId}`,
        method: 'PUT',
        body: mapAdminTopicDto(value),
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminTopics],
    }),
    changeAdminTopicStatus: build.mutation<
      void,
      { topicId: number; action: 'publish' | 'deactivate' | 'restore' | 'archive' }
    >({
      query: ({ topicId, action }) => ({
        url:
          action === 'archive' ? `/admin/topics/${topicId}` : `/admin/topics/${topicId}/${action}`,
        method: action === 'archive' ? 'DELETE' : 'POST',
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminTopics],
    }),
    createAdminTheoryBlock: build.mutation<AdminTheoryBlock, TopicItem<AdminTheoryBlock>>({
      query: ({ topicId, value }) => ({
        url: `/admin/topics/${topicId}/theory-blocks`,
        method: 'POST',
        body: mapTheoryBlockDto(value),
      }),
      transformResponse: (response: AdminTheoryBlockDto) =>
        mapTheoryBlock(adminTheoryBlockDtoSchema.parse(response)),
      invalidatesTags: [apiTags.adminTopics],
    }),
    updateAdminTheoryBlock: build.mutation<void, TopicNestedItem<AdminTheoryBlock>>({
      query: ({ topicId, itemId, value }) => ({
        url: `/admin/topics/${topicId}/theory-blocks/${itemId}`,
        method: 'PUT',
        body: mapTheoryBlockDto(value),
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminTopics],
    }),
    deleteAdminTheoryBlock: build.mutation<void, { topicId: number; itemId: number }>({
      query: ({ topicId, itemId }) => ({
        url: `/admin/topics/${topicId}/theory-blocks/${itemId}`,
        method: 'DELETE',
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminTopics],
    }),
    createAdminQuizQuestion: build.mutation<AdminQuizQuestion, TopicItem<AdminQuizQuestion>>({
      query: ({ topicId, value }) => ({
        url: `/admin/topics/${topicId}/quiz-questions`,
        method: 'POST',
        body: mapQuizQuestionDto(value),
      }),
      transformResponse: (response: AdminQuizQuestionDto) =>
        mapQuizQuestion(adminQuizQuestionDtoSchema.parse(response)),
      invalidatesTags: [apiTags.adminTopics],
    }),
    updateAdminQuizQuestion: build.mutation<void, TopicNestedItem<AdminQuizQuestion>>({
      query: ({ topicId, itemId, value }) => ({
        url: `/admin/topics/${topicId}/quiz-questions/${itemId}`,
        method: 'PUT',
        body: mapQuizQuestionDto(value),
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminTopics],
    }),
    deleteAdminQuizQuestion: build.mutation<void, { topicId: number; itemId: number }>({
      query: ({ topicId, itemId }) => ({
        url: `/admin/topics/${topicId}/quiz-questions/${itemId}`,
        method: 'DELETE',
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminTopics],
    }),
    createAdminQuizOption: build.mutation<AdminQuizOption, QuizOptionItem>({
      query: ({ topicId, questionId, value }) => ({
        url: `/admin/topics/${topicId}/quiz-questions/${questionId}/options`,
        method: 'POST',
        body: mapQuizOptionDto(value),
      }),
      transformResponse: (response: AdminQuizOptionDto) =>
        mapQuizOption(adminQuizOptionDtoSchema.parse(response)),
      invalidatesTags: [apiTags.adminTopics],
    }),
    updateAdminQuizOption: build.mutation<void, QuizOptionNestedItem>({
      query: ({ topicId, questionId, optionId, value }) => ({
        url: `/admin/topics/${topicId}/quiz-questions/${questionId}/options/${optionId}`,
        method: 'PUT',
        body: mapQuizOptionDto(value),
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminTopics],
    }),
    deleteAdminQuizOption: build.mutation<
      void,
      { topicId: number; questionId: number; optionId: number }
    >({
      query: ({ topicId, questionId, optionId }) => ({
        url: `/admin/topics/${topicId}/quiz-questions/${questionId}/options/${optionId}`,
        method: 'DELETE',
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminTopics],
    }),
    getAdminScenarios: build.query<AdminScenario[], void>({
      query: () => '/admin/scenarios',
      transformResponse: (response: AdminScenarioDto[]) =>
        adminScenarioDtoSchema.array().parse(response).map(mapAdminScenario),
      providesTags: [apiTags.adminScenarios],
    }),
    createAdminScenario: build.mutation<AdminScenario, AdminScenarioDraft>({
      query: (value) => ({
        url: '/admin/scenarios',
        method: 'POST',
        body: mapAdminScenarioDto(value),
      }),
      transformResponse: (response: AdminScenarioDto) =>
        mapAdminScenario(adminScenarioDtoSchema.parse(response)),
      invalidatesTags: [apiTags.adminScenarios],
    }),
    updateAdminScenario: build.mutation<void, ScenarioItem<AdminScenarioDraft>>({
      query: ({ scenarioId, value }) => ({
        url: `/admin/scenarios/${scenarioId}`,
        method: 'PUT',
        body: mapAdminScenarioDto(value),
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminScenarios],
    }),
    changeAdminScenarioStatus: build.mutation<
      void,
      { scenarioId: number; action: 'publish' | 'deactivate' | 'restore' | 'archive' }
    >({
      query: ({ scenarioId, action }) => ({
        url:
          action === 'archive'
            ? `/admin/scenarios/${scenarioId}`
            : `/admin/scenarios/${scenarioId}/${action}`,
        method: action === 'archive' ? 'DELETE' : 'POST',
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminScenarios],
    }),
    createAdminScenarioStep: build.mutation<AdminScenarioStep, ScenarioItem<AdminScenarioStep>>({
      query: ({ scenarioId, value }) => ({
        url: `/admin/scenarios/${scenarioId}/steps`,
        method: 'POST',
        body: mapScenarioStepDto(value),
      }),
      transformResponse: (response: AdminScenarioStepDto) =>
        mapScenarioStep(adminScenarioStepDtoSchema.parse(response)),
      invalidatesTags: [apiTags.adminScenarios],
    }),
    updateAdminScenarioStep: build.mutation<void, ScenarioNestedItem<AdminScenarioStep>>({
      query: ({ scenarioId, itemId, value }) => ({
        url: `/admin/scenarios/${scenarioId}/steps/${itemId}`,
        method: 'PUT',
        body: mapScenarioStepDto(value),
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminScenarios],
    }),
    deleteAdminScenarioStep: build.mutation<void, { scenarioId: number; itemId: number }>({
      query: ({ scenarioId, itemId }) => ({
        url: `/admin/scenarios/${scenarioId}/steps/${itemId}`,
        method: 'DELETE',
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminScenarios],
    }),
    createAdminScenarioOption: build.mutation<AdminScenarioOption, StepOptionItem>({
      query: ({ stepId, value }) => ({
        url: `/admin/steps/${stepId}/options`,
        method: 'POST',
        body: mapScenarioOptionDto(value),
      }),
      transformResponse: (response: AdminScenarioOptionDto) =>
        mapScenarioOption(adminScenarioOptionDtoSchema.parse(response)),
      invalidatesTags: [apiTags.adminScenarios],
    }),
    updateAdminScenarioOption: build.mutation<void, StepOptionNestedItem>({
      query: ({ stepId, optionId, value }) => ({
        url: `/admin/steps/${stepId}/options/${optionId}`,
        method: 'PUT',
        body: mapScenarioOptionDto(value),
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminScenarios],
    }),
    deleteAdminScenarioOption: build.mutation<void, { stepId: number; optionId: number }>({
      query: ({ stepId, optionId }) => ({
        url: `/admin/steps/${stepId}/options/${optionId}`,
        method: 'DELETE',
        ...emptyResponse,
      }),
      invalidatesTags: [apiTags.adminScenarios],
    }),
  }),
})

export const {
  useGetAdminTopicsQuery,
  useGetAdminTopicQuery,
  useCreateAdminTopicMutation,
  useUpdateAdminTopicMutation,
  useChangeAdminTopicStatusMutation,
  useCreateAdminTheoryBlockMutation,
  useUpdateAdminTheoryBlockMutation,
  useDeleteAdminTheoryBlockMutation,
  useCreateAdminQuizQuestionMutation,
  useUpdateAdminQuizQuestionMutation,
  useDeleteAdminQuizQuestionMutation,
  useCreateAdminQuizOptionMutation,
  useUpdateAdminQuizOptionMutation,
  useDeleteAdminQuizOptionMutation,
  useGetAdminScenariosQuery,
  useCreateAdminScenarioMutation,
  useUpdateAdminScenarioMutation,
  useChangeAdminScenarioStatusMutation,
  useCreateAdminScenarioStepMutation,
  useUpdateAdminScenarioStepMutation,
  useDeleteAdminScenarioStepMutation,
  useCreateAdminScenarioOptionMutation,
  useUpdateAdminScenarioOptionMutation,
  useDeleteAdminScenarioOptionMutation,
} = adminContentApi

export const statusLabels: Record<ContentStatus, string> = {
  draft: 'Черновик',
  published: 'Опубликовано',
  archived: 'В архиве',
}
