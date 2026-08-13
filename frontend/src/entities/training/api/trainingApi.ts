import { api } from '@/shared/http-client'
import type { UserRole } from '@/entities/user'
import type {
  AnswerCommandDto,
  AttemptResultDto,
  GameStateDto,
  LevelStateDto,
  MicroQuestionAnswerDto,
} from './contracts'
import {
  attemptResultDtoSchema,
  gameStateDtoSchema,
  levelStateDtoSchema,
  microQuestionAnswerDtoSchema,
} from './contracts'
import { mapAttemptResult, mapLevelState, mapTrainingSession } from '../lib/mappers'
import type {
  AnswerResult,
  AttemptResult,
  LevelState,
  TrainingAnswer,
  TrainingSession,
  MicroQuestionAnswer,
} from '../model/types'

interface TopicLevelRequest {
  role: UserRole
  topicId: number
}

interface StartLevelRequest extends TopicLevelRequest {
  level: number
}

interface SubmitAnswerRequest {
  attemptId: number
  answer: TrainingAnswer
}

interface AnswerMicroQuestionRequest {
  attemptId: number
  answerIndex: 0 | 1
}

export const trainingApi = api.injectEndpoints({
  endpoints: (build) => ({
    getLevels: build.query<LevelState[], TopicLevelRequest>({
      query: ({ role, topicId }) => ({
        url: '/training/levels',
        params: { role, topic_id: topicId },
      }),
      transformResponse: (response: LevelStateDto[]) =>
        levelStateDtoSchema.array().parse(response).map(mapLevelState),
      providesTags: ['Levels'],
    }),
    startLevel: build.mutation<TrainingSession, StartLevelRequest>({
      query: ({ role, topicId, level }) => ({
        url: `/training/levels/${level}/start`,
        method: 'POST',
        params: { role, topic_id: topicId },
      }),
      transformResponse: (response: GameStateDto) =>
        mapTrainingSession(gameStateDtoSchema.parse(response)),
      invalidatesTags: ['Attempt'],
    }),
    startFreePlay: build.mutation<TrainingSession, UserRole>({
      query: (role) => ({
        url: '/training/free-play/start',
        method: 'POST',
        params: { role },
      }),
      transformResponse: (response: GameStateDto) =>
        mapTrainingSession(gameStateDtoSchema.parse(response)),
      invalidatesTags: ['Attempt'],
    }),
    getAttempt: build.query<TrainingSession, number>({
      query: (attemptId) => `/attempts/${attemptId}`,
      transformResponse: (response: GameStateDto) =>
        mapTrainingSession(gameStateDtoSchema.parse(response)),
      providesTags: (_result, _error, attemptId) => [{ type: 'Attempt', id: attemptId }],
    }),
    getAttemptResult: build.query<AttemptResult, number>({
      query: (attemptId) => `/attempts/${attemptId}/result`,
      transformResponse: (response: AttemptResultDto) =>
        mapAttemptResult(attemptResultDtoSchema.parse(response)),
      providesTags: (_result, _error, attemptId) => [{ type: 'Attempt', id: attemptId }],
    }),
    answerMicroQuestion: build.mutation<MicroQuestionAnswer, AnswerMicroQuestionRequest>({
      query: ({ attemptId, answerIndex }) => ({
        url: `/attempts/${attemptId}/micro-question/answer`,
        method: 'POST',
        body: { answer_index: answerIndex },
      }),
      transformResponse: (response: MicroQuestionAnswerDto) => {
        const dto = microQuestionAnswerDtoSchema.parse(response)
        return { isCorrect: dto.correct, safeAction: dto.safe_action }
      },
    }),
    submitAnswer: build.mutation<AnswerResult, SubmitAnswerRequest>({
      query: ({ attemptId, answer }) => {
        const body: AnswerCommandDto =
          answer.type === 'option'
            ? { step_id: answer.stepId, option_id: answer.optionId }
            : { step_id: answer.stepId, free_text: answer.text, finish: answer.finish }

        return { url: `/attempts/${attemptId}/answers`, method: 'POST', body }
      },
      transformResponse: (response: GameStateDto | AttemptResultDto) =>
        'score' in response
          ? mapAttemptResult(attemptResultDtoSchema.parse(response))
          : mapTrainingSession(gameStateDtoSchema.parse(response)),
      invalidatesTags: (_result, _error, { attemptId }) => [
        'Topics',
        'Dashboard',
        'Progress',
        'Achievements',
        'Levels',
        { type: 'Attempt', id: attemptId },
      ],
    }),
    abandonAttempt: build.mutation<void, number>({
      query: (attemptId) => ({ url: `/attempts/${attemptId}/abandon`, method: 'POST' }),
      invalidatesTags: (_result, _error, attemptId) => [{ type: 'Attempt', id: attemptId }],
    }),
  }),
})

export const {
  useGetLevelsQuery,
  useStartLevelMutation,
  useStartFreePlayMutation,
  useGetAttemptQuery,
  useGetAttemptResultQuery,
  useAnswerMicroQuestionMutation,
  useSubmitAnswerMutation,
  useAbandonAttemptMutation,
} = trainingApi
