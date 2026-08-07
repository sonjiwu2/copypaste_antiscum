import { describe, it, expect, beforeEach } from 'vitest'
import { useUserProgressStore, calculateLevel } from '../model/userProgressStore'

describe('userProgressStore', () => {
  beforeEach(() => {
    useUserProgressStore.getState().resetProgress()
  })

  it('calculates levels correctly based on total XP', () => {
    expect(calculateLevel(0)).toBe(1)
    expect(calculateLevel(50)).toBe(1)
    expect(calculateLevel(100)).toBe(2)
    expect(calculateLevel(250)).toBe(3)
  })

  it('records mission completion and updates total XP and level', () => {
    const store = useUserProgressStore.getState()
    expect(store.totalXp).toBe(0)

    store.recordMissionCompletion('buyer-fake-delivery', 100, 20)

    const updated = useUserProgressStore.getState()
    expect(updated.totalXp).toBe(20)
    expect(updated.completedMissions['buyer-fake-delivery']).toBeDefined()
    expect(updated.completedMissions['buyer-fake-delivery'].bestScore).toBe(100)
    expect(updated.completedMissions['buyer-fake-delivery'].attemptsCount).toBe(1)
  })

  it('updates bestScore and adds only XP delta on replaying mission with better score', () => {
    const store = useUserProgressStore.getState()

    // 1st attempt: 50% score -> 10 XP
    store.recordMissionCompletion('buyer-fake-delivery', 50, 20)
    expect(useUserProgressStore.getState().totalXp).toBe(10)

    // 2nd attempt: 100% score -> 20 XP (+10 delta)
    useUserProgressStore.getState().recordMissionCompletion('buyer-fake-delivery', 100, 20)
    expect(useUserProgressStore.getState().totalXp).toBe(20)
    expect(useUserProgressStore.getState().completedMissions['buyer-fake-delivery'].bestScore).toBe(100)
    expect(useUserProgressStore.getState().completedMissions['buyer-fake-delivery'].attemptsCount).toBe(2)
  })

  it('unlocks achievements on milestones', () => {
    const store = useUserProgressStore.getState()

    store.recordMissionCompletion('buyer-fake-delivery', 100, 20)
    const updated = useUserProgressStore.getState()

    expect(updated.unlockedAchievements).toContain('first_mission')
    expect(updated.unlockedAchievements).toContain('perfect_score')
  })
})
