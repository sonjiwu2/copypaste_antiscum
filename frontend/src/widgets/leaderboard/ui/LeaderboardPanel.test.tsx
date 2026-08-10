import { render, screen } from '@testing-library/react'
import { LeaderboardPanel } from './LeaderboardPanel'

describe('LeaderboardPanel', () => {
  it('does not show fictional players before multiplayer is available', () => {
    render(<LeaderboardPanel />)

    expect(screen.getByText('Мультиплеер в разработке')).toBeTruthy()
    expect(screen.queryByText('Марина')).toBeNull()
    expect(screen.queryByText('Дима_Трейдер')).toBeNull()
    expect(screen.queryByText('МастерМагазина')).toBeNull()
  })
})
