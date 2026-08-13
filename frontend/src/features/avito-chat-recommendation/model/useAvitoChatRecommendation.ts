import { useRecommendFromAvitoChatMutation } from '@/entities/learning'
import type { UserRole } from '@/entities/user'
import { getApiErrorMessage } from '@/shared/http-error'

const safeDemoSnapshot = [
  { role: 'assistant' as const, text: 'Велосипед ещё продаётся?' },
  { role: 'user' as const, text: 'Да, ещё актуален.' },
  {
    role: 'assistant' as const,
    text: 'Я уже оплатил доставку. Подтвердите получение на внешней странице.',
  },
]

export function useAvitoChatRecommendation(role: UserRole) {
  const [recommend, state] = useRecommendFromAvitoChatMutation()

  const submit = async () => {
    try {
      await recommend({
        source: 'avito_chat_demo',
        role,
        messages: safeDemoSnapshot,
        risk_type: 'phishing',
        risk_signals: ['внешняя страница', 'просьба подтвердить получение'],
      }).unwrap()
    } catch {
      // The RTK Query error is exposed below and the prepared Snapshot stays only in memory.
    }
  }

  return {
    recommendation: state.data,
    isLoading: state.isLoading,
    error: state.error ? getApiErrorMessage(state.error) : '',
    submit,
    snapshot: safeDemoSnapshot,
  }
}
