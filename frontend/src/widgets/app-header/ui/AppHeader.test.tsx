import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Provider } from 'react-redux'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import type { Account } from '@/entities/user'
import { PreviewModeProvider } from '@/shared/runtime-mode'
import { createTestStore } from '@/test/createTestStore'
import { AppHeader } from './AppHeader'

const account: Account = {
  id: 1,
  username: 'demo',
  accessRole: 'user',
  trainingRole: 'seller',
  streak: { current: 3, longest: 7, isActiveToday: true },
}

function renderHeader() {
  return render(
    <Provider store={createTestStore()}>
      <PreviewModeProvider>
        <MemoryRouter>
          <AppHeader account={account} basePath="/preview" />
        </MemoryRouter>
      </PreviewModeProvider>
    </Provider>,
  )
}

describe('AppHeader profile menu', () => {
  it('closes on Escape and on a pointer press outside the menu', async () => {
    const user = userEvent.setup()
    renderHeader()

    const trigger = screen.getByRole('button', { name: 'Открыть меню профиля' })

    await user.click(trigger)
    expect(screen.getByRole('menu')).toBeInTheDocument()
    expect(trigger).toHaveAttribute('aria-expanded', 'true')

    fireEvent.keyDown(document, { key: 'Escape' })
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    expect(trigger).toHaveAttribute('aria-expanded', 'false')

    await user.click(trigger)
    expect(screen.getByRole('menu')).toBeInTheDocument()

    fireEvent.pointerDown(document.body)
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
  })
})
