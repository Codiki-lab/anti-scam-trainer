import { useState } from 'react'
import { useCurrentAccount, type UserRole } from '@/entities/user'
import { getApiErrorMessage } from '@/shared/http-error'

export function useChangeTrainingRole() {
  const { account, changeTrainingRole } = useCurrentAccount()
  const [isChanging, setIsChanging] = useState(false)
  const [error, setError] = useState('')

  const changeRole = async (role: UserRole) => {
    if (role === account.trainingRole || isChanging) return
    setError('')
    setIsChanging(true)

    try {
      await changeTrainingRole(role)
    } catch (reason) {
      setError(getApiErrorMessage(reason))
    } finally {
      setIsChanging(false)
    }
  }

  return { role: account.trainingRole, changeRole, isChanging, error }
}
