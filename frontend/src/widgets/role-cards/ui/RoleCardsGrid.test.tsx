import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import type { Progress, Scenario } from '../../../shared/api'
import * as client from '../../../shared/api/client'
import { RoleCardsGrid } from './RoleCardsGrid'

vi.mock('../../../shared/api/client', async () => {
  const actual = await vi.importActual<typeof import('../../../shared/api/client')>(
    '../../../shared/api/client'
  )

  return { ...actual, fetchProgress: vi.fn(), fetchScenarios: vi.fn() }
})

const scenarios: Scenario[] = [
  {
    id: 'buyer-one',
    version: 1,
    slug: 'buyer-one',
    role: 'buyer',
    title: 'Покупатель 1',
    description: '',
    difficulty: 'easy',
    estimatedMinutes: 3,
    maxDecisions: 2,
  },
  {
    id: 'buyer-two',
    version: 1,
    slug: 'buyer-two',
    role: 'buyer',
    title: 'Покупатель 2',
    description: '',
    difficulty: 'medium',
    estimatedMinutes: 4,
    maxDecisions: 3,
  },
  {
    id: 'seller-one',
    version: 1,
    slug: 'seller-one',
    role: 'seller',
    title: 'Продавец 1',
    description: '',
    difficulty: 'hard',
    estimatedMinutes: 5,
    maxDecisions: 4,
  },
]

const progress: Progress = {
  summary: {
    completedAttempts: 4,
    completedScenarios: 3,
    averageScore: 83,
    bestScore: 100,
    latestScore: 70,
  },
  activeAttempts: [],
  scenarioProgress: [
    {
      scenarioId: 'buyer-one',
      role: 'buyer',
      attempts: 1,
      lastScore: 80,
      bestScore: 80,
    },
    {
      scenarioId: 'buyer-two',
      role: 'buyer',
      attempts: 2,
      lastScore: 90,
      bestScore: 100,
    },
    {
      scenarioId: 'seller-one',
      role: 'seller',
      attempts: 1,
      lastScore: 70,
      bestScore: 70,
    },
  ],
  skills: [],
  weakRiskTags: [],
  recentAttempts: [],
  recommendedScenarios: [],
}

function renderGrid() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <RoleCardsGrid onNavigate={vi.fn()} />
    </QueryClientProvider>
  )
}

describe('RoleCardsGrid', () => {
  beforeEach(() => {
    vi.mocked(client.fetchScenarios).mockResolvedValue(scenarios)
    vi.mocked(client.fetchProgress).mockResolvedValue(progress)
  })

  it('shows mission totals and role progress reported by the backend', async () => {
    renderGrid()

    const buyerStats = screen.getByLabelText('Статистика: Покупатель')
    const sellerStats = screen.getByLabelText('Статистика: Продавец')

    await waitFor(() => expect(within(buyerStats).getAllByText('2')).toHaveLength(2))
    expect(within(buyerStats).getByText('90%')).toBeTruthy()

    expect(within(sellerStats).getAllByText('1')).toHaveLength(2)
    expect(within(sellerStats).getByText('70%')).toBeTruthy()
  })
})
