import { useEffect, useState } from 'react'
import { statusLabels, useGetAdminTopicsQuery, type AdminTopic } from '@/entities/admin-content'
import { ErrorState } from '@/shared/error-state'
import { AdminTopicEditor } from './AdminTopicEditor'
import styles from './AdminPage.module.scss'

export function AdminTopics() {
  const { data = [], isLoading, error, refetch } = useGetAdminTopicsQuery()
  const [selectedId, setSelectedId] = useState<number>()
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    if (!creating && !selectedId && data[0]?.id) setSelectedId(data[0].id)
  }, [creating, data, selectedId])

  if (isLoading) return <div className={styles.loading}>Загружаем Темы…</div>
  if (error) return <ErrorState message="Не удалось загрузить Темы." onRetry={refetch} />

  const openTopic = (topic: AdminTopic) => {
    setCreating(false)
    setSelectedId(topic.id)
  }

  return (
    <div className={styles.workspace}>
      <aside className={styles.catalog}>
        <div className={styles.catalogHeader}>
          <div>
            <b>Темы</b>
            <span>{data.length} материалов</span>
          </div>
          <button
            type="button"
            onClick={() => {
              setSelectedId(undefined)
              setCreating(true)
            }}
          >
            + Новая
          </button>
        </div>
        <div className={styles.catalogList}>
          {data.map((topic) => (
            <button
              key={topic.id}
              className={selectedId === topic.id && !creating ? styles.selectedItem : undefined}
              type="button"
              onClick={() => openTopic(topic)}
            >
              <span>
                <b>{topic.title}</b>
                <small>{topic.role === 'buyer' ? 'Покупатель' : 'Продавец'}</small>
              </span>
              {topic.status && <em data-status={topic.status}>{statusLabels[topic.status]}</em>}
            </button>
          ))}
          {data.length === 0 && <p>Тем пока нет. Создайте первый черновик.</p>}
        </div>
      </aside>

      <AdminTopicEditor
        key={creating ? 'new' : selectedId}
        topicId={creating ? undefined : selectedId}
        onCreated={(id) => {
          setCreating(false)
          setSelectedId(id)
        }}
      />
    </div>
  )
}
