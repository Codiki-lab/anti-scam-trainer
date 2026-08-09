import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  useLoginMutation,
  useRegisterMutation,
  type Credentials,
  type UserRole,
} from '@/entities/user'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

export function useAuthFlow(mode: 'login' | 'register', trainingRole: UserRole) {
  const navigate = useNavigate()
  const isPreview = useIsPreview()
  const [login, loginState] = useLoginMutation()
  const [registerAccount, registerState] = useRegisterMutation()
  const [error, setError] = useState('')

  const submit = async (credentials: Credentials) => {
    setError('')

    if (isPreview) {
      navigate('/preview/dashboard', { replace: true })
      return
    }

    try {
      if (mode === 'register') {
        await registerAccount({ ...credentials, trainingRole }).unwrap()
        navigate('/login?registered=1', { replace: true })
        return
      }
      await login(credentials).unwrap()
      navigate('/dashboard', { replace: true })
    } catch (reason) {
      setError(getApiErrorMessage(reason))
    }
  }

  return {
    submit,
    error,
    isSubmitting: loginState.isLoading || registerState.isLoading,
  }
}
