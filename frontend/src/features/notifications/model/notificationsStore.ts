import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

/**
 * Память уведомлений: о чём пользователю уже сообщали.
 *
 * Без неё напоминание повторялось бы при каждом открытии страницы, а награда
 * объявлялась бы заново после перезагрузки. Настройки здесь не хранятся —
 * включённость тумблеров живёт в настройках приложения.
 */
export interface NotificationsState {
  /** Дата последнего показанного напоминания в формате ГГГГ-ММ-ДД. */
  lastDailyReminder: string | null
  /** Сценарии, которые пользователь уже видел в каталоге. */
  seenScenarioIds: string[]
  /** Награды, о получении которых уже сообщали. */
  announcedAchievements: string[]
  /** Уровень, о достижении которого уже сообщали. */
  announcedLevel: number

  markDailyReminder: (date: string) => void
  markScenariosSeen: (ids: string[]) => void
  markAchievementsAnnounced: (ids: string[]) => void
  markLevelAnnounced: (level: number) => void
  resetNotifications: () => void
}

const EMPTY = {
  lastDailyReminder: null,
  seenScenarioIds: [],
  announcedAchievements: [],
  announcedLevel: 0,
} satisfies Pick<
  NotificationsState,
  'lastDailyReminder' | 'seenScenarioIds' | 'announcedAchievements' | 'announcedLevel'
>

export const useNotificationsStore = create<NotificationsState>()(
  persist(
    (set, get) => ({
      ...EMPTY,

      markDailyReminder: (date) => set({ lastDailyReminder: date }),

      markScenariosSeen: (ids) =>
        set({ seenScenarioIds: Array.from(new Set([...get().seenScenarioIds, ...ids])) }),

      markAchievementsAnnounced: (ids) =>
        set({
          announcedAchievements: Array.from(new Set([...get().announcedAchievements, ...ids])),
        }),

      markLevelAnnounced: (level) => set({ announcedLevel: level }),

      resetNotifications: () => set({ ...EMPTY }),
    }),
    {
      name: 'antiscam.notifications.v1',
      storage: createJSONStorage(() => localStorage),
    }
  )
)
