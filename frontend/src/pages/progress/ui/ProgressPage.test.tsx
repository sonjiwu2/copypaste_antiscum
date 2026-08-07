import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ProgressPage } from './ProgressPage'
import { useUserProgressStore } from '../../../entities/user-progress'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>)
}

describe('ProgressPage', () => {
  beforeEach(() => {
    useUserProgressStore.getState().resetProgress()
  })

  it('renders progress header and statistics dashboard', () => {
    renderWithProviders(<ProgressPage />)

    expect(screen.getByText('ТАБЛИЦА ПРОГРЕССА')).toBeTruthy()
    expect(screen.getByText('ИГРОВОЙ УРОВЕНЬ')).toBeTruthy()
    expect(screen.getByText('СТАТИСТИКА')).toBeTruthy()
    expect(screen.getByText('ДОСТИЖЕНИЯ И НАГРАДЫ')).toBeTruthy()
    expect(screen.getByText('ТАБЛИЦА ПРОЙДЕННЫХ МИССИЙ')).toBeTruthy()
  })

  it('shows empty state when no missions are completed', () => {
    renderWithProviders(<ProgressPage />)

    expect(screen.getByText('Вы пока не завершили ни одной тренировочной миссии.')).toBeTruthy()
  })

  it('triggers onNavigate when clicking "К миссиям"', async () => {
    const handleNavigate = vi.fn()
    const user = userEvent.setup()

    renderWithProviders(<ProgressPage onNavigate={handleNavigate} />)

    const missionBtn = screen.getByRole('button', { name: /к миссиям/i })
    await user.click(missionBtn)

    expect(handleNavigate).toHaveBeenCalledWith('MISSIONS')
  })

  it('displays user progress when user has completed missions', () => {
    useUserProgressStore.getState().recordMissionCompletion('test-mission-1', 100, 50)

    renderWithProviders(<ProgressPage />)

    expect(screen.getByText('100%')).toBeTruthy()
    expect(screen.getAllByText('Разблокировано').length).toBeGreaterThan(0)
  })
})
