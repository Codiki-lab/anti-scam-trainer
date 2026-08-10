import { useEffect, useState } from 'react'
import {
  statusLabels,
  useGetAdminScenariosQuery,
  type AdminScenario,
} from '@/entities/admin-content'
import { ErrorState } from '@/shared/error-state'
import { AdminScenarioEditor } from './AdminScenarioEditor'
import styles from './AdminPage.module.scss'

export function AdminScenarios() {
  const { data = [], isLoading, error, refetch } = useGetAdminScenariosQuery()
  const [selectedId, setSelectedId] = useState<number>()
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    if (!creating && !selectedId && data[0]?.id) setSelectedId(data[0].id)
  }, [creating, data, selectedId])

  if (isLoading) return <div className={styles.loading}>Загружаем Сценарии…</div>
  if (error) return <ErrorState message="Не удалось загрузить Сценарии." onRetry={refetch} />

  const openScenario = (scenario: AdminScenario) => {
    setCreating(false)
    setSelectedId(scenario.id)
  }

  return (
    <div className={styles.workspace}>
      <aside className={styles.catalog}>
        <div className={styles.catalogHeader}>
          <div>
            <b>Сценарии</b>
            <span>{data.length} тренировок</span>
          </div>
          <button
            type="button"
            onClick={() => {
              setSelectedId(undefined)
              setCreating(true)
            }}
          >
            + Новый
          </button>
        </div>
        <div className={styles.catalogList}>
          {data.map((scenario) => (
            <button
              key={scenario.id}
              className={selectedId === scenario.id && !creating ? styles.selectedItem : undefined}
              type="button"
              onClick={() => openScenario(scenario)}
            >
              <span>
                <b>{scenario.title}</b>
                <small>
                  {scenario.role === 'buyer' ? 'Покупатель' : 'Продавец'} · уровень{' '}
                  {scenario.levelId}
                </small>
              </span>
              {scenario.status && (
                <em data-status={scenario.status}>{statusLabels[scenario.status]}</em>
              )}
            </button>
          ))}
          {data.length === 0 && <p>Сценариев пока нет.</p>}
        </div>
      </aside>

      <AdminScenarioEditor
        key={creating ? 'new' : selectedId}
        scenario={creating ? undefined : data.find((item) => item.id === selectedId)}
        onCreated={(id) => {
          setCreating(false)
          setSelectedId(id)
        }}
      />
    </div>
  )
}
