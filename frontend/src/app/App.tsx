import { useState } from 'react'
import { Navigate, Outlet, Route, Routes, useParams } from 'react-router-dom'
import {
  CurrentAccountProvider,
  useCurrentAccount,
  useGetMeQuery,
  useUpdateTrainingRoleMutation,
  type Account,
  type UserRole,
} from '@/entities/user'
import { AuthPage } from '@/pages/auth'
import { DashboardPage } from '@/pages/dashboard'
import { LessonsPage, QuizPage, TheoryPage } from '@/pages/lessons'
import { AchievementsPage, ProgressPage } from '@/pages/profile'
import { ChatTrainingPage, ResultPage, TrainingPage } from '@/pages/training'
import { PreviewModeProvider } from '@/shared/runtime-mode'
import { AppHeader } from '@/widgets/app-header'
import {
  createPreviewDashboard,
  createPreviewProgress,
  createPreviewTheory,
  createPreviewTopics,
  previewAccount,
  previewAchievements,
  previewFreePlaySession,
  previewLevels,
  previewQuiz,
  previewQuizOutcome,
  previewResult,
  previewSession,
} from './preview/data'
import styles from './App.module.scss'

function LoadingScreen() {
  return <div className={styles.splash}>Загружаем тренажёр безопасности…</div>
}

function ProtectedLayout() {
  const { data: account, isLoading, isError } = useGetMeQuery()
  const [updateRole] = useUpdateTrainingRoleMutation()

  if (isLoading) return <LoadingScreen />
  if (isError || !account) return <Navigate to="/login" replace />

  const changeTrainingRole = async (role: UserRole) => {
    await updateRole(role).unwrap()
  }

  return (
    <CurrentAccountProvider value={{ account, changeTrainingRole }}>
      <div>
        <AppHeader account={account} />
        <main className={styles.page}>
          <Outlet />
        </main>
      </div>
    </CurrentAccountProvider>
  )
}

function AuthOnly({ mode }: { mode: 'login' | 'register' }) {
  const { data: account, isLoading } = useGetMeQuery()

  if (isLoading) return <LoadingScreen />

  return account ? <Navigate to="/dashboard" replace /> : <AuthPage mode={mode} />
}

function PreviewLayout() {
  const [account, setAccount] = useState<Account>(previewAccount)
  const changeTrainingRole = async (role: UserRole) => {
    setAccount((current) => ({ ...current, trainingRole: role }))
  }

  return (
    <PreviewModeProvider>
      <CurrentAccountProvider value={{ account, changeTrainingRole }}>
        <div>
          <AppHeader account={account} basePath="/preview" />
          <main className={styles.page}>
            <Outlet />
          </main>
        </div>
      </CurrentAccountProvider>
    </PreviewModeProvider>
  )
}

function PreviewAuth({ mode }: { mode: 'login' | 'register' }) {
  return (
    <PreviewModeProvider>
      <AuthPage mode={mode} />
    </PreviewModeProvider>
  )
}

function PreviewDashboardRoute() {
  const { account } = useCurrentAccount()
  return <DashboardPage previewDashboard={createPreviewDashboard(account.trainingRole)} />
}

function PreviewLessonsRoute() {
  const { account } = useCurrentAccount()
  return <LessonsPage previewTopics={createPreviewTopics(account.trainingRole)} />
}

function PreviewTheoryRoute() {
  const { account } = useCurrentAccount()
  const { lessonId } = useParams()
  const topics = createPreviewTopics(account.trainingRole)
  const topic = topics.find((item) => String(item.id) === lessonId || item.slug === lessonId)

  return <TheoryPage previewTheory={createPreviewTheory(account.trainingRole, topic?.id)} />
}

function PreviewQuizRoute() {
  const { account } = useCurrentAccount()
  const { lessonId } = useParams()
  const topics = createPreviewTopics(account.trainingRole)
  const topic = topics.find((item) => String(item.id) === lessonId || item.slug === lessonId)

  return (
    <QuizPage
      previewQuiz={previewQuiz}
      previewOutcome={previewQuizOutcome}
      previewTopicId={topic?.id ?? 1}
    />
  )
}

function PreviewTrainingRoute() {
  const { account } = useCurrentAccount()
  return (
    <TrainingPage
      preview={{
        topics: createPreviewTopics(account.trainingRole),
        levels: previewLevels,
        session: previewSession,
        result: previewResult,
      }}
    />
  )
}

function PreviewProgressRoute() {
  const { account } = useCurrentAccount()
  return <ProgressPage previewProgress={createPreviewProgress(account.trainingRole)} />
}

function PreviewChatRoute() {
  const { sessionId } = useParams()
  const session = sessionId === 'free-play' ? previewFreePlaySession : previewSession

  return <ChatTrainingPage preview={{ session, result: previewResult }} />
}

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<AuthOnly mode="login" />} />
      <Route path="/register" element={<AuthOnly mode="register" />} />

      <Route path="/preview/login" element={<PreviewAuth mode="login" />} />
      <Route path="/preview/register" element={<PreviewAuth mode="register" />} />
      <Route element={<PreviewLayout />}>
        <Route path="/preview/dashboard" element={<PreviewDashboardRoute />} />
        <Route path="/preview/lessons" element={<PreviewLessonsRoute />} />
        <Route path="/preview/lessons/:lessonId" element={<PreviewTheoryRoute />} />
        <Route path="/preview/lessons/:lessonId/quiz" element={<PreviewQuizRoute />} />
        <Route path="/preview/chats" element={<PreviewTrainingRoute />} />
        <Route path="/preview/sessions/:sessionId" element={<PreviewChatRoute />} />
        <Route
          path="/preview/sessions/:sessionId/result"
          element={<ResultPage previewResult={previewResult} />}
        />
        <Route path="/preview/progress" element={<PreviewProgressRoute />} />
        <Route
          path="/preview/achievements"
          element={<AchievementsPage previewAchievements={previewAchievements} />}
        />
      </Route>
      <Route path="/preview" element={<Navigate to="/preview/dashboard" replace />} />

      <Route element={<ProtectedLayout />}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/lessons" element={<LessonsPage />} />
        <Route path="/lessons/:lessonId" element={<TheoryPage />} />
        <Route path="/lessons/:lessonId/quiz" element={<QuizPage />} />
        <Route path="/chats" element={<TrainingPage />} />
        <Route path="/sessions/:sessionId" element={<ChatTrainingPage />} />
        <Route path="/sessions/:sessionId/result" element={<ResultPage />} />
        <Route path="/progress" element={<ProgressPage />} />
        <Route path="/achievements" element={<AchievementsPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}
