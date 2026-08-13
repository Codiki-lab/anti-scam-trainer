import { api } from '@/shared/http-client'
import type { UserRole } from '@/entities/user'
import type {
  MarkTheoryReadResponse,
  Quiz as QuizDto,
  QuizAnswer,
  QuizResult as QuizResultDto,
  TheoryResponse as TheoryResponseDto,
  TopicContract as TopicDto,
  RecommendationDto,
  SkillCheckDto,
  AvitoChatRecommendationRequestDto,
} from './contracts'
import {
  quizResultSchema,
  quizSchema,
  markTheoryReadResponseSchema,
  theoryResponseSchema,
  topicContractSchema,
  recommendationDtoSchema,
  skillCheckDtoSchema,
} from './contracts'
import {
  mapQuiz,
  mapQuizOutcome,
  mapRecommendation,
  mapSkillCheck,
  mapTheory,
  mapTopic,
} from '../lib/mappers'
import type {
  LearningRecommendation,
  Quiz,
  QuizOutcome,
  QuizSubmission,
  SkillCheck,
  Theory,
  Topic,
} from '../model/types'

export const learningApi = api.injectEndpoints({
  endpoints: (build) => ({
    getTopics: build.query<Topic[], UserRole>({
      query: (role) => ({ url: '/topics', params: { role } }),
      transformResponse: (response: TopicDto[]) =>
        topicContractSchema.array().parse(response).map(mapTopic),
      providesTags: (result) => [
        { type: 'Topics', id: 'LIST' },
        ...(result ?? []).map((topic) => ({ type: 'Topics' as const, id: topic.id })),
      ],
    }),
    getTheory: build.query<Theory, number>({
      query: (topicId) => `/topics/${topicId}/theory`,
      transformResponse: (response: TheoryResponseDto) =>
        mapTheory(theoryResponseSchema.parse(response)),
      providesTags: (_result, _error, topicId) => [{ type: 'Topics', id: topicId }],
    }),
    markTheoryRead: build.mutation<MarkTheoryReadResponse, number>({
      query: (topicId) => ({ url: `/topics/${topicId}/theory/read`, method: 'POST' }),
      transformResponse: (response: MarkTheoryReadResponse) =>
        markTheoryReadResponseSchema.parse(response),
      invalidatesTags: (_result, _error, topicId) => ['Dashboard', { type: 'Topics', id: topicId }],
    }),
    getQuiz: build.query<Quiz, number>({
      query: (topicId) => `/topics/${topicId}/quiz`,
      transformResponse: (response: QuizDto) => mapQuiz(quizSchema.parse(response)),
      providesTags: (_result, _error, topicId) => [{ type: 'Quiz', id: topicId }],
    }),
    submitQuiz: build.mutation<QuizOutcome, { topicId: number; submission: QuizSubmission }>({
      query: ({ topicId, submission }) => {
        const answers: QuizAnswer[] = submission.answers.map((answer) => ({
          question_id: answer.questionId,
          option_id: answer.choiceId,
        }))

        return { url: `/topics/${topicId}/quiz/attempts`, method: 'POST', body: { answers } }
      },
      transformResponse: (response: QuizResultDto) =>
        mapQuizOutcome(quizResultSchema.parse(response)),
      invalidatesTags: (_result, _error, { topicId }) => [
        'Dashboard',
        'Progress',
        { type: 'Topics', id: topicId },
        { type: 'Quiz', id: topicId },
      ],
    }),
    getRecommendation: build.query<LearningRecommendation, UserRole>({
      query: (role) => ({ url: '/recommendations/next', params: { role } }),
      transformResponse: (response: RecommendationDto) =>
        mapRecommendation(recommendationDtoSchema.parse(response)),
      providesTags: ['Topics'],
    }),
    recommendFromAvitoChat: build.mutation<
      LearningRecommendation,
      AvitoChatRecommendationRequestDto
    >({
      query: (body) => ({
        url: '/integrations/avito-chat/recommendations',
        method: 'POST',
        body,
      }),
      transformResponse: (response: RecommendationDto) =>
        mapRecommendation(recommendationDtoSchema.parse(response)),
    }),
    startSkillCheck: build.mutation<SkillCheck, number>({
      query: (topicId) => ({ url: `/topics/${topicId}/skill-check/start`, method: 'POST' }),
      transformResponse: (response: SkillCheckDto) =>
        mapSkillCheck(skillCheckDtoSchema.parse(response)),
    }),
    getSkillCheck: build.query<SkillCheck, number>({
      query: (checkId) => `/skill-checks/${checkId}`,
      transformResponse: (response: SkillCheckDto) =>
        mapSkillCheck(skillCheckDtoSchema.parse(response)),
      providesTags: (_result, _error, checkId) => [{ type: 'SkillChecks', id: checkId }],
    }),
    answerSkillCheck: build.mutation<SkillCheck, { checkId: number; answer: boolean }>({
      query: ({ checkId, answer }) => ({
        url: `/skill-checks/${checkId}/answers`,
        method: 'POST',
        body: { answer },
      }),
      transformResponse: (response: SkillCheckDto) =>
        mapSkillCheck(skillCheckDtoSchema.parse(response)),
      invalidatesTags: (_result, _error, { checkId }) => [{ type: 'SkillChecks', id: checkId }],
    }),
  }),
})

export const {
  useGetTopicsQuery,
  useGetTheoryQuery,
  useMarkTheoryReadMutation,
  useGetQuizQuery,
  useSubmitQuizMutation,
  useGetRecommendationQuery,
  useRecommendFromAvitoChatMutation,
  useStartSkillCheckMutation,
  useGetSkillCheckQuery,
  useLazyGetSkillCheckQuery,
  useAnswerSkillCheckMutation,
} = learningApi
