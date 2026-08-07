import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import type { UserProgressState, CompletedMissionRecord } from './types'

export function calculateLevel(totalXp: number): number {
  return Math.floor(totalXp / 100) + 1
}

function getTodayString(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function getYesterdayString(): string {
  const d = new Date()
  d.setDate(d.getDate() - 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export const useUserProgressStore = create<UserProgressState>()(
  persist(
    (set, get) => ({
      totalXp: 0,
      level: 1,
      completedMissions: {},
      streakDays: 1,
      lastActiveDate: getTodayString(),
      unlockedAchievements: [],

      recordMissionCompletion: (missionId: string, score: number, baseXp = 20) => {
        const state = get()
        const completedMap = state.completedMissions ?? {}
        const today = getTodayString()
        const prevRecord: CompletedMissionRecord | undefined = completedMap[missionId]

        const earnedXp = Math.max(10, Math.round(baseXp * (score / 100)))
        const prevXp = prevRecord?.xpEarned ?? 0
        const xpDelta = Math.max(0, earnedXp - prevXp)

        const newTotalXp = (state.totalXp ?? 0) + xpDelta
        const newLevel = calculateLevel(newTotalXp)

        const updatedMissions: Record<string, CompletedMissionRecord> = {
          ...completedMap,
          [missionId]: {
            bestScore: Math.max(prevRecord?.bestScore ?? 0, score),
            xpEarned: Math.max(prevXp, earnedXp),
            completedAt: Date.now(),
            attemptsCount: (prevRecord?.attemptsCount ?? 0) + 1,
          },
        }

        // Вычисление ачивок
        const newAchievements = new Set(state.unlockedAchievements ?? [])
        const completedCount = Object.keys(updatedMissions).length
        if (completedCount >= 1) newAchievements.add('first_mission')
        if (completedCount >= 5) newAchievements.add('five_missions')
        if (score === 100) newAchievements.add('perfect_score')

        // Обновление стРИКа при активности
        let newStreak = state.streakDays ?? 1
        if (state.lastActiveDate === getYesterdayString()) {
          newStreak += 1
        } else if (state.lastActiveDate !== today) {
          newStreak = 1
        }

        set({
          totalXp: newTotalXp,
          level: newLevel,
          completedMissions: updatedMissions,
          streakDays: newStreak,
          lastActiveDate: today,
          unlockedAchievements: Array.from(newAchievements),
        })
      },

      checkAndUpdateStreak: () => {
        const state = get()
        const today = getTodayString()
        const yesterday = getYesterdayString()

        if (state.lastActiveDate === today) return

        if (state.lastActiveDate === yesterday) {
          set({ streakDays: (state.streakDays ?? 1) + 1, lastActiveDate: today })
        } else {
          set({ streakDays: 1, lastActiveDate: today })
        }
      },

      resetProgress: () => {
        set({
          totalXp: 0,
          level: 1,
          completedMissions: {},
          streakDays: 1,
          lastActiveDate: getTodayString(),
          unlockedAchievements: [],
        })
      },
    }),
    {
      name: 'antiscam.user_progress.v1',
      storage: createJSONStorage(() => localStorage),
    }
  )
)
