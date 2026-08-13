import type { ContinueAction } from '@/entities/learning-path'

export function getContinueLearningPath(action: ContinueAction | null, basePath: string): string {
  if (!action) return `${basePath}/lessons`

  switch (action.type) {
    case 'resume_attempt':
      return action.attemptId ? `${basePath}/sessions/${action.attemptId}` : `${basePath}/chats`
    case 'read_theory':
      return action.topicId ? `${basePath}/lessons/${action.topicId}` : `${basePath}/lessons`
    case 'take_quiz':
      return action.topicId ? `${basePath}/lessons/${action.topicId}/quiz` : `${basePath}/lessons`
    case 'start_level':
      return action.topicId ? `${basePath}/chats?topic=${action.topicId}` : `${basePath}/chats`
    case 'start_free_play':
      return `${basePath}/dashboard#free-play`
  }
}

export function getContinueLearningLabel(action: ContinueAction | null): string {
  if (!action) return 'Выбрать тему'

  switch (action.type) {
    case 'resume_attempt':
      return 'Продолжить тренировку'
    case 'read_theory':
      return 'Продолжить теорию'
    case 'take_quiz':
      return 'Пройти Quiz'
    case 'start_level':
      return action.level ? `Начать Уровень ${action.level}` : 'Начать следующий Уровень'
    case 'start_free_play':
      return 'Открыть Свободную игру'
  }
}
