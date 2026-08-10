import type {
  AdminQuizOptionDto,
  AdminQuizQuestionDto,
  AdminQuizQuestionWriteDto,
  AdminScenarioDto,
  AdminScenarioOptionDto,
  AdminScenarioStepDto,
  AdminScenarioStepWriteDto,
  AdminScenarioWriteDto,
  AdminTheoryBlockDto,
  AdminTopicDto,
  AdminTopicWriteDto,
} from '../api/contracts'
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
} from '../model/types'

export const mapTheoryBlock = (value: AdminTheoryBlockDto): AdminTheoryBlock => ({
  id: value.id,
  sortOrder: value.sort_order,
  kind: value.kind,
  title: value.title,
  body: value.body,
})

export const mapTheoryBlockDto = (value: AdminTheoryBlock): AdminTheoryBlockDto => ({
  sort_order: value.sortOrder,
  kind: value.kind,
  title: value.title,
  body: value.body,
})

export const mapQuizOption = (value: AdminQuizOptionDto): AdminQuizOption => ({
  id: value.id,
  sortOrder: value.sort_order,
  text: value.text,
  isCorrect: value.is_correct,
})

export const mapQuizOptionDto = (value: AdminQuizOption): AdminQuizOptionDto => ({
  sort_order: value.sortOrder,
  text: value.text,
  is_correct: value.isCorrect,
})

export const mapQuizQuestion = (value: AdminQuizQuestionDto): AdminQuizQuestion => ({
  id: value.id,
  sortOrder: value.sort_order,
  text: value.text,
  explanation: value.explanation,
  options: value.options.map(mapQuizOption),
})

export const mapQuizQuestionDto = (value: AdminQuizQuestion): AdminQuizQuestionWriteDto => ({
  sort_order: value.sortOrder,
  text: value.text,
  explanation: value.explanation,
})

export const mapAdminTopic = (value: AdminTopicDto): AdminTopic => ({
  id: value.id,
  slug: value.slug,
  role: value.role,
  title: value.title,
  description: value.description,
  sortOrder: value.sort_order,
  status: value.status,
  theory: value.theory.map(mapTheoryBlock),
  quiz: value.quiz.map(mapQuizQuestion),
})

export const mapAdminTopicDto = (value: AdminTopicDraft): AdminTopicWriteDto => ({
  slug: value.slug,
  role: value.role,
  title: value.title,
  description: value.description,
  sort_order: value.sortOrder,
})

export const mapScenarioOption = (value: AdminScenarioOptionDto): AdminScenarioOption => ({
  id: value.id,
  text: value.text,
  explanation: value.explanation,
  points: value.points,
  sortOrder: value.sort_order,
})

export const mapScenarioOptionDto = (value: AdminScenarioOption): AdminScenarioOptionDto => ({
  text: value.text,
  explanation: value.explanation,
  points: value.points,
  sort_order: value.sortOrder,
})

export const mapScenarioStep = (value: AdminScenarioStepDto): AdminScenarioStep => ({
  id: value.id,
  number: value.number,
  responseType: value.response_type,
  goal: value.goal,
  counterpartyMessage: value.counterparty_message,
  maxPoints: value.max_points,
  aiInstruction: value.ai_instruction,
  fallbackMessage: value.fallback_message,
  options: value.options.map(mapScenarioOption),
})

export const mapScenarioStepDto = (value: AdminScenarioStep): AdminScenarioStepWriteDto => ({
  number: value.number,
  response_type: value.responseType,
  goal: value.goal,
  counterparty_message: value.counterpartyMessage,
  max_points: value.maxPoints,
  ai_instruction: value.aiInstruction,
  fallback_message: value.fallbackMessage,
})

export const mapAdminScenario = (value: AdminScenarioDto): AdminScenario => ({
  id: value.id,
  title: value.title,
  description: value.description,
  levelId: value.level_id,
  topicId: value.topic_id,
  role: value.role,
  status: value.status,
  scamScheme: value.scam_scheme,
  productContext: value.product_context,
  aiSystemPrompt: value.ai_system_prompt,
  finalRubric: value.final_rubric,
  steps: value.steps.map(mapScenarioStep),
})

export const mapAdminScenarioDto = (value: AdminScenarioDraft): AdminScenarioWriteDto => ({
  title: value.title,
  description: value.description,
  level_id: value.levelId,
  topic_id: value.topicId,
  role: value.role,
  scam_scheme: value.scamScheme,
  product_context: value.productContext,
  ai_system_prompt: value.aiSystemPrompt,
  final_rubric: value.finalRubric,
})
