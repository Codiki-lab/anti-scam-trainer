import { describe, expect, it } from 'vitest'
import { getContinueLearningLabel, getContinueLearningPath } from './getContinueLearningPath'

describe('continue learning target', () => {
  it('restores an active attempt', () => {
    const action = { type: 'resume_attempt' as const, attemptId: 42 }

    expect(getContinueLearningPath(action, '/preview')).toBe('/preview/sessions/42')
    expect(getContinueLearningLabel(action)).toBe('Продолжить тренировку')
  })

  it('keeps the selected topic when starting the next level', () => {
    const action = { type: 'start_level' as const, topicId: 7, level: 3 }

    expect(getContinueLearningPath(action, '')).toBe('/chats?topic=7')
    expect(getContinueLearningLabel(action)).toBe('Начать Уровень 3')
  })

  it('points to the dashboard control that owns free-play start', () => {
    const action = { type: 'start_free_play' as const }

    expect(getContinueLearningPath(action, '')).toBe('/dashboard#free-play')
  })
})
