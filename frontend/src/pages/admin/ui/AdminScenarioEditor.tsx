import { useEffect, useState } from 'react'
import {
  useChangeAdminScenarioStatusMutation,
  useCreateAdminScenarioMutation,
  useGetAdminTopicsQuery,
  useUpdateAdminScenarioMutation,
  type AdminScenario,
  type AdminScenarioDraft,
} from '@/entities/admin-content'
import type { UserRole } from '@/entities/user'
import { getApiErrorMessage } from '@/shared/http-error'
import { ContentStatusActions } from './ContentStatusActions'
import { EditorActions, PrimaryAction, SelectField, TextAreaField, TextField } from './AdminFields'
import styles from './AdminPage.module.scss'

const emptyScenario: AdminScenarioDraft = {
  title: '',
  description: '',
  levelId: 1,
  topicId: 1,
  role: 'buyer',
  scamScheme: '',
  productContext: {},
  aiSystemPrompt: '',
  finalRubric: {},
}

interface AdminScenarioEditorProps {
  scenario?: AdminScenario
  onCreated: (id: number) => void
}

export function AdminScenarioEditor({ scenario, onCreated }: AdminScenarioEditorProps) {
  const scenarioId = scenario?.id
  const { data: topics = [] } = useGetAdminTopicsQuery()
  const [draft, setDraft] = useState<AdminScenarioDraft>(emptyScenario)
  const [productContext, setProductContext] = useState('{}')
  const [finalRubric, setFinalRubric] = useState('{}')
  const [message, setMessage] = useState<string>()
  const [createScenario, createState] = useCreateAdminScenarioMutation()
  const [updateScenario, updateState] = useUpdateAdminScenarioMutation()
  const [changeStatus, statusState] = useChangeAdminScenarioStatusMutation()

  useEffect(() => {
    if (!scenario) return
    setDraft({
      title: scenario.title,
      description: scenario.description,
      levelId: scenario.levelId,
      topicId: scenario.topicId,
      role: scenario.role,
      scamScheme: scenario.scamScheme,
      productContext: scenario.productContext,
      aiSystemPrompt: scenario.aiSystemPrompt,
      finalRubric: scenario.finalRubric,
    })
    setProductContext(JSON.stringify(scenario.productContext, null, 2))
    setFinalRubric(JSON.stringify(scenario.finalRubric, null, 2))
  }, [scenario])

  useEffect(() => {
    const firstTopic = topics[0]
    const firstTopicId = firstTopic?.id
    if (!scenarioId && firstTopic && firstTopicId) {
      setDraft((current) => ({ ...current, topicId: firstTopicId, role: firstTopic.role }))
    }
  }, [scenarioId, topics])

  const withJson = (): AdminScenarioDraft => {
    const parsedContext = JSON.parse(productContext) as unknown
    const parsedRubric = JSON.parse(finalRubric) as unknown
    if (!isJsonObject(parsedContext) || !isJsonObject(parsedRubric)) {
      throw new Error('Контекст товара и итоговая рубрика должны быть JSON-объектами.')
    }
    return { ...draft, productContext: parsedContext, finalRubric: parsedRubric }
  }

  const save = async () => {
    setMessage(undefined)
    try {
      const value = withJson()
      if (scenarioId) {
        await updateScenario({ scenarioId, value }).unwrap()
        setMessage('Основные данные Сценария сохранены.')
      } else {
        const created = await createScenario(value).unwrap()
        if (!created.id) throw new Error('Backend не вернул id Сценария.')
        onCreated(created.id)
      }
    } catch (requestError) {
      setMessage(
        requestError instanceof SyntaxError
          ? 'Проверьте синтаксис JSON в контексте товара и итоговой рубрике.'
          : getApiErrorMessage(requestError),
      )
    }
  }

  const applyStatus = async (action: 'publish' | 'deactivate' | 'restore' | 'archive') => {
    if (!scenarioId) return
    setMessage(undefined)
    try {
      await changeStatus({ scenarioId, action }).unwrap()
      setMessage('Статус Сценария обновлён.')
    } catch (requestError) {
      setMessage(getApiErrorMessage(requestError))
    }
  }

  const pending = createState.isLoading || updateState.isLoading || statusState.isLoading
  const editable = !scenario?.status || scenario.status === 'draft'
  const matchingTopics = topics.filter(
    (topic) => topic.role === draft.role && topic.status !== 'archived',
  )

  return (
    <article className={styles.editor}>
      <div className={styles.editorTitle}>
        <div>
          <span className={styles.eyebrow}>
            {scenarioId ? `Сценарий #${scenarioId}` : 'Новая тренировка'}
          </span>
          <h2>{scenarioId ? scenario?.title : 'Создать Сценарий'}</h2>
        </div>
        {scenario?.status && <span data-status={scenario.status}>{scenario.status}</span>}
      </div>

      <div className={styles.formGrid}>
        <TextField
          className={styles.wideField}
          disabled={!editable}
          label="Название"
          value={draft.title}
          onChange={(event) => setDraft({ ...draft, title: event.target.value })}
        />
        <SelectField
          disabled={!editable}
          label="Ролевая ветка"
          value={draft.role}
          onChange={(event) => {
            const role = event.target.value as UserRole
            const firstTopic = topics.find(
              (topic) => topic.role === role && topic.status !== 'archived',
            )
            setDraft({ ...draft, role, topicId: firstTopic?.id ?? draft.topicId })
          }}
        >
          <option value="buyer">Покупатель</option>
          <option value="seller">Продавец</option>
        </SelectField>
        <SelectField
          disabled={!editable}
          label="Тема"
          value={draft.topicId}
          onChange={(event) => setDraft({ ...draft, topicId: Number(event.target.value) })}
        >
          {matchingTopics.map((topic) => (
            <option key={topic.id} value={topic.id}>
              {topic.title}
            </option>
          ))}
        </SelectField>
        <SelectField
          disabled={!editable}
          label="Уровень"
          hint="Уровни 1–4 используют соответствующие backend id"
          value={draft.levelId}
          onChange={(event) => setDraft({ ...draft, levelId: Number(event.target.value) })}
        >
          {[1, 2, 3, 4].map((level) => (
            <option key={level} value={level}>
              Уровень {level}
            </option>
          ))}
        </SelectField>
        <TextAreaField
          className={styles.wideField}
          disabled={!editable}
          label="Описание"
          value={draft.description}
          onChange={(event) => setDraft({ ...draft, description: event.target.value })}
        />
        <TextField
          className={styles.wideField}
          disabled={!editable}
          label="Схема мошенничества"
          value={draft.scamScheme}
          onChange={(event) => setDraft({ ...draft, scamScheme: event.target.value })}
        />
        <TextAreaField
          className={styles.wideField}
          disabled={!editable}
          label="Контекст товара (JSON)"
          value={productContext}
          onChange={(event) => setProductContext(event.target.value)}
        />
        <TextAreaField
          className={styles.wideField}
          disabled={!editable}
          label="Системная инструкция AI"
          value={draft.aiSystemPrompt}
          onChange={(event) => setDraft({ ...draft, aiSystemPrompt: event.target.value })}
        />
        <TextAreaField
          className={styles.wideField}
          disabled={!editable}
          label="Итоговая рубрика (JSON)"
          value={finalRubric}
          onChange={(event) => setFinalRubric(event.target.value)}
        />
      </div>

      {message && <p className={styles.formMessage}>{message}</p>}
      {editable && (
        <EditorActions>
          <PrimaryAction disabled={pending} onClick={save}>
            {scenarioId ? 'Сохранить Сценарий' : 'Создать черновик'}
          </PrimaryAction>
        </EditorActions>
      )}
      <ContentStatusActions disabled={pending} status={scenario?.status} onAction={applyStatus} />

      {scenarioId && (
        <aside className={styles.contractNotice}>
          <b>Редактор Шагов ограничен текущим API</b>
          <p>
            Backend принимает команды создания и изменения Шагов, но не возвращает их состав
            админскому клиенту. Поэтому интерфейс не показывает неподтверждённые данные и не хранит
            их только в памяти браузера.
          </p>
        </aside>
      )}
    </article>
  )
}

function isJsonObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
