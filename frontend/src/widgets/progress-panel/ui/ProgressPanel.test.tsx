import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import type { Progress } from '../../../shared/api'
import * as client from '../../../shared/api/client'
import { useUserProgressStore } from '../../../entities/user-progress'
import { ProgressPanel } from './ProgressPanel'

vi.mock('../../../shared/api/client', async () => {
  const actual = await vi.importActual<typeof import('../../../shared/api/client')>(
    '../../../shared/api/client'
  )

  return { ...actual, fetchProgress: vi.fn() }
})

const progress: Progress = {
  summary: {
    completedAttempts: 5,
    completedScenarios: 3,
    averageScore: 84,
    bestScore: 100,
    latestScore: 90,
  },
  activeAttempts: [],
  scenarioProgress: [],
  skills: [],
  weakRiskTags: [],
  recentAttempts: [],
  recommendedScenarios: [],
}

describe('ProgressPanel', () => {
  beforeEach(() => {
    useUserProgressStore.getState().resetProgress()
    vi.mocked(client.fetchProgress).mockResolvedValue(progress)
  })

  it('uses the backend completed-scenario count', async () => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })

    render(
      <QueryClientProvider client={queryClient}>
        <ProgressPanel />
      </QueryClientProvider>
    )

    const completed = screen.getByText('Пройдено').closest('.progress-stat')
    expect(completed).toBeTruthy()
    await waitFor(() => expect(within(completed as HTMLElement).getByText('3')).toBeTruthy())
  })
})
