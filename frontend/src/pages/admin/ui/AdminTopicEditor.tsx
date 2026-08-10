import { useEffect, useState } from 'react'
import {
  useChangeAdminTopicStatusMutation,
  useCreateAdminTopicMutation,
  useGetAdminTopicQuery,
  useUpdateAdminTopicMutation,
  type AdminTopicDraft,
} from '@/entities/admin-content'
import type { UserRole } from '@/entities/user'
import { getApiErrorMessage } from '@/shared/http-error'
import { ContentStatusActions } from './ContentStatusActions'
import { EditorActions, PrimaryAction, SelectField, TextAreaField, TextField } from './AdminFields'
import { QuizEditor } from './QuizEditor'
import { TheoryEditor } from './TheoryEditor'
import styles from './AdminPage.module.scss'

const emptyTopic: AdminTopicDraft = {
  slug: '',
  role: 'buyer',
  title: '',
  description: '',
  sortOrder: 1,
}

interface AdminTopicEditorProps {
  topicId?: number
  onCreated: (id: number) => void
}

export function AdminTopicEditor({ topicId, onCreated }: AdminTopicEditorProps) {
  const { data: topic, isLoading, error } = useGetAdminTopicQuery(topicId ?? 0, { skip: !topicId })
  const [draft, setDraft] = useState<AdminTopicDraft>(emptyTopic)
  const [message, setMessage] = useState<string>()
  const [createTopic, createState] = useCreateAdminTopicMutation()
  const [updateTopic, updateState] = useUpdateAdminTopicMutation()
  const [changeStatus, statusState] = useChangeAdminTopicStatusMutation()

  useEffect(() => {
    if (!topic) return
    setDraft({
      slug: topic.slug,
      role: topic.role,
      title: topic.title,
      description: topic.description,
      sortOrder: topic.sortOrder,
    })
  }, [topic])

  const save = async () => {
    setMessage(undefined)
    try {
      if (topicId) {
        await updateTopic({ topicId, value: draft }).unwrap()
        setMessage('Основные данные Темы сохранены.')
      } else {
        const created = await createTopic(draft).unwrap()
        if (!created.id) throw new Error('Backend не вернул id Темы.')
        onCreated(created.id)
      }
    } catch (requestError) {
      setMessage(getApiErrorMessage(requestError))
    }
  }

  const applyStatus = async (action: 'publish' | 'deactivate' | 'restore' | 'archive') => {
    if (!topicId) return
    setMessage(undefined)
    try {
      await changeStatus({ topicId, action }).unwrap()
      setMessage('Статус Темы обновлён.')
    } catch (requestError) {
      setMessage(getApiErrorMessage(requestError))
    }
  }

  if (isLoading) return <div className={styles.editor}>Загружаем Тему…</div>
  if (error) return <div className={styles.editorError}>{getApiErrorMessage(error)}</div>

  const pending = createState.isLoading || updateState.isLoading || statusState.isLoading
  const editable = !topic?.status || topic.status === 'draft'

  return (
    <article className={styles.editor}>
      <div className={styles.editorTitle}>
        <div>
          <span className={styles.eyebrow}>{topicId ? `Тема #${topicId}` : 'Новый материал'}</span>
          <h2>{topicId ? topic?.title : 'Создать Тему'}</h2>
        </div>
        {topic?.status && <span data-status={topic.status}>{topic.status}</span>}
      </div>

      <div className={styles.formGrid}>
        <TextField
          disabled={!editable}
          label="Название"
          value={draft.title}
          onChange={(event) => setDraft({ ...draft, title: event.target.value })}
        />
        <TextField
          disabled={!editable}
          label="Slug"
          hint="Латиницей, например phishing-links"
          value={draft.slug}
          onChange={(event) => setDraft({ ...draft, slug: event.target.value })}
        />
        <SelectField
          disabled={!editable}
          label="Ролевая ветка"
          value={draft.role}
          onChange={(event) => setDraft({ ...draft, role: event.target.value as UserRole })}
        >
          <option value="buyer">Покупатель</option>
          <option value="seller">Продавец</option>
        </SelectField>
        <TextField
          disabled={!editable}
          label="Порядок"
          max={6}
          min={1}
          type="number"
          value={draft.sortOrder}
          onChange={(event) => setDraft({ ...draft, sortOrder: Number(event.target.value) })}
        />
        <TextAreaField
          className={styles.wideField}
          disabled={!editable}
          label="Описание"
          value={draft.description}
          onChange={(event) => setDraft({ ...draft, description: event.target.value })}
        />
      </div>

      {message && <p className={styles.formMessage}>{message}</p>}
      {editable && (
        <EditorActions>
          <PrimaryAction disabled={pending} onClick={save}>
            {topicId ? 'Сохранить Тему' : 'Создать черновик'}
          </PrimaryAction>
        </EditorActions>
      )}
      <ContentStatusActions disabled={pending} status={topic?.status} onAction={applyStatus} />

      {topicId && topic && (
        <>
          <TheoryEditor editable={editable} topic={topic} />
          <QuizEditor editable={editable} topic={topic} />
        </>
      )}
    </article>
  )
}
