import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import type { Scenario } from '../../../shared/api'
import * as client from '../../../shared/api/client'
import { FeaturedMissionsPanel } from './FeaturedMissionsPanel'

vi.mock('../../../shared/api/client', async () => {
  const actual = await vi.importActual<typeof import('../../../shared/api/client')>(
    '../../../shared/api/client'
  )

  return { ...actual, fetchScenarios: vi.fn() }
})

const scenarios: Scenario[] = [
  ['buyer-airpods-counterfeit', 'Поддельные AirPods', 'easy'],
  ['buyer-fake-delivery', 'Ссылка на доставку', 'medium'],
  ['buyer-gpu-hidden-repair', 'Видеокарта после ремонта', 'medium'],
  ['buyer-iphone-deposit', 'Задаток владельцу аккаунта', 'hard'],
  ['buyer-macbook-corporate-lock', 'MacBook с корпоративной блокировкой', 'medium'],
  ['buyer-overpayment-scam', 'Шестая миссия', 'easy'],
].map(([id, title, difficulty]) => ({
  id,
  version: 1,
  slug: id,
  role: 'buyer',
  title,
  description: `Описание: ${title}`,
  difficulty,
  estimatedMinutes: 5,
  maxDecisions: 3,
})) as Scenario[]

function renderPanel() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <FeaturedMissionsPanel onAction={vi.fn()} />
    </QueryClientProvider>
  )
}

describe('FeaturedMissionsPanel', () => {
  beforeEach(() => {
    vi.mocked(client.fetchScenarios).mockResolvedValue(scenarios)
  })

  it('shows five missions with their own icons and difficulty XP', async () => {
    const { container } = renderPanel()

    await screen.findByText('MacBook с корпоративной блокировкой')
    expect(screen.queryByText('Шестая миссия')).toBeNull()

    const rows = Array.from(container.querySelectorAll<HTMLButtonElement>('.mission-row'))
    expect(rows).toHaveLength(5)

    const expected = [
      ['mission-earbuds-case.png', '+10 XP'],
      ['mission-delivery-truck.png', '+15 XP'],
      ['mission-gpu-card.png', '+15 XP'],
      ['mission-phone.png', '+20 XP'],
      ['mission-laptop-lock.png', '+15 XP'],
    ]

    await waitFor(() => {
      rows.forEach((row, index) => {
        const icon = row.querySelector<HTMLImageElement>('.mission-tile .pixel-icon')
        expect(icon?.getAttribute('src')).toContain(expected[index][0])
        expect(within(row).getByText(expected[index][1])).toBeTruthy()
      })
    })
  })
})
