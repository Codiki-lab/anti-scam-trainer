import { useEffect, useState } from 'react'
import {
  useCreateAdminTheoryBlockMutation,
  useDeleteAdminTheoryBlockMutation,
  useUpdateAdminTheoryBlockMutation,
  type AdminTheoryBlock,
  type AdminTopic,
  type TheoryKind,
} from '@/entities/admin-content'
import { getApiErrorMessage } from '@/shared/http-error'
import {
  DangerAction,
  EditorActions,
  PrimaryAction,
  SelectField,
  TextAreaField,
  TextField,
} from './AdminFields'
import styles from './AdminPage.module.scss'

const theoryKinds: Array<[TheoryKind, string]> = [
  ['intro', 'Введение'],
  ['risk', 'Риск'],
  ['example', 'Пример'],
  ['safe_action', 'Безопасное действие'],
  ['summary', 'Итог'],
]

const createEmptyBlock = (sortOrder: number): AdminTheoryBlock => ({
  sortOrder,
  kind: theoryKinds[Math.min(sortOrder - 1, theoryKinds.length - 1)][0],
  title: '',
  body: '',
})

export function TheoryEditor({ topic, editable }: { topic: AdminTopic; editable: boolean }) {
  const [adding, setAdding] = useState(false)
  const topicId = topic.id
  if (!topicId) return null

  return (
    <section className={styles.contentSection}>
      <div className={styles.contentSectionHeader}>
        <div>
          <h3>Теория</h3>
          <p>{topic.theory.length} из 5 обязательных блоков</p>
        </div>
        {editable && topic.theory.length < 5 && !adding && (
          <button type="button" onClick={() => setAdding(true)}>
            + Добавить блок
          </button>
        )}
      </div>
      <div className={styles.nestedList}>
        {topic.theory.map((block) => (
          <TheoryBlockCard key={block.id} block={block} editable={editable} topicId={topicId} />
        ))}
        {adding && (
          <TheoryBlockCard
            block={createEmptyBlock(topic.theory.length + 1)}
            editable
            topicId={topicId}
            onCancel={() => setAdding(false)}
            onCreated={() => setAdding(false)}
          />
        )}
      </div>
    </section>
  )
}

interface TheoryBlockCardProps {
  topicId: number
  block: AdminTheoryBlock
  editable: boolean
  onCancel?: () => void
  onCreated?: () => void
}

function TheoryBlockCard({ topicId, block, editable, onCancel, onCreated }: TheoryBlockCardProps) {
  const [draft, setDraft] = useState(block)
  const [message, setMessage] = useState<string>()
  const [createBlock, createState] = useCreateAdminTheoryBlockMutation()
  const [updateBlock, updateState] = useUpdateAdminTheoryBlockMutation()
  const [deleteBlock, deleteState] = useDeleteAdminTheoryBlockMutation()

  useEffect(() => setDraft(block), [block])

  const save = async () => {
    setMessage(undefined)
    try {
      if (block.id) {
        await updateBlock({ topicId, itemId: block.id, value: draft }).unwrap()
        setMessage('Блок сохранён.')
      } else {
        await createBlock({ topicId, value: draft }).unwrap()
        onCreated?.()
      }
    } catch (requestError) {
      setMessage(getApiErrorMessage(requestError))
    }
  }

  const remove = async () => {
    if (!block.id || !window.confirm('Удалить блок Теории?')) return
    try {
      await deleteBlock({ topicId, itemId: block.id }).unwrap()
    } catch (requestError) {
      setMessage(getApiErrorMessage(requestError))
    }
  }

  const pending = createState.isLoading || updateState.isLoading || deleteState.isLoading

  return (
    <details className={styles.nestedCard} open={!block.id}>
      <summary>
        <span>{String(draft.sortOrder).padStart(2, '0')}</span>
        <b>{draft.title || 'Новый блок Теории'}</b>
        <em>{theoryKinds.find(([kind]) => kind === draft.kind)?.[1]}</em>
      </summary>
      <div className={styles.nestedBody}>
        <div className={styles.formGrid}>
          <TextField
            disabled={!editable}
            label="Порядок"
            max={5}
            min={1}
            type="number"
            value={draft.sortOrder}
            onChange={(event) => setDraft({ ...draft, sortOrder: Number(event.target.value) })}
          />
          <SelectField
            disabled={!editable}
            label="Тип блока"
            value={draft.kind}
            onChange={(event) => setDraft({ ...draft, kind: event.target.value as TheoryKind })}
          >
            {theoryKinds.map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </SelectField>
          <TextField
            className={styles.wideField}
            disabled={!editable}
            label="Заголовок"
            value={draft.title}
            onChange={(event) => setDraft({ ...draft, title: event.target.value })}
          />
          <TextAreaField
            className={styles.wideField}
            disabled={!editable}
            label="Текст"
            value={draft.body}
            onChange={(event) => setDraft({ ...draft, body: event.target.value })}
          />
        </div>
        {message && <p className={styles.formMessage}>{message}</p>}
        {editable && (
          <EditorActions>
            <PrimaryAction disabled={pending} onClick={save}>
              Сохранить блок
            </PrimaryAction>
            {onCancel && <button onClick={onCancel}>Отмена</button>}
            {block.id && (
              <DangerAction disabled={pending} onClick={remove}>
                Удалить
              </DangerAction>
            )}
          </EditorActions>
        )}
      </div>
    </details>
  )
}
