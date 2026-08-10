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
    isTheoryRead: index < 2,
    isQuizPassed: index === 0,
    bestQuizScore: index === 0 ? 80 : 0,
    isCompleted: false,
    levels: [1, 2, 3, 4].map((number) => ({
      number,
      isOpened: index === 0 && number <= 2,
      bestScore: number === 1 ? 75 : 0,
      stars: number === 1 ? 2 : 0,
      attempts: number === 1 ? 1 : 0,
      lastAttemptId: null,
    })),
  }))
}

export function createPreviewTheory(role: UserRole, topicId = 1): Theory {
  const topics = createPreviewTopics(role)
  return {
    topic: topics.find((topic) => topic.id === topicId) ?? topics[0],
    sections: [
      {
        id: 1,
        order: 1,
        kind: 'intro',
        title: 'Чему вы научитесь',
        body: 'Оплата и доставка оформляются только внутри сервиса.',
      },
      {
        id: 2,
        order: 2,
        kind: 'risk',
        title: 'Как распознать риск',
        body: 'Вас торопят, присылают внешнюю ссылку или просят секретные данные.',
      },
      {
        id: 3,
        order: 3,
        kind: 'example',
        title: 'Пример переписки',
        body: 'Собеседник предлагает оформить доставку на стороннем сайте.',
      },
      {
        id: 4,
        order: 4,
        kind: 'safe_action',
        title: 'Что сделать безопасно',
        body: 'Остановитесь. Проверьте действие в приложении. Обратитесь в официальную поддержку.',
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
    scenarioTitle: 'Заметьте сигнал',
    scenarioDescription: 'Выберите однозначно безопасный ответ.',
    responseType: 'multiple_choice',
  },
  {
    number: 2,
    isOpened: true,
    scenarioId: 102,
    scenarioTitle: 'Сравните варианты',
    scenarioDescription: 'Найдите наиболее безопасную формулировку.',
    responseType: 'similar_choice',
    inProgressAttemptId: 9001,
  },
  {
    number: 3,
    isOpened: false,
    scenarioId: 103,
    scenarioTitle: 'Ответьте своими словами',
    scenarioDescription: 'Сочетайте выбор и свободный текст.',
    responseType: 'mixed',
  },
  {
    number: 4,
    isOpened: false,
    scenarioId: 104,
    scenarioTitle: 'Ведите диалог',
    scenarioDescription: 'Самостоятельно проведите безопасную сделку.',
    responseType: 'free_text',
  },
]

export const previewSession: TrainingSession = {
  attemptId: 9001,
  status: 'IN_PROGRESS',
  scenarioId: 101,
  scenarioTitle: 'Безопасная продажа смартфона',
  scenarioDescription: 'Проверьте факт оплаты и не уходите из штатного оформления.',
  topicId: 1,
  topicTitle: 'Поддельная оплата',
  level: 2,
  userRole: 'seller',
  counterpartyRole: 'buyer',
  productContext: {
    itemTitle: 'Смартфон',
    category: 'Электроника',
    dealMethod: 'delivery',
    price: 42000,
    currency: 'RUB',
    location: 'Москва',
    imageKey: 'smartphone',
  },
  mode: 'multiple_choice',
  progress: { currentStep: 2, answeredSteps: 1, totalSteps: 2 },
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
  answers: [
    {
      stepId: 1,
      answerType: 'option',
      optionId: 2,
      optionText: 'Оформим доставку только через приложение Avito',
      points: 100,
    },
  ],
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
  scenarioTitle: 'Свободная игра',
  scenarioDescription: 'Непредсказуемая тренировка безопасной сделки',
  topicId: 0,
  topicTitle: 'Все изученные Темы',
  level: 0,
  mode: 'free_text',
  productContext: {
    itemTitle: 'Игровая приставка',
    category: 'Электроника',
    dealMethod: 'delivery',
    price: 39000,
    currency: 'RUB',
    location: 'Москва',
    imageKey: 'console',
  },
  progress: { currentStep: 1, answeredSteps: 0, totalSteps: 5 },
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
  stars: 2,
  decisionReview: [
    {
      stepId: 1,
      stepNumber: 1,
      answerType: 'option',
      optionId: 2,
      optionText: 'Оформим доставку только внутри сервиса',
      points: 100,
      assessment: 'safe',
      explanation: 'Безопасный способ.',
      safeAction: 'Проверить оформление внутри приложения.',
      riskSignals: [],
    },
    {
      stepId: 2,
      stepNumber: 2,
      answerType: 'option',
      optionId: 1,
      optionText: 'Проверю ссылку позже',
      points: 50,
      assessment: 'risky',
      explanation: 'Ссылку лучше не открывать.',
      safeAction: 'Не открывать ссылку и проверить заказ самостоятельно.',
      riskSignals: [{ code: 'external_link', label: 'Внешняя ссылка вместо штатного экрана' }],
    },
  ],
  riskSignals: [{ code: 'external_link', label: 'Внешняя ссылка вместо штатного экрана' }],
  safeActions: ['Оставаться внутри сервиса'],
  levelProgress: createPreviewTopics('buyer')[0].levels[0],
  topicId: 1,
  isTopicCompleted: false,
  nextAction: null,
  newAchievements: [
    {
      code: 'first_training',
      title: 'Первая тренировка',
      description: 'Завершено первое Прохождение',
      icon: 'star',
    },
  ],
  streak: previewAccount.streak,
}

export const previewAchievements: Achievements = {
  earned: [
    {
      code: 'first_training',
      title: 'Первая тренировка',
      description: 'Пройдите первый сценарий',
      icon: 'star',
      earned: true,
      current: 1,
      target: 1,
    },
  ],
  available: [
    {
      code: 'all_buyer_topics',
      title: 'Безопасный пользователь',
      description: 'Пройдите все темы роли',
      icon: 'buyer',
      earned: false,
      current: 0,
      target: 6,
    },
    {
      code: 'streak_7',
      title: 'Серия 7 дней',
      description: 'Учитесь неделю подряд',
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
      completedTopics: 0,
      totalTopics: 6,
      completedLevels: 1,
      totalLevels: 24,
      stars: 2,
      averageScore: 75,
    },
    topics: createPreviewTopics(role),
    recentAttempts: [
      {
        attemptId: 9000,
        topicId: 1,
        level: 1,
        score: 75,
        stars: 2,
        finishedAt: '2026-08-08T09:00:00Z',
      },
    ],
  }
}
