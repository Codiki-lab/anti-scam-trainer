import type { Quiz, QuizOutcome, Theory, Topic } from '@/entities/learning'
import type { Achievements, Dashboard, Progress } from '@/entities/progress'
import type { AttemptResult, LevelState, TrainingSession } from '@/entities/training'
import type { Account, UserRole } from '@/entities/user'

export const previewAccount: Account = {
  id: 1,
  username: 'demo',
  accessRole: 'user',
  trainingRole: 'buyer',
  streak: { current: 3, longest: 7, isActiveToday: true, lastActivityDate: '2026-08-09' },
}

const topicDefinitions = [
  ['phishing-links', 'Фишинговые ссылки', 'Распознавайте поддельные страницы оплаты и доставки.'],
  ['prepayment', 'Предоплата', 'Проверяйте просьбы об оплате товара до получения.'],
  ['fake-delivery', 'Поддельная доставка', 'Отличайте штатную доставку от мошеннической.'],
  ['off-platform', 'Общение вне Avito', 'Сохраняйте переписку и доказательства внутри сервиса.'],
  ['sms-codes', 'Коды из SMS', 'Не передавайте секретные коды собеседникам.'],
  [
    'too-good-offer',
    'Слишком выгодное предложение',
    'Замечайте давление и нереалистичные условия.',
  ],
] as const

export function createPreviewTopics(role: UserRole): Topic[] {
  return topicDefinitions.map(([slug, title, description], index) => ({
    id: index + 1,
    slug,
    role,
    title,
    description,
    order: index + 1,
    isTheoryRead: index === 0,
    isQuizPassed: index === 0,
    bestQuizScore: index === 0 ? 80 : 0,
    isCompleted: index === 0,
    levels: [1, 2, 3, 4].map((number) => ({
      number,
      isOpened: index === 0,
      bestScore: index === 0 ? 75 : 0,
      stars: index === 0 ? 2 : 0,
      attempts: index === 0 ? 1 : 0,
      lastAttemptId: index === 0 ? 9000 + number : null,
    })),
  }))
}

export function createPreviewTrainingTopics(role: UserRole): Topic[] {
  return createPreviewTopics(role).map((topic, topicIndex) => {
    if (topicIndex !== 0) return topic

    return {
      ...topic,
      isCompleted: false,
      levels: topic.levels.map((level) => ({
        ...level,
        isOpened: level.number <= 2,
        bestScore: level.number === 1 ? 75 : 0,
        stars: level.number === 1 ? 2 : 0,
        attempts: level.number === 1 ? 1 : 0,
        lastAttemptId: level.number === 1 ? 9001 : null,
      })),
    }
  })
}

export function createPreviewTheory(role: UserRole, topicId = 1): Theory {
  const topics = createPreviewTopics(role)
  return {
    topic: topics.find((topic) => topic.id === topicId) ?? topics[0],
    sections: [
      {
        id: 1,
        order: 1,
        kind: 'rule',
        title: 'Главное правило',
        body: 'Оплата и доставка оформляются только внутри сервиса.',
      },
      {
        id: 2,
        order: 2,
        kind: 'signs',
        title: 'Как распознать риск',
        body: 'Вас торопят, присылают внешнюю ссылку или просят секретные данные.',
      },
      {
        id: 3,
        order: 3,
        kind: 'action',
        title: 'Что сделать безопасно',
        body: 'Остановитесь и проверьте действие в приложении.',
      },
      {
        id: 4,
        order: 4,
        kind: 'example',
        title: 'Пример',
        body: 'Собеседник предлагает оформить доставку на стороннем сайте.',
      },
      {
        id: 5,
        order: 5,
        kind: 'summary',
        title: 'Запомните',
        body: 'Не открывайте платёжные ссылки из переписки.',
      },
    ],
  }
}

export const previewQuiz: Quiz = {
  passThreshold: 80,
  questions: Array.from({ length: 5 }, (_, index) => ({
    id: index + 1,
    order: index + 1,
    text: 'Что безопаснее сделать, если собеседник прислал ссылку для оплаты?',
    choices: [
      { id: 1, text: 'Перейти по ссылке и проверить её' },
      { id: 2, text: 'Попросить прислать ссылку ещё раз' },
      { id: 3, text: 'Оформить оплату только внутри сервиса' },
      { id: 4, text: 'Перейти в другой мессенджер' },
    ],
  })),
}

export const previewQuizOutcome: QuizOutcome = {
  score: 100,
  isPassed: true,
  bestScore: 100,
  isFirstPass: true,
  streak: previewAccount.streak,
}

export const previewLevels: LevelState[] = [
  {
    number: 1,
    isOpened: true,
    scenarioId: 101,
    scenarioTitle: 'Безопасная ссылка на оплату',
    scenarioDescription: 'Выберите безопасный ответ в переписке о доставке.',
    responseMode: 'multiple_choice',
  },
  {
    number: 2,
    isOpened: true,
    scenarioId: 102,
    scenarioTitle: 'Похожие варианты ответа',
    scenarioDescription: 'Найдите безопасное действие среди похожих формулировок.',
    responseMode: 'similar_choice',
  },
  {
    number: 3,
    isOpened: false,
    scenarioId: 103,
    scenarioTitle: 'Ответ своими словами',
    scenarioDescription: 'Сначала выберите вариант, затем сформулируйте ответ самостоятельно.',
    responseMode: 'mixed',
  },
  {
    number: 4,
    isOpened: false,
    scenarioId: 104,
    scenarioTitle: 'Свободный диалог',
    scenarioDescription: 'Ведите разговор самостоятельно и остановите опасную сделку.',
    responseMode: 'free_text',
  },
]

export const previewSession: TrainingSession = {
  attemptId: 9001,
  status: 'IN_PROGRESS',
  scenarioId: 101,
  topicId: 1,
  productContext: { item: 'Смартфон', price: 42000 },
  mode: 'multiple_choice',
  progress: { currentStep: 2, answeredSteps: 1 },
  step: {
    id: 2,
    number: 2,
    counterpartyMessage: 'Я уже почти оплатил. Вот ссылка, подтвердите получение денег.',
    options: [
      { id: 1, text: 'Хорошо, пришлите ссылку' },
      { id: 2, text: 'Оформим доставку только через приложение Avito' },
      { id: 3, text: 'Давайте перейдём в Telegram' },
      { id: 4, text: 'Лучше встретимся лично без проверки' },
    ],
  },
  answers: [{ stepId: 1, optionId: 2 }],
  messages: [
    { role: 'assistant', text: 'Здравствуйте! Товар ещё актуален?' },
    { role: 'user', text: 'Да, актуален.' },
    { role: 'assistant', text: 'Я уже почти оплатил. Вот ссылка, подтвердите получение денег.' },
  ],
  canFinishEarly: false,
}

export const previewFreePlaySession: TrainingSession = {
  ...previewSession,
  attemptId: 9002,
  scenarioId: 0,
  topicId: 0,
  mode: 'free_text',
  productContext: { item: 'Игровая приставка', price: 39000, difficulty: 'adaptive' },
  progress: { currentStep: 1, answeredSteps: 0 },
  step: {
    id: 1,
    number: 1,
    counterpartyMessage:
      'Здравствуйте! Можно сегодня забрать товар? Как вам удобнее оформить сделку?',
    options: [],
  },
  answers: [],
  messages: [
    {
      role: 'assistant',
      text: 'Здравствуйте! Можно сегодня забрать товар? Как вам удобнее оформить сделку?',
    },
  ],
}

export const previewResult: AttemptResult = {
  attemptId: 9001,
  score: 75,
  stars: 3,
  decisionReview: [
    {
      stepId: 1,
      optionId: 2,
      optionText: 'Оформим доставку только внутри сервиса',
      points: 100,
      explanation: 'Безопасный способ.',
      riskSignals: [],
    },
    {
      stepId: 2,
      optionId: 1,
      optionText: 'Проверю ссылку позже',
      points: 50,
      explanation: 'Ссылку лучше не открывать.',
      riskSignals: ['external_link'],
    },
  ],
  riskSignals: ['external_link'],
  safeActions: ['Оставаться внутри сервиса'],
  levelProgress: createPreviewTopics('buyer')[0].levels[0],
  topicId: 1,
  isTopicCompleted: false,
  nextAction: null,
  newAchievements: [],
  streak: previewAccount.streak,
  isScam: true,
}

export const previewAchievements: Achievements = {
  earned: [
    {
      code: 'first_training',
      title: 'Первое прохождение',
      description: 'Завершить первое Прохождение.',
      icon: 'star',
      earned: true,
      earnedAt: '2026-08-08T10:00:00Z',
      current: 1,
      target: 1,
    },
    {
      code: 'perfect_score',
      title: 'Без ошибки',
      description: 'Получить 100 Баллов.',
      icon: 'shield',
      earned: true,
      earnedAt: '2026-08-09T10:00:00Z',
      current: 100,
      target: 100,
    },
    {
      code: 'first_topic_completed',
      title: 'Первая Тема',
      description: 'Завершить первую Тему.',
      icon: 'book',
      earned: true,
      earnedAt: '2026-08-09T11:00:00Z',
      current: 1,
      target: 1,
    },
    {
      code: 'streak_3',
      title: 'Серия 3 дня',
      description: 'Заниматься три дня подряд.',
      icon: 'flame',
      earned: true,
      earnedAt: '2026-08-10T11:00:00Z',
      current: 3,
      target: 3,
    },
  ],
  available: [
    {
      code: 'five_trainings',
      title: 'Пять прохождений',
      description: 'Завершить пять Прохождений.',
      icon: 'stack',
      earned: false,
      current: 3,
      target: 5,
    },
    {
      code: 'all_buyer_topics',
      title: 'Покупатель: все Темы',
      description: 'Завершить шесть Тем покупателя.',
      icon: 'buyer',
      earned: false,
      current: 1,
      target: 6,
    },
    {
      code: 'all_seller_topics',
      title: 'Продавец: все Темы',
      description: 'Завершить шесть Тем продавца.',
      icon: 'seller',
      earned: false,
      current: 0,
      target: 6,
    },
    {
      code: 'streak_7',
      title: 'Серия 7 дней',
      description: 'Заниматься семь дней подряд.',
      icon: 'flame',
      earned: false,
      current: 3,
      target: 7,
    },
  ],
}

export function createPreviewDashboard(role: UserRole): Dashboard {
  return {
    profile: { id: previewAccount.id, username: previewAccount.username, trainingRole: role },
    streak: previewAccount.streak,
    topics: createPreviewTopics(role),
    achievements: previewAchievements.earned,
    continueAction: { type: 'start_level', topicId: 1, level: 2 },
    dailyTask: {
      date: '2026-08-09',
      role,
      messages: [
        { role: 'assistant', text: 'Покупатель просит перейти по ссылке для оплаты доставки.' },
        { role: 'user', text: 'Ссылка выглядит как страница сервиса объявлений.' },
      ],
      isCompleted: false,
      signals: [],
    },
  }
}

export function createPreviewProgress(role: UserRole): Progress {
  return {
    role,
    summary: {
      completedTopics: 1,
      totalTopics: 6,
      completedLevels: 4,
      totalLevels: 24,
      stars: 8,
      averageScore: 76,
    },
    topics: createPreviewTopics(role),
    recentAttempts: [
      {
        attemptId: 9001,
        topicId: 1,
        level: 2,
        score: 75,
        stars: 2,
        finishedAt: '2026-08-09T09:00:00Z',
      },
      {
        attemptId: 9000,
        topicId: 5,
        level: 1,
        score: 100,
        stars: 3,
        finishedAt: '2026-08-08T09:00:00Z',
      },
    ],
  }
}
