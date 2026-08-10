import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { RulesPage } from './RulesPage'

describe('RulesPage', () => {
  afterEach(() => cleanup())

  it('показывает этапы тренажёра и правила безопасной сделки', () => {
    render(<RulesPage onNavigate={vi.fn()} />)

    expect(screen.getByRole('heading', { name: 'Правила сервиса' })).toBeTruthy()
    expect(screen.getByRole('heading', { name: 'КАК РАБОТАЕТ СЕРВИС' })).toBeTruthy()
    expect(screen.getByText('1. МИССИИ')).toBeTruthy()
    expect(screen.getByText(/Не сообщайте коды, пароли/)).toBeTruthy()
  })

  it('открывает каталог миссий из нижней кнопки', async () => {
    const user = userEvent.setup()
    const onNavigate = vi.fn()
    render(<RulesPage onNavigate={onNavigate} />)

    await user.click(screen.getByRole('button', { name: /К СПИСКУ МИССИЙ/ }))

    expect(onNavigate).toHaveBeenCalledWith('MISSIONS')
  })
})
