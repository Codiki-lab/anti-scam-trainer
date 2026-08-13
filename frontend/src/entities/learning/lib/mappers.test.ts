import { describe, expect, it } from 'vitest'
import { mapRecommendation, mapSkillCheck, mapTopic } from './mappers'

describe('mapTopic', () => {
  it('isolates transport fields from the UI model', () => {
    const topic = mapTopic({
      id: 7,
      slug: 'safe-payment',
      role: 'buyer',
      title: 'Безопасная оплата',
      description: 'Описание',
      sort_order: 2,
      theory_read: true,
      quiz_passed: false,
      quiz_best_score: 60,
      completed: false,
      levels: [
        { number: 1, opened: true, best_score: 75, stars: 2, attempts: 1, last_attempt_id: 42 },
      ],
    })

    expect(topic.order).toBe(2)
    expect(topic.isTheoryRead).toBe(true)
    expect(topic.levels[0]).toEqual({
      number: 1,
      isOpened: true,
      bestScore: 75,
      stars: 2,
      attempts: 1,
      lastAttemptId: 42,
    })
  })
})

describe('personal learning mappers', () => {
  it('uses the server action without deriving a client-side topic', () => {
    const topic = {
      id: 7,
      slug: 'safe-payment',
      role: 'buyer' as const,
      title: 'Безопасная оплата',
      description: 'Описание',
      sort_order: 2,
      theory_read: false,
      quiz_passed: false,
      quiz_best_score: 0,
      completed: false,
      levels: [],
    }

    expect(
      mapRecommendation({
        topic,
        explanation: 'Начните с Теории.',
        next_action: { type: 'read_theory', topic_id: 7 },
        fallback: false,
      }).nextAction,
    ).toEqual({ type: 'read_theory', topicId: 7, level: undefined, attemptId: undefined })
  })

  it('preserves the backend skill-check phase', () => {
    expect(mapSkillCheck({ id: 3, topic_id: 7, phase: 'after_locked' }).phase).toBe('after_locked')
  })
})
