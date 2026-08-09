import { useState } from 'react'
import { useGetTheoryQuery, useMarkTheoryReadMutation, type Theory } from '@/entities/learning'
import { getApiErrorMessage } from '@/shared/http-error'
import { useIsPreview } from '@/shared/runtime-mode'

export function useTheory(topicId: number, previewTheory?: Theory) {
  const isPreview = useIsPreview()
  const query = useGetTheoryQuery(topicId, { skip: isPreview || topicId < 1 })

  return {
    theory: isPreview ? previewTheory : query.data,
    isLoading: isPreview ? false : query.isLoading,
    error: isPreview || !query.error ? '' : getApiErrorMessage(query.error),
    retry: query.refetch,
  }
}

export function useFinishTheory(topicId: number) {
  const isPreview = useIsPreview()
  const [markRead, state] = useMarkTheoryReadMutation()
  const [error, setError] = useState('')

  const finishTheory = async () => {
    setError('')
    if (isPreview) return true
    try {
      await markRead(topicId).unwrap()
      return true
    } catch (reason) {
      setError(getApiErrorMessage(reason))
      return false
    }
  }

  return { finishTheory, isSaving: !isPreview && state.isLoading, error }
}
