import type { AttemptResultDto, GameStateDto, LevelStateDto } from '../api/contracts'
import type { AttemptResult, LevelState, TrainingSession } from '../model/types'
import { mapStreak } from '@/entities/user'

export function mapLevelState(dto: LevelStateDto): LevelState {
  return {
    number: dto.number,
    isOpened: dto.opened,
    scenarioId: dto.scenario_id,
    scenarioTitle: dto.scenario_title,
    scenarioDescription: dto.scenario_description,
    responseMode: dto.response_type,
    inProgressAttemptId: dto.in_progress_attempt_id,
  }
}

export function mapTrainingSession(dto: GameStateDto): TrainingSession {
  return {
    attemptId: dto.attempt_id,
    status: dto.status,
    scenarioId: dto.scenario_id,
    scenarioTitle: dto.scenario_title,
    scenarioDescription: dto.scenario_description,
    topicId: dto.topic_id,
    topicTitle: dto.topic_title,
    level: dto.level,
    userRole: dto.user_role,
    counterpartyRole: dto.counterparty_role,
    productContext: {
      itemTitle: dto.product_context.item_title,
      category: dto.product_context.category,
      dealMethod: dto.product_context.deal_method,
      price: dto.product_context.price,
      currency: dto.product_context.currency,
      location: dto.product_context.location,
      imageKey: dto.product_context.image_key,
    },
    mode: dto.mode,
    progress: {
      currentStep: dto.step_progress.current,
      answeredSteps: dto.step_progress.answered,
      totalSteps: dto.step_progress.total,
    },
    step: {
      id: dto.step.id,
      number: dto.step.number,
      counterpartyMessage: dto.step.counterparty_message,
      options: dto.step.options,
    },
    answers: dto.answers.map((answer) => ({
      stepId: answer.step_id,
      answerType: answer.answer_type,
      optionId: answer.option_id,
      optionText: answer.option_text,
      freeText: answer.free_text,
      points: answer.points,
    })),
    messages: dto.messages,
    canFinishEarly: dto.can_finish_early,
  }
}

export function mapAttemptResult(dto: AttemptResultDto): AttemptResult {
  return {
    attemptId: dto.attempt_id,
    score: dto.score,
    stars: dto.stars,
    decisionReview: dto.decision_review.map((answer) => ({
      stepId: answer.step_id,
      stepNumber: answer.step_number,
      answerType: answer.answer_type,
      optionId: answer.option_id,
      optionText: answer.option_text,
      freeText: answer.free_text,
      points: answer.points,
      assessment: answer.assessment,
      explanation: answer.explanation,
      safeAction: answer.safe_action,
      riskSignals: answer.risk_signals,
    })),
    riskSignals: dto.risk_signals,
    safeActions: dto.safe_actions,
    feedback: {
      reason: dto.feedback.reason,
      riskSignals: dto.feedback.risk_signals,
      safeAlternative: dto.feedback.safe_alternative,
    },
    microQuestion: dto.micro_question
      ? {
          patternCode: dto.micro_question.pattern_code,
          question: dto.micro_question.question,
          options: dto.micro_question.options,
        }
      : undefined,
    levelProgress: {
      number: dto.level_progress.number,
      isOpened: dto.level_progress.opened,
      bestScore: dto.level_progress.best_score,
      stars: dto.level_progress.stars,
      attempts: dto.level_progress.attempts,
      lastAttemptId: dto.level_progress.last_attempt_id,
    },
    topicId: dto.topic_id,
    isTopicCompleted: dto.topic_completed,
    nextAction: dto.next_action
      ? {
          type: dto.next_action.type,
          topicId: dto.next_action.topic_id,
          level: dto.next_action.level,
          attemptId: dto.next_action.attempt_id,
        }
      : null,
    newAchievements: dto.new_achievements.map((achievement) => ({
      code: achievement.code,
      title: achievement.title,
      description: achievement.description,
      icon: achievement.icon,
    })),
    streak: mapStreak(dto.streak),
    isScam: dto.is_scam,
  }
}
