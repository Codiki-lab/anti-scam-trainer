import type {
  Quiz as QuizDto,
  QuizResult as QuizResultDto,
  TheoryResponse as TheoryResponseDto,
  TopicContract as TopicDto,
  RecommendationDto,
  SkillCheckDto,
} from '../api/contracts'
import type {
  LearningRecommendation,
  Quiz,
  QuizOutcome,
  SkillCheck,
  Theory,
  Topic,
} from '../model/types'
import { mapStreak } from '@/entities/user'

export function mapTopic(dto: TopicDto): Topic {
  return {
    id: dto.id,
    slug: dto.slug,
    role: dto.role,
    title: dto.title,
    description: dto.description,
    order: dto.sort_order,
    isTheoryRead: dto.theory_read,
    isQuizPassed: dto.quiz_passed,
    bestQuizScore: dto.quiz_best_score,
    isCompleted: dto.completed,
    levels: dto.levels.map((level) => ({
      number: level.number,
      isOpened: level.opened,
      bestScore: level.best_score,
      stars: level.stars,
      attempts: level.attempts,
      lastAttemptId: level.last_attempt_id,
    })),
  }
}

export function mapTheory(dto: TheoryResponseDto): Theory {
  return {
    topic: mapTopic(dto.topic),
    sections: dto.blocks.map((block) => ({
      id: block.id,
      order: block.sort_order,
      kind: block.kind,
      title: block.title,
      body: block.body,
    })),
  }
}

export function mapQuiz(dto: QuizDto): Quiz {
  return {
    passThreshold: dto.pass_threshold,
    questions: dto.questions.map((question) => ({
      id: question.id,
      order: question.sort_order,
      text: question.text,
      choices: question.options.map((option) => ({ id: option.id, text: option.text })),
    })),
  }
}

export function mapQuizOutcome(dto: QuizResultDto): QuizOutcome {
  return {
    score: dto.score,
    isPassed: dto.passed,
    bestScore: dto.best_score,
    isFirstPass: dto.newly_passed,
    streak: mapStreak(dto.streak),
  }
}

export function mapRecommendation(dto: RecommendationDto): LearningRecommendation {
  return {
    topic: mapTopic(dto.topic),
    explanation: dto.explanation,
    nextAction: {
      type: dto.next_action.type,
      topicId: dto.next_action.topic_id,
      level: dto.next_action.level,
      attemptId: dto.next_action.attempt_id,
    },
    isFallback: dto.fallback,
  }
}

export function mapSkillCheck(dto: SkillCheckDto): SkillCheck {
  return {
    id: dto.id,
    topicId: dto.topic_id,
    phase: dto.phase,
    snapshot: dto.snapshot,
    beforeCorrect: dto.before_correct,
    afterCorrect: dto.after_correct,
    isVerdictImproved: dto.verdict_improved,
    beforePattern: dto.before_pattern,
    afterPattern: dto.after_pattern,
    isPatternImproved: dto.pattern_improved,
    isImproved: dto.improved,
  }
}
