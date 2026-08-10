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

test('free play stays locked until all topics are completed', async ({ page }) => {
  await page.goto('/preview/dashboard')

  await expect(page.getByRole('button', { name: 'Начать игру →' })).toBeDisabled()
  await expect(page.getByText(/Закрыта: завершите все 6 Темы.*\(0\/6\)/)).toBeVisible()
})

test('navigation and focus remain accessible at every breakpoint', async ({ page }) => {
  await page.goto('/preview/dashboard')

  await expect(page.getByRole('link', { name: 'Главная', exact: true })).toHaveAttribute(
    'aria-current',
    'page',
  )
  await page.keyboard.press('Tab')
  const focus = await page.evaluate(() => {
    const element = document.activeElement as HTMLElement | null
    const style = element ? getComputedStyle(element) : null
    return {
      tag: element?.tagName,
      outlineStyle: style?.outlineStyle,
      outlineWidth: style?.outlineWidth,
    }
  })
  expect(focus.tag).not.toBe('BODY')
  expect(focus.outlineStyle).not.toBe('none')
  expect(focus.outlineWidth).not.toBe('0px')

  const viewport = page.viewportSize()
  if (viewport && viewport.width <= 560) {
    const targets = await page.locator('nav a').evaluateAll((links) =>
      links.map((link) => {
        const rect = link.getBoundingClientRect()
        return { width: rect.width, height: rect.height }
      }),
    )
    expect(targets.some(({ width, height }) => width >= 44 && height >= 44)).toBe(true)
  }
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
