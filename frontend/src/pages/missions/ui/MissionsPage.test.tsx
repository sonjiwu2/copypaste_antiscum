import { render, screen, within, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MissionsPage } from './MissionsPage'
import * as client from '../../../shared/api/client'

vi.mock('../../../shared/api/client', async () => {
  const actual = await vi.importActual<typeof import('../../../shared/api/client')>(
    '../../../shared/api/client'
  )
  return {
    ...actual,
    fetchScenarios: vi.fn(),
  }
})

const MOCK_SCENARIOS = [
  {
    id: 'buyer-fake-delivery',
    version: 1,
    slug: 'buyer-fake-delivery',
    role: 'buyer' as const,
    title: 'Ссылка на доставку',
    description: 'Покупка товара с доставкой: продавец предлагает оплату по ссылке.',
    difficulty: 'medium' as const,
    estimatedMinutes: 4,
  },
  {
    id: 'seller-gpu-return-swap',
    version: 1,
    slug: 'seller-gpu-return-swap',
    role: 'seller' as const,
    title: 'Подмена видеокарты при возврате',
    description: 'Продажа RTX 4070: покупатель заявляет о дефекте и возвращает подделку.',
    difficulty: 'hard' as const,
    estimatedMinutes: 5,
  },
]

function renderWithProviders(ui: React.ReactElement, { initialEntries }: { initialEntries?: string[] } = {}) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={initialEntries}>{ui}</MemoryRouter>
    </QueryClientProvider>
  )
}

describe('MissionsPage (Integration)', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.mocked(client.fetchScenarios).mockResolvedValue(MOCK_SCENARIOS)
  })

  afterEach(() => {
    cleanup()
  })

  it('renders the sidebar filters and catalog sections', async () => {
    const handleToast = vi.fn()
    const handleLaunch = vi.fn()

    const { container } = renderWithProviders(
      <MissionsPage onToast={handleToast} onLaunch={handleLaunch} />
    )

    expect(await screen.findByRole('heading', { name: /РЕКОМЕНДУЕМЫЕ МИССИИ|FEATURED MISSIONS/i })).toBeTruthy()
    expect(container.querySelector('.mission-filters')).toBeTruthy()
  })

  it('allows filtering missions by role', async () => {
    const user = userEvent.setup({ pointerEventsCheck: 0 })
    const handleToast = vi.fn()
    const handleLaunch = vi.fn()

    const { container } = renderWithProviders(
      <MissionsPage onToast={handleToast} onLaunch={handleLaunch} />,
      { initialEntries: ['/missions'] }
    )

    await screen.findByRole('heading', { name: /РЕКОМЕНДУЕМЫЕ МИССИИ|FEATURED MISSIONS/i })
    const sidebar = container.querySelector('.mission-filters') as HTMLElement
    const buyerButton = within(sidebar).getByRole('button', { name: /Buyer/i })
    await user.click(buyerButton)

    expect(buyerButton.getAttribute('aria-pressed')).toBe('true')
  })

  it('opens the detail dialog when an unlocked card body is clicked and launches the mission', async () => {
    const user = userEvent.setup({ pointerEventsCheck: 0 })
    const handleToast = vi.fn()
    const handleLaunch = vi.fn()

    const { container } = renderWithProviders(
      <MissionsPage onToast={handleToast} onLaunch={handleLaunch} />
    )

    await screen.findByRole('heading', { name: /РЕКОМЕНДУЕМЫЕ МИССИИ|FEATURED MISSIONS/i })
    const cardTop = container.querySelector('.catalog-card__top') as HTMLElement
    expect(cardTop).toBeTruthy()
    await user.click(cardTop)

    const dialog = screen.getByRole('dialog')
    expect(dialog).toBeTruthy()

    const beginBtn = within(dialog).getByRole('button', { name: /Begin Mission/i })
    await user.click(beginBtn)

    expect(handleToast).toHaveBeenCalledWith(expect.stringContaining('started'))
    expect(handleLaunch).toHaveBeenCalledWith('buyer-fake-delivery')
  })

  it('launches the mission directly when the card action button is clicked', async () => {
    const user = userEvent.setup({ pointerEventsCheck: 0 })
    const handleToast = vi.fn()
    const handleLaunch = vi.fn()

    renderWithProviders(<MissionsPage onToast={handleToast} onLaunch={handleLaunch} />)

    await screen.findByRole('heading', { name: /РЕКОМЕНДУЕМЫЕ МИССИИ|FEATURED MISSIONS/i })
    const startButtons = screen.getAllByRole('button', { name: /Начать|Start Mission/i })
    expect(startButtons.length).toBeGreaterThan(0)

    await user.click(startButtons[0])

    expect(handleLaunch).toHaveBeenCalledWith('buyer-fake-delivery')
  })
})
