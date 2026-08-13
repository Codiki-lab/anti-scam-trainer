// @vitest-environment node

import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import { createTestStore } from '@/test/createTestStore'
import { server } from '@/test/server'
import { userApi } from './userApi'

describe('account API boundary', () => {
  it('validates and maps the backend DTO before exposing it to UI', async () => {
    server.use(
      http.get('http://localhost/api/v1/auth/me', () =>
        HttpResponse.json({
          id: 7,
          username: 'Ирина',
          access_role: 'user',
          training_role: 'seller',
          streak: { current: 3, longest: 5, active_today: true },
        }),
      ),
    )

    const store = createTestStore()
    const account = await store.dispatch(userApi.endpoints.getMe.initiate()).unwrap()

    expect(account).toMatchObject({
      id: 7,
      username: 'Ирина',
      accessRole: 'user',
      trainingRole: 'seller',
      streak: { current: 3, longest: 5, isActiveToday: true },
    })
  })
})
