import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Provider } from 'react-redux'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { AuthForm } from '@/features/auth'
import { PreviewModeProvider } from '@/shared/runtime-mode'
import { createTestStore } from './createTestStore'

describe('registration flow', () => {
  it('keeps the selected role in form state and follows the preview route', async () => {
    const user = userEvent.setup()

    render(
      <Provider store={createTestStore()}>
        <PreviewModeProvider>
          <MemoryRouter initialEntries={['/preview/register']}>
            <Routes>
              <Route path="/preview/register" element={<AuthForm mode="register" />} />
              <Route path="/preview/dashboard" element={<p>Главная открыта</p>} />
            </Routes>
          </MemoryRouter>
        </PreviewModeProvider>
      </Provider>,
    )

    await user.click(screen.getByRole('button', { name: /продавец/i }))
    await user.type(screen.getByRole('textbox', { name: /логин/i }), 'irina')
    await user.type(screen.getByLabelText(/пароль/i), 'secure-password')
    await user.click(screen.getByRole('button', { name: /создать аккаунт/i }))

    expect(await screen.findByText('Главная открыта')).toBeInTheDocument()
  })
})
