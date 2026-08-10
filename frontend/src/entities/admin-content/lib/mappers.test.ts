import { describe, expect, it } from 'vitest'
import { adminScenarioDtoSchema, adminTopicDtoSchema } from '../api/contracts'
import { mapAdminScenario, mapAdminScenarioDto, mapAdminTopic, mapAdminTopicDto } from './mappers'

describe('admin content contracts', () => {
  it('maps a nested Topic response into the UI domain', () => {
    const topic = mapAdminTopic(
      adminTopicDtoSchema.parse({
        id: 7,
        slug: 'phishing-links',
        role: 'buyer',
        title: 'Фишинговые ссылки',
        description: 'Описание',
        sort_order: 1,
        status: 'draft',
        theory: [{ id: 8, sort_order: 1, kind: 'intro', title: 'Введение', body: 'Текст' }],
        quiz: [
          {
            id: 9,
            sort_order: 1,
            text: 'Вопрос',
            explanation: 'Объяснение',
            options: [{ id: 10, sort_order: 1, text: 'Ответ', is_correct: true }],
          },
        ],
      }),
    )

    expect(topic.sortOrder).toBe(1)
    expect(topic.theory[0]).toMatchObject({ id: 8, sortOrder: 1, kind: 'intro' })
    expect(topic.quiz[0].options[0]).toMatchObject({ id: 10, isCorrect: true })
  })

  it('sends only writable Topic fields to the backend', () => {
    expect(
      mapAdminTopicDto({
        slug: 'safe-payment',
        role: 'seller',
        title: 'Безопасная оплата',
        description: 'Описание',
        sortOrder: 2,
      }),
    ).toEqual({
      slug: 'safe-payment',
      role: 'seller',
      title: 'Безопасная оплата',
      description: 'Описание',
      sort_order: 2,
    })
  })

  it('maps Scenario JSON fields and excludes read-only fields from writes', () => {
    const scenario = mapAdminScenario(
      adminScenarioDtoSchema.parse({
        id: 11,
        title: 'Проверка сделки',
        description: 'Описание',
        level_id: 2,
        topic_id: 7,
        role: 'buyer',
        status: 'published',
        scam_scheme: 'fake-payment',
        product_context: { category: 'electronics' },
        ai_system_prompt: 'Инструкция',
        final_rubric: { safe: 100 },
      }),
    )

    expect(scenario.productContext).toEqual({ category: 'electronics' })
    expect(mapAdminScenarioDto(scenario)).toEqual({
      title: 'Проверка сделки',
      description: 'Описание',
      level_id: 2,
      topic_id: 7,
      role: 'buyer',
      scam_scheme: 'fake-payment',
      product_context: { category: 'electronics' },
      ai_system_prompt: 'Инструкция',
      final_rubric: { safe: 100 },
    })
  })
})
