import { expect, test } from '@playwright/test'

test('preview user can open the main learning flow', async ({ page }) => {
  await page.goto('/preview/login')
  await page.getByRole('textbox', { name: 'Логин' }).fill('demo-user')
  await page.getByLabel('Пароль').fill('demo-password')
  await page.getByRole('button', { name: 'Войти' }).click()

  await expect(
    page.getByRole('heading', { name: 'Учитесь защищать себя и свои сделки' }),
  ).toBeVisible()

  await page.getByRole('link', { name: /все темы/i }).click()
  await expect(page.getByRole('heading', { name: /темы для/i })).toBeVisible()
  await expect(page.locator('main a')).toHaveCount(6)
})

test('free play is available without completed levels', async ({ page }) => {
  await page.goto('/preview/dashboard')

  await page.getByRole('button', { name: 'Начать игру →' }).click()

  await expect(page).toHaveURL(/\/preview\/sessions\/free-play$/)
  await expect(page.getByPlaceholder('Напишите безопасный ответ…')).toBeVisible()
})

test('each topic opens its own theory', async ({ page }) => {
  await page.goto('/preview/lessons')

  await page.getByRole('link', { name: /предоплата/i }).click()

  await expect(page).toHaveURL(/\/preview\/lessons\/2$/)
  await expect(page.getByRole('heading', { name: 'Предоплата' })).toBeVisible()
})

test('unavailable training levels are visibly locked', async ({ page }) => {
  await page.goto('/preview/chats')

  await expect(page.getByRole('button', { name: /закрыто/i })).toHaveCount(2)
  await expect(page.getByRole('button', { name: /закрыто/i }).first()).toBeDisabled()
})
