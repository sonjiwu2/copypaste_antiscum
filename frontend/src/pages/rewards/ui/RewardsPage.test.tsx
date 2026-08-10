import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useUserProgressStore } from '../../../entities/user-progress'
import { RewardsPage } from './RewardsPage'

describe('RewardsPage', () => {
  beforeEach(() => {
    localStorage.clear()
    useUserProgressStore.getState().resetProgress()
  })

  afterEach(() => cleanup())

  it('начисляет 25 XP один раз и показывает собранное состояние', async () => {
    const user = userEvent.setup()
    render(<RewardsPage onBack={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: /Забрать награду/i }))

    expect(useUserProgressStore.getState().totalXp).toBe(25)
    expect(screen.getByText(/Текущий баланс:/).textContent).toContain('25 XP')
    expect(
      (screen.getByRole('button', { name: /Награда собрана/i }) as HTMLButtonElement).disabled
    ).toBe(true)
  })
})
