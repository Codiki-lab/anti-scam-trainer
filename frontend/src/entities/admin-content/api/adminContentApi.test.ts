// @vitest-environment node

import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import { createTestStore } from '@/test/createTestStore'
import { server } from '@/test/server'
import { adminContentApi } from './adminContentApi'

const topicResponse = {
  id: 4,
  slug: 'safe-payment',
  role: 'seller',
  title: 'Безопасная оплата',
  description: 'Описание',
  sort_order: 2,
  status: 'draft',
  theory: [],
  quiz: [],
}

describe('admin content API boundary', () => {
  it('loads the Topic aggregate through the admin endpoint', async () => {
    server.use(
      http.get('http://localhost/api/v1/admin/topics/4', () => HttpResponse.json(topicResponse)),
    )

    const store = createTestStore()
    const topic = await store.dispatch(adminContentApi.endpoints.getAdminTopic.initiate(4)).unwrap()

    expect(topic).toMatchObject({ id: 4, sortOrder: 2, role: 'seller', status: 'draft' })
  })

  it('posts only the writable Topic contract', async () => {
    server.use(
      http.post('http://localhost/api/v1/admin/topics', async ({ request }) => {
        expect(await request.json()).toEqual({
          slug: 'safe-payment',
          role: 'seller',
          title: 'Безопасная оплата',
          description: 'Описание',
          sort_order: 2,
        })
        return HttpResponse.json(topicResponse, { status: 201 })
      }),
    )

    const store = createTestStore()
    const topic = await store
      .dispatch(
        adminContentApi.endpoints.createAdminTopic.initiate({
          slug: 'safe-payment',
          role: 'seller',
          title: 'Безопасная оплата',
          description: 'Описание',
          sortOrder: 2,
        }),
      )
      .unwrap()

    expect(topic.id).toBe(4)
  })
})
