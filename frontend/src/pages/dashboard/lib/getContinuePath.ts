import type { ContinueAction } from '@/entities/progress'

export function getContinuePath(action: ContinueAction | null, basePath: string): string {
  if (!action) return `${basePath}/lessons`

  switch (action.type) {
    case 'resume_attempt':
      return action.attemptId ? `${basePath}/sessions/${action.attemptId}` : `${basePath}/chats`
    case 'read_theory':
      return action.topicId ? `${basePath}/lessons/${action.topicId}` : `${basePath}/lessons`
    case 'take_quiz':
      return action.topicId ? `${basePath}/lessons/${action.topicId}/quiz` : `${basePath}/lessons`
    case 'start_level':
    case 'start_free_play':
      return action.topicId ? `${basePath}/chats?topic=${action.topicId}` : `${basePath}/chats`
  }
}
