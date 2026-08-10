import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import type { UserProgressState, CompletedMissionRecord } from './types'

export function calculateLevel(totalXp: number): number {
  return Math.floor(totalXp / 100) + 1
}

/**
 * XP за прохождение сценария с указанным результатом.
 *
 * Формула объявлена здесь одна на всё приложение: экран прохождения показывает
 * ту же награду, которую потом записывает store.
 */
export function earnedXpFor(baseXp: number, score: number): number {
  return Math.max(10, Math.round(baseXp * (score / 100)))
}

/**
 * Дата в виде ГГГГ-ММ-ДД по местному времени.
 *
 * Серию считаем по календарным дням пользователя, поэтому берём локальную
 * дату, а не UTC: иначе вечерняя активность попадала бы в следующий день.
 */
function dateKey(daysAgo = 0): string {
  const date = new Date()
  date.setDate(date.getDate() - daysAgo)

  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')

  return `${date.getFullYear()}-${month}-${day}`
}

const today = () => dateKey()
const yesterday = () => dateKey(1)

/** Текущий локальный календарный день для экранов ежедневных наград. */
export const currentLocalDateKey = () => today()

/** Пороги наград. Открытая награда обратно не закрывается. */
const ACHIEVEMENT_THRESHOLDS = {
  firstMission: 1,
  fiveMissions: 5,
  perfectScore: 100,
  streakDays: 3,
} as const

/**
 * Добавляет награду за серию, если она набрана.
 *
 * Серия растёт и от миссии, и от недельного теста, поэтому проверка вынесена
 * отдельно и вызывается из обоих мест.
 */
function withStreakAchievement(unlocked: Iterable<string>, streakDays: number): string[] {
  const achievements = new Set(unlocked)

  if (streakDays >= ACHIEVEMENT_THRESHOLDS.streakDays) {
    achievements.add('streak_3')
  }

  return Array.from(achievements)
}

/**
 * Длина серии после активности сегодня.
 *
 * Серия растёт только при активности день за днём: пропущенный день начинает
 * отсчёт заново, а второе прохождение за сутки её не удваивает.
 */
function nextStreak(currentStreak: number, lastActiveDate?: string): number {
  if (lastActiveDate === today()) {
    return currentStreak
  }

  return lastActiveDate === yesterday() ? currentStreak + 1 : 1
}

export const useUserProgressStore = create<UserProgressState>()(
  persist(
    (set, get) => ({
      totalXp: 0,
      level: 1,
      completedMissions: {},
      streakDays: 1,
      lastActiveDate: today(),
      unlockedAchievements: [],
      completedWeeklyTests: {},
      lastDailyRewardDate: null,

      recordMissionCompletion: (missionId: string, score: number, baseXp = 20) => {
        const state = get()
        const completedMap = state.completedMissions ?? {}
        const previous: CompletedMissionRecord | undefined = completedMap[missionId]

        // Начисляем только прирост награды: повторное прохождение с худшим
        // результатом не должно отнимать уже полученный XP.
        const earnedXp = earnedXpFor(baseXp, score)
        const previousXp = previous?.xpEarned ?? 0
        const totalXp = (state.totalXp ?? 0) + Math.max(0, earnedXp - previousXp)

        const completedMissions: Record<string, CompletedMissionRecord> = {
          ...completedMap,
          [missionId]: {
            bestScore: Math.max(previous?.bestScore ?? 0, score),
            xpEarned: Math.max(previousXp, earnedXp),
            completedAt: Date.now(),
            attemptsCount: (previous?.attemptsCount ?? 0) + 1,
          },
        }

        const achievements = new Set(state.unlockedAchievements ?? [])
        const completedCount = Object.keys(completedMissions).length
        if (completedCount >= ACHIEVEMENT_THRESHOLDS.firstMission) {
          achievements.add('first_mission')
        }
        if (completedCount >= ACHIEVEMENT_THRESHOLDS.fiveMissions) {
          achievements.add('five_missions')
        }
        if (score >= ACHIEVEMENT_THRESHOLDS.perfectScore) {
          achievements.add('perfect_score')
        }

        const streakDays = nextStreak(state.streakDays ?? 1, state.lastActiveDate)

        set({
          totalXp,
          level: calculateLevel(totalXp),
          completedMissions,
          streakDays,
          lastActiveDate: today(),
          unlockedAchievements: withStreakAchievement(achievements, streakDays),
        })
      },

      recordWeeklyTestCompletion: (weekStart: string, score: number, earnedXp: number) => {
        const state = get()
        const completedWeeklyTests = state.completedWeeklyTests ?? {}
        const previous = completedWeeklyTests[weekStart]
        const safeXp = Math.max(0, Math.round(earnedXp))
        const totalXp = (state.totalXp ?? 0) + Math.max(0, safeXp - (previous?.xpEarned ?? 0))
        const streakDays = nextStreak(state.streakDays ?? 1, state.lastActiveDate)

        set({
          totalXp,
          level: calculateLevel(totalXp),
          completedWeeklyTests: {
            ...completedWeeklyTests,
            [weekStart]: {
              score: Math.max(previous?.score ?? 0, score),
              xpEarned: Math.max(previous?.xpEarned ?? 0, safeXp),
              completedAt: previous?.completedAt ?? Date.now(),
            },
          },
          streakDays,
          lastActiveDate: today(),
          unlockedAchievements: withStreakAchievement(
            state.unlockedAchievements ?? [],
            streakDays
          ),
        })
      },

      claimDailyReward: (rewardXp = 25) => {
        const state = get()
        const claimDate = today()
        if (state.lastDailyRewardDate === claimDate) {
          return false
        }

        const safeXp = Math.max(0, Math.round(rewardXp))
        if (safeXp === 0) {
          return false
        }

        const totalXp = (state.totalXp ?? 0) + safeXp
        const streakDays = nextStreak(state.streakDays ?? 1, state.lastActiveDate)
        set({
          totalXp,
          level: calculateLevel(totalXp),
          lastDailyRewardDate: claimDate,
          streakDays,
          lastActiveDate: claimDate,
          unlockedAchievements: withStreakAchievement(
            state.unlockedAchievements ?? [],
            streakDays
          ),
        })

        return true
      },

      checkAndUpdateStreak: () => {
        const state = get()
        if (state.lastActiveDate === today()) {
          return
        }

        set({
          streakDays: nextStreak(state.streakDays ?? 1, state.lastActiveDate),
          lastActiveDate: today(),
        })
      },

      resetProgress: () => {
        set({
          totalXp: 0,
          level: 1,
          completedMissions: {},
          streakDays: 1,
          lastActiveDate: today(),
          unlockedAchievements: [],
          completedWeeklyTests: {},
          lastDailyRewardDate: null,
        })
      },
    }),
    {
      name: 'antiscam.user_progress.v1',
      storage: createJSONStorage(() => localStorage),
    }
  )
)
