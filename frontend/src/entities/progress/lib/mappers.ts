import { mapTopic } from '@/entities/learning'
import type { AchievementDto, AchievementsDto, DashboardDto, ProgressDto } from '../api/contracts'
import type { Achievement, Achievements, Dashboard, Progress } from '../model/types'
import { mapStreak } from '@/entities/user'

export const mapAchievement = (dto: AchievementDto): Achievement => ({
  code: dto.code,
  title: dto.title,
  description: dto.description,
  icon: dto.icon,
  earned: dto.earned,
  earnedAt: dto.earned_at,
  current: dto.progress.current,
  target: dto.progress.target,
})

export const mapDashboard = (dto: DashboardDto): Dashboard => ({
  profile: {
    id: dto.profile.id,
    username: dto.profile.username,
    trainingRole: dto.profile.training_role,
  },
  streak: mapStreak(dto.streak),
  topics: dto.topics.map(mapTopic),
  achievements: dto.achievements.map(mapAchievement),
  continueAction: dto.continue_action && {
    type: dto.continue_action.type,
    topicId: dto.continue_action.topic_id,
    level: dto.continue_action.level,
    attemptId: dto.continue_action.attempt_id,
  },
})

export const mapProgress = (dto: ProgressDto): Progress => ({
  role: dto.role,
  summary: {
    completedTopics: dto.summary.completed_topics,
    totalTopics: dto.summary.total_topics,
    completedLevels: dto.summary.completed_levels,
    totalLevels: dto.summary.total_levels,
    stars: dto.summary.stars,
    averageScore: dto.summary.average_score,
  },
  topics: dto.topics.map(mapTopic),
  recentAttempts: dto.recent_attempts.map((attempt) => ({
    attemptId: attempt.attempt_id,
    topicId: attempt.topic_id,
    level: attempt.level,
    score: attempt.score,
    stars: attempt.stars,
    finishedAt: attempt.finished_at,
  })),
})

export const mapAchievements = (dto: AchievementsDto): Achievements => ({
  earned: dto.earned.map(mapAchievement),
  available: dto.available.map(mapAchievement),
})
