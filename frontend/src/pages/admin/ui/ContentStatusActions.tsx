import type { ContentStatus } from '@/entities/admin-content'
import { DangerAction, EditorActions, PrimaryAction, SecondaryAction } from './AdminFields'

interface ContentStatusActionsProps {
  status?: ContentStatus
  disabled?: boolean
  onAction: (action: 'publish' | 'deactivate' | 'restore' | 'archive') => void
}

export function ContentStatusActions({ status, disabled, onAction }: ContentStatusActionsProps) {
  if (!status) return null

  return (
    <EditorActions>
      {status === 'draft' && (
        <PrimaryAction disabled={disabled} onClick={() => onAction('publish')}>
          Опубликовать
        </PrimaryAction>
      )}
      {status === 'published' && (
        <SecondaryAction disabled={disabled} onClick={() => onAction('deactivate')}>
          Вернуть в черновик
        </SecondaryAction>
      )}
      {status === 'archived' ? (
        <SecondaryAction disabled={disabled} onClick={() => onAction('restore')}>
          Восстановить
        </SecondaryAction>
      ) : (
        <DangerAction disabled={disabled} onClick={() => onAction('archive')}>
          Архивировать
        </DangerAction>
      )}
    </EditorActions>
  )
}
