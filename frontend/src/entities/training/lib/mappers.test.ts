import { describe, expect, it } from 'vitest'
import { mapAttemptResult, mapLevelState, mapTrainingSession } from './mappers'

describe('mapLevelState', () => {
  it('preserves the complete backend card contract', () => {
    expect(
      mapLevelState({
        number: 3,
        opened: true,
        scenario_id: 13,
        scenario_title: 'Ответ своими словами',
        scenario_description: 'Сформулируйте безопасный ответ.',
        response_type: 'mixed',
        in_progress_attempt_id: 42,
      }),
    ).toEqual({
      number: 3,
      isOpened: true,
      scenarioId: 13,
      scenarioTitle: 'Ответ своими словами',
      scenarioDescription: 'Сформулируйте безопасный ответ.',
      responseMode: 'mixed',
      inProgressAttemptId: 42,
    })
  })
})

describe('mapAttemptResult', () => {
  it('keeps structured feedback and optional micro-question', () => {
    const result = mapAttemptResult({
      attempt_id: 9,
      score: 75,
      stars: 2,
      decision_review: [
        {
          step_id: 1,
          step_number: 1,
          answer_type: 'free_text',
          free_text: 'Проверю в приложении',
          points: 75,
          assessment: 'mostly_safe',
          explanation: 'Верное направление.',
          safe_action: 'Оставайтесь в сервисе.',
          risk_signals: [{ code: 'external_link', label: 'Внешняя ссылка' }],
        },
      ],
      risk_signals: [{ code: 'external_link', label: 'Внешняя ссылка' }],
      safe_actions: ['Оставайтесь в сервисе.'],
      feedback: {
        reason: 'В чате была внешняя ссылка.',
        risk_signals: [{ code: 'external_link', label: 'Внешняя ссылка' }],
        safe_alternative: 'Откройте заказ в приложении.',
      },
      micro_question: {
        pattern_code: 'external_link',
        question: 'Где проверить сделку?',
        options: ['По ссылке', 'В приложении'],
      },
      level_progress: {
        number: 1,
        opened: true,
        best_score: 75,
        stars: 2,
        attempts: 1,
        last_attempt_id: 9,
      },
      topic_id: 1,
      topic_completed: false,
      next_action: { type: 'read_theory', topic_id: 1 },
      new_achievements: [],
      streak: {
        current: 1,
        longest: 1,
        last_activity_date: '2026-08-13',
        active_today: true,
      },
    })

    expect(result.feedback.riskSignals[0].label).toBe('Внешняя ссылка')
    expect(result.microQuestion?.options).toEqual(['По ссылке', 'В приложении'])
    expect(result.nextAction).toEqual({
      type: 'read_theory',
      topicId: 1,
      level: undefined,
      attemptId: undefined,
    })
  })
})

describe('mapTrainingSession', () => {
  it('maps a resumable backend attempt into the training domain', () => {
    const session = mapTrainingSession({
      attempt_id: 9,
      status: 'IN_PROGRESS',
      scenario_id: 3,
      scenario_title: 'Сценарий',
      scenario_description: 'Описание',
      topic_id: 1,
      topic_title: 'Фишинговые ссылки',
      level: 2,
      user_role: 'buyer',
      counterparty_role: 'seller',
      product_context: {
        item_title: 'Велосипед',
        category: 'Спорт',
        deal_method: 'delivery',
      },
      mode: 'multiple_choice',
      step_progress: { current: 2, answered: 1, total: 4 },
      step: {
        id: 12,
        number: 2,
        counterparty_message: 'Сообщение',
        options: [{ id: 5, text: 'Ответ' }],
      },
      answers: [{ step_id: 11, answer_type: 'option', option_id: 4, points: 100 }],
      messages: [{ role: 'assistant', text: 'Сообщение' }],
      can_finish_early: false,
    })

    expect(session.attemptId).toBe(9)
    expect(session.progress).toEqual({ currentStep: 2, answeredSteps: 1, totalSteps: 4 })
    expect(session.step.options[0]).toEqual({ id: 5, text: 'Ответ' })
  })
})
