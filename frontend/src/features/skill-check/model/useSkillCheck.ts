import { useState } from 'react'
import {
  useAnswerSkillCheckMutation,
  useGetSkillCheckQuery,
  useStartSkillCheckMutation,
  type SkillCheck,
} from '@/entities/learning'
import { getApiErrorMessage } from '@/shared/http-error'

interface SkillCheckState {
  check?: SkillCheck
  error: string
  isLoading: boolean
  start: () => Promise<void>
  answer: (answer: boolean) => Promise<void>
  refresh: () => Promise<void>
}

export function useSkillCheck(topicId: number): SkillCheckState {
  const [startRequest, startState] = useStartSkillCheckMutation()
  const [answerRequest, answerState] = useAnswerSkillCheckMutation()
  const [checkId, setCheckId] = useState<number>()
  const checkQuery = useGetSkillCheckQuery(checkId ?? 0, { skip: !checkId })
  const [error, setError] = useState('')

  const run = async (request: () => Promise<SkillCheck>) => {
    setError('')
    try {
      await request()
    } catch (requestError) {
      setError(getApiErrorMessage(requestError))
    }
  }

  return {
    check: checkQuery.data,
    error,
    isLoading: startState.isLoading || answerState.isLoading || checkQuery.isFetching,
    start: () =>
      run(async () => {
        const started = await startRequest(topicId).unwrap()
        setCheckId(started.id)
        return started
      }),
    answer: (answer) => {
      const currentCheck = checkQuery.data
      if (!currentCheck) return Promise.resolve()
      return run(() => answerRequest({ checkId: currentCheck.id, answer }).unwrap())
    },
    refresh: () => {
      if (!checkId) return Promise.resolve()
      return run(() => checkQuery.refetch().unwrap())
    },
  }
}
