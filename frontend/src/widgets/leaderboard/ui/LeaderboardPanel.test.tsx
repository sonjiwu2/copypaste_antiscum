import { render, screen } from '@testing-library/react'
import { vi } from 'vitest'
import { LeaderboardPanel } from './LeaderboardPanel'

const useLeaderboardQuery = vi.fn()

vi.mock('../../../shared/api', () => ({
  useLeaderboardQuery: () => useLeaderboardQuery(),
}))

describe('LeaderboardPanel', () => {
  beforeEach(() => {
    useLeaderboardQuery.mockReset()
  })

  it('shows the real global leaderboard', () => {
    useLeaderboardQuery.mockReturnValue({
      data: {
        leaders: [
          {
            rank: 1,
            displayName: 'Дмитрий',
            avatar: 'leader-2',
            rating: 525,
            completedScenarios: 7,
            averageScore: 75,
            currentPlayer: true,
          },
        ],
        current: {
          rank: 1,
          displayName: 'Дмитрий',
          avatar: 'leader-2',
          rating: 525,
          completedScenarios: 7,
          averageScore: 75,
          currentPlayer: true,
        },
      },
      isLoading: false,
      isError: false,
    })

    render(<LeaderboardPanel />)

    expect(screen.getByText('Дмитрий')).toBeTruthy()
    expect(screen.getByText('525 RP')).toBeTruthy()
    expect(screen.getByText('7 мисс. · 75%')).toBeTruthy()
  })

  it('invites the first player when nobody has completed a mission yet', () => {
    useLeaderboardQuery.mockReturnValue({
      data: { leaders: [] },
      isLoading: false,
      isError: false,
    })

    render(<LeaderboardPanel />)

    expect(screen.getByText('Станьте первым лидером')).toBeTruthy()
  })
})
