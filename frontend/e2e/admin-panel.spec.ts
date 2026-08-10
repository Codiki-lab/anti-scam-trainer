import { expect, test } from '@playwright/test'

const topic = {
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

test('administrator can open the content workspace', async ({ page }) => {
  await page.route('**/api/v1/auth/me', (route) =>
    route.fulfill({
      json: {
        id: 1,
        username: 'admin',
        access_role: 'admin',
        training_role: 'seller',
        streak: { current: 0, longest: 0, active_today: false },
      },
    }),
  )
  await page.route('**/api/v1/admin/topics', (route) => route.fulfill({ json: [topic] }))
  await page.route('**/api/v1/admin/topics/4', (route) => route.fulfill({ json: topic }))

  await page.goto('/admin')

  await expect(page.getByRole('heading', { name: 'Админ-панель' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Безопасная оплата' })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Админ-панель' })).toBeVisible()
})
