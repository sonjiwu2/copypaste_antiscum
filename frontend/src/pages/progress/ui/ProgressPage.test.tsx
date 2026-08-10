import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ProgressPage } from './ProgressPage'
import { useUserProgressStore } from '../../../entities/user-progress'
import * as client from '../../../shared/api/client'
import type { Progress, Scenario } from '../../../shared/api'

vi.mock('../../../shared/api/client', async () => {
  const actual = await vi.importActual<typeof import('../../../shared/api/client')>(
    '../../../shared/api/client'
  )

  return { ...actual, fetchProgress: vi.fn(), fetchScenarios: vi.fn() }
})

const SCENARIO: Scenario = {
  id: 'buyer-fake-delivery',
  version: 1,
  slug: 'buyer-fake-delivery',
  role: 'buyer',
  title: 'Ссылка на доставку',
  description: 'Продавец предлагает оплату по ссылке.',
  difficulty: 'hard',
  estimatedMinutes: 4,
  maxDecisions: 3,
}

const EMPTY_PROGRESS: Progress = {
  summary: {
    completedAttempts: 0,
    completedScenarios: 0,
    averageScore: 0,
    bestScore: 0,
    latestScore: 0,
  },
  activeAttempts: [],
  scenarioProgress: [],
  skills: [],
  weakRiskTags: [],
  recentAttempts: [],
  recommendedScenarios: [],
}

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>)
}

describe('ProgressPage', () => {
  beforeEach(() => {
    useUserProgressStore.getState().resetProgress()
    vi.mocked(client.fetchProgress).mockResolvedValue(EMPTY_PROGRESS)
    vi.mocked(client.fetchScenarios).mockResolvedValue([SCENARIO])
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

  // Таблица прохождений полностью строится по данным сервера: награда
  // считается по сложности сценария и лучшему результату.
  it('displays completed scenarios reported by the backend', async () => {
    vi.mocked(client.fetchProgress).mockResolvedValue({
      ...EMPTY_PROGRESS,
      summary: { ...EMPTY_PROGRESS.summary, completedScenarios: 1 },
      scenarioProgress: [
        {
          scenarioId: 'buyer-fake-delivery',
          role: 'buyer',
          title: 'Ссылка на доставку',
          attempts: 2,
          lastScore: 80,
          bestScore: 100,
          lastCompletedAt: '2026-08-08T10:00:00Z',
        },
      ],
    })

    renderWithProviders(<ProgressPage />)

    await waitFor(() => expect(screen.getByText('Ссылка на доставку')).toBeTruthy())
    expect(screen.getByText('100%')).toBeTruthy()
    // hard = 20 XP при результате 100.
    expect(screen.getByText('+20 XP')).toBeTruthy()
    expect(screen.getByText('2')).toBeTruthy()
  })
})
