import { execFileSync } from 'node:child_process'
import { expect, test } from '@playwright/test'

test.beforeEach(({ page }, testInfo) => {
  void page
  test.skip(testInfo.project.name === 'tablet', 'Tablet uses responsive smoke coverage')
  execFileSync('make', ['demo-reset'], { cwd: '..', stdio: 'pipe' })
})

test('returning seller completes the saved smartphone attempt through HTTP API', async ({
  page,
}, testInfo) => {
  await page.goto('/login')
  await page.getByRole('textbox', { name: 'Логин' }).fill('demo-seller')
  await page.getByLabel('Пароль').fill('demo1234')
  await page.getByRole('button', { name: 'Войти' }).click()

  await expect(page.getByRole('heading', { name: 'Продолжить Прохождение' })).toBeVisible()
  await page.getByRole('link', { name: 'Продолжить Прохождение' }).click()
  const productTitle = page.getByText('iPhone 15, 128 ГБ', { exact: true })
  await expect(
    testInfo.project.name === 'desktop' ? productTitle.last() : productTitle.first(),
  ).toBeVisible()
  await expect(page.getByText('Шаг 1 из 2')).toBeVisible()

  await page.locator('button[aria-pressed]').nth(1).click()
  await page.getByRole('button', { name: 'Подтвердить ответ' }).click()
  await expect(page.getByText('Шаг 2 из 2')).toBeVisible()
  await page.locator('button[aria-pressed]').first().click()
  await page.getByRole('button', { name: 'Подтвердить ответ' }).click()

  await expect(page).toHaveURL(/\/sessions\/\d+\/result$/)
  await expect(page.getByText('75/100')).toBeVisible()
  await expect(page.getByLabel('2 из 3 звёзд')).toBeVisible()
  await expect(page.getByText('Новые достижения')).toBeVisible()
  await expect(page.getByText(/Серия 3 дня/)).toBeVisible()

  await page.getByRole('link', { name: 'Мой прогресс' }).click()
  await expect(page.getByRole('heading', { name: 'Прогресс' })).toBeVisible()
  await expect(page.getByRole('link', { name: /Поддельная оплата/ }).first()).toBeVisible()
})
