import { describe, expect, it } from 'vitest'
import { mapTopic } from './mappers'

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
