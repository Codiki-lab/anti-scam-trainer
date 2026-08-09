import { api } from '@/shared/http-client'
import type { UserRole } from '@/entities/user'
import type {
  MarkTheoryReadResponse,
  Quiz as QuizDto,
  QuizAnswer,
  QuizResult as QuizResultDto,
  TheoryResponse as TheoryResponseDto,
  TopicContract as TopicDto,
} from './contracts'
import {
  quizResultSchema,
  quizSchema,
  markTheoryReadResponseSchema,
  theoryResponseSchema,
  topicContractSchema,
} from './contracts'
import { mapQuiz, mapQuizOutcome, mapTheory, mapTopic } from '../lib/mappers'
import type { Quiz, QuizOutcome, QuizSubmission, Theory, Topic } from '../model/types'

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
  }),
})

export const {
  useGetTopicsQuery,
  useGetTheoryQuery,
  useMarkTheoryReadMutation,
  useGetQuizQuery,
  useSubmitQuizMutation,
} = learningApi
