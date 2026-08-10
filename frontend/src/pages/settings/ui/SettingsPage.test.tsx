import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { DEFAULT_SETTINGS, useSettingsStore } from '../../../entities/settings'
import { SettingsPage } from './SettingsPage'

describe('SettingsPage', () => {
  beforeEach(() => {
    localStorage.clear()
    useSettingsStore.setState({ settings: DEFAULT_SETTINGS })
  })

  afterEach(() => {
    cleanup()
  })

  it('показывает все разделы настроек', () => {
    render(<SettingsPage />)

    for (const title of ['ОБЩИЕ', 'ЗВУК', 'УВЕДОМЛЕНИЯ', 'ИНТЕРФЕЙС', 'АККАУНТ', 'ДАННЫЕ']) {
      expect(screen.getByRole('heading', { name: title })).toBeTruthy()
    }
  })

  it('показывает будущие настройки интерфейса как находящиеся в разработке', () => {
    render(<SettingsPage />)

    for (const label of ['Анимации интерфейса', 'Масштаб интерфейса', 'Высокая контрастность']) {
      expect(screen.getByText(label)).toBeTruthy()
    }
    expect(screen.getAllByText('В разработке')).toHaveLength(3)
  })

  it('применяет изменения только после сохранения', async () => {
    const user = userEvent.setup({ pointerEventsCheck: 0 })
    const handleToast = vi.fn()
    render(<SettingsPage onToast={handleToast} />)

    const themeToggle = screen.getByRole('checkbox', { name: /Тёмная тема/ })
    await user.click(themeToggle)

    // До нажатия «Сохранить» store остаётся прежним.
    expect(useSettingsStore.getState().settings.theme).toBe('light')

    await user.click(screen.getByRole('button', { name: /Сохранить/ }))

    expect(useSettingsStore.getState().settings.theme).toBe('dark')
    expect(handleToast).toHaveBeenCalledWith(expect.stringContaining('сохранены'))
  })

  it('отмена возвращает черновик к сохранённым значениям', async () => {
    const user = userEvent.setup({ pointerEventsCheck: 0 })
    render(<SettingsPage onToast={vi.fn()} />)

    const darkTheme = screen.getByRole('checkbox', { name: /Тёмная тема/ })
    await user.click(darkTheme)
    expect((darkTheme as HTMLInputElement).checked).toBe(true)

    await user.click(screen.getByRole('button', { name: /Отмена/ }))

    expect((darkTheme as HTMLInputElement).checked).toBe(false)
    expect(useSettingsStore.getState().settings.theme).toBe('light')
  })

  it('сброс по умолчанию сохраняет имя и аватар игрока', async () => {
    const user = userEvent.setup({ pointerEventsCheck: 0 })
    useSettingsStore.setState({
      settings: {
        ...DEFAULT_SETTINGS,
        theme: 'dark',
        playerName: 'Игрок',
        avatar: 'leader-2',
      },
    })
    render(<SettingsPage onToast={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: /По умолчанию/ }))
    await user.click(screen.getByRole('button', { name: /Сохранить/ }))

    const settings = useSettingsStore.getState().settings
    expect(settings.theme).toBe('light')
    expect(settings.playerName).toBe('Игрок')
    expect(settings.avatar).toBe('leader-2')
  })

  it('очистка прогресса требует подтверждения', async () => {
    const user = userEvent.setup({ pointerEventsCheck: 0 })
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
    const handleToast = vi.fn()
    render(<SettingsPage onToast={handleToast} />)

    await user.click(screen.getByRole('button', { name: /Очистить прогресс/ }))

    expect(confirmSpy).toHaveBeenCalled()
    expect(handleToast).not.toHaveBeenCalled()
    confirmSpy.mockRestore()
  })
})
