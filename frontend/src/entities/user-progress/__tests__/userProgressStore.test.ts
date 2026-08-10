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

  it('выдаёт награду за серию, когда она дошла до порога', () => {
    // Активность вчера при серии в два дня: сегодняшнее прохождение делает третий.
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    const key = `${yesterday.getFullYear()}-${String(yesterday.getMonth() + 1).padStart(2, '0')}-${String(yesterday.getDate()).padStart(2, '0')}`

    useUserProgressStore.setState({ streakDays: 2, lastActiveDate: key })
    useUserProgressStore.getState().recordMissionCompletion('buyer-fake-delivery', 60, 20)

    const updated = useUserProgressStore.getState()
    expect(updated.streakDays).toBe(3)
    expect(updated.unlockedAchievements).toContain('streak_3')
  })

  it('не выдаёт награду за серию раньше порога', () => {
    useUserProgressStore.getState().recordMissionCompletion('buyer-fake-delivery', 60, 20)

    expect(useUserProgressStore.getState().unlockedAchievements).not.toContain('streak_3')
  })

  it('начисляет ежедневную награду только один раз за календарный день', () => {
    const firstClaim = useUserProgressStore.getState().claimDailyReward(25)
    const secondClaim = useUserProgressStore.getState().claimDailyReward(25)

    const updated = useUserProgressStore.getState()
    expect(firstClaim).toBe(true)
    expect(secondClaim).toBe(false)
    expect(updated.totalXp).toBe(25)
    expect(updated.level).toBe(1)
    expect(updated.lastDailyRewardDate).toBeTruthy()
  })

  it('позволяет забрать ежедневную награду заново после полного сброса', () => {
    expect(useUserProgressStore.getState().claimDailyReward()).toBe(true)
    expect(useUserProgressStore.getState().claimDailyReward()).toBe(false)

    useUserProgressStore.getState().resetProgress()

    expect(useUserProgressStore.getState().lastDailyRewardDate).toBeNull()
    expect(useUserProgressStore.getState().claimDailyReward()).toBe(true)
    expect(useUserProgressStore.getState().totalXp).toBe(25)
  })
})
