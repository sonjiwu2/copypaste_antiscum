import { renderHook, act } from '@testing-library/react'
import { useMissionFilters } from './useMissionFilters'
import type { Mission } from '../../../entities/mission'

const TEST_MISSIONS: Mission[] = [
  {
    id: 'buyer-fake-delivery',
    title: 'Ссылка на доставку',
    description: 'Покупка товара с доставкой: продавец предлагает оплату по ссылке.',
    difficulty: 'medium',
    role: 'buyer',
    category: 'shipping',
    xp: 15,
    section: 'featured',
    status: 'not-started',
    tone: 'purple',
    icon: 'delivery',
  },
  {
    id: 'seller-gpu-return-swap',
    title: 'Подмена видеокарты при возврате',
    description: 'Продажа RTX 4070: покупатель заявляет о дефекте и возвращает подделку.',
    difficulty: 'hard',
    role: 'seller',
    category: 'payment',
    xp: 20,
    section: 'featured',
    status: 'not-started',
    tone: 'green',
    icon: 'monitor',
  },
]

describe('useMissionFilters custom hook', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it("should initialize with default 'both' role filter and all missions", () => {
    const { result } = renderHook(() =>
      useMissionFilters({ allMissions: TEST_MISSIONS, initialRole: 'both' })
    )

    expect(result.current.role).toBe('both')
    expect(result.current.difficulty).toBe('all')
    expect(result.current.visibleMissions.length).toBe(TEST_MISSIONS.length)
    expect(result.current.isFiltering).toBe(false)
  })

  it('should filter missions by role when role is changed', () => {
    const { result } = renderHook(() =>
      useMissionFilters({ allMissions: TEST_MISSIONS, initialRole: 'both' })
    )

    act(() => {
      result.current.setRole('buyer')
    })

    expect(result.current.role).toBe('buyer')
    expect(result.current.isFiltering).toBe(true)
    expect(
      result.current.visibleMissions.every((m) => m.role === 'buyer' || m.role === 'both')
    ).toBe(true)
  })

  it('should filter missions by difficulty', () => {
    const { result } = renderHook(() =>
      useMissionFilters({ allMissions: TEST_MISSIONS, initialRole: 'both' })
    )

    act(() => {
      result.current.setDifficulty('hard')
    })

    expect(result.current.difficulty).toBe('hard')
    expect(result.current.visibleMissions.every((m) => m.difficulty === 'hard')).toBe(true)
  })

  it('should toggle categories and filter accordingly', () => {
    const { result } = renderHook(() =>
      useMissionFilters({ allMissions: TEST_MISSIONS, initialRole: 'both' })
    )

    act(() => {
      result.current.toggleCategory('payment')
    })

    expect(result.current.categories).toContain('payment')
    expect(result.current.visibleMissions.every((m) => m.category === 'payment')).toBe(true)
  })

  it('should reset all filters when clearFilters is called', () => {
    const { result } = renderHook(() =>
      useMissionFilters({ allMissions: TEST_MISSIONS, initialRole: 'both' })
    )

    act(() => {
      result.current.setRole('seller')
      result.current.setDifficulty('medium')
      result.current.toggleCategory('phishing')
    })

    expect(result.current.isFiltering).toBe(true)

    act(() => {
      result.current.clearFilters()
    })

    expect(result.current.role).toBe('both')
    expect(result.current.difficulty).toBe('all')
    expect(result.current.categories.length).toBe(0)
    expect(result.current.isFiltering).toBe(false)
  })

  it('should trigger onToast and onLaunch when beginMission is executed', () => {
    const handleToast = vi.fn()
    const handleLaunch = vi.fn()

    const { result } = renderHook(() =>
      useMissionFilters({
        allMissions: TEST_MISSIONS,
        initialRole: 'both',
        onToast: handleToast,
        onLaunch: handleLaunch,
      })
    )

    act(() => {
      result.current.beginMission(TEST_MISSIONS[0])
    })

    expect(handleToast).toHaveBeenCalledWith(expect.stringContaining('начата'))
    expect(handleLaunch).toHaveBeenCalledWith(TEST_MISSIONS[0].id)
    expect(result.current.activeMission).toBe(TEST_MISSIONS[0].id)
  })
})
