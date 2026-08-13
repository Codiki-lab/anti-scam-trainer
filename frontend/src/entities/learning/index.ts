export {
  useGetQuizQuery,
  useGetTheoryQuery,
  useGetTopicsQuery,
  useGetRecommendationQuery,
  useRecommendFromAvitoChatMutation,
  useStartSkillCheckMutation,
  useGetSkillCheckQuery,
  useLazyGetSkillCheckQuery,
  useAnswerSkillCheckMutation,
  useMarkTheoryReadMutation,
  useSubmitQuizMutation,
} from './api/learningApi'
export {
  mapQuiz,
  mapQuizOutcome,
  mapRecommendation,
  mapSkillCheck,
  mapTheory,
  mapTopic,
} from './lib/mappers'
export type { TopicContract, TopicLevelProgress } from './api/contracts'
export { topicContractSchema, topicLevelProgressSchema } from './api/contracts'
export type {
  LevelProgress,
  LearningAction,
  LearningRecommendation,
  Quiz,
  QuizOutcome,
  QuizQuestion,
  QuizSubmission,
  Theory,
  TheorySection,
  Topic,
  SkillCheck,
  SkillCheckPhase,
} from './model/types'
