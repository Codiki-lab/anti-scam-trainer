import { describe, expect, it } from 'vitest'
import { mapTrainingSession } from './mappers'

describe('mapTrainingSession', () => {
  it('maps a resumable backend attempt into the training domain', () => {
    const session = mapTrainingSession({
      attempt_id: 9,
      status: 'IN_PROGRESS',
      scenario_id: 3,
      topic_id: 1,
      product_context: {},
      mode: 'multiple_choice',
      step_progress: { current: 2, answered: 1 },
      step: {
        id: 12,
        number: 2,
        counterparty_message: 'Сообщение',
        options: [{ id: 5, text: 'Ответ' }],
      },
      answers: [{ step_id: 11, option_id: 4 }],
      messages: [{ role: 'assistant', text: 'Сообщение' }],
      can_finish_early: false,
    })

    expect(session.attemptId).toBe(9)
    expect(session.progress).toEqual({ currentStep: 2, answeredSteps: 1 })
    expect(session.step.options[0]).toEqual({ id: 5, text: 'Ответ' })
  })
})
