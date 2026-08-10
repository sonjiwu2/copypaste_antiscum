import { useCallback, useEffect, useRef } from 'react'
import { useSettingsStore } from '../../../entities/settings'
import { useUserProgressStore } from '../../../entities/user-progress'
import { useScenarios } from '../../../shared/api'
import { useToastStore } from '../../toast'
import { useNotificationsStore } from './notificationsStore'
import { isFirstRun, planNotifications } from './rules'

/** Пауза между сообщениями: подсказка живёт 2.6 с, следующая не должна её сбивать. */
const MESSAGE_GAP_MS = 3000

/** Дата по местному времени в том же виде, в каком её хранит прогресс. */
function todayKey(): string {
  const date = new Date()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')

  return `${date.getFullYear()}-${month}-${day}`
}

function isToday(timestamp: number): boolean {
  const moment = new Date(timestamp)
  const now = new Date()

  return (
    moment.getFullYear() === now.getFullYear() &&
    moment.getMonth() === now.getMonth() &&
    moment.getDate() === now.getDate()
  )
}

/**
 * Уведомления приложения.
 *
 * Сообщения показываются внутри приложения через ту же ленту подсказок, что и
 * остальные события. Системные push здесь невозможны: для них нужен сервер
 * рассылки и service worker, а тумблер, который ничего не делает, хуже, чем
 * его отсутствие.
 *
 * Решение о том, что показать, принимает `planNotifications`; хук только
 * собирает состояние, показывает сообщения и запоминает показанное.
 */
export function useAppNotifications(): void {
  const notifyDaily = useSettingsStore((state) => state.settings.notifyDaily)
  const notifyNewMissions = useSettingsStore((state) => state.settings.notifyNewMissions)
  const notifyProgress = useSettingsStore((state) => state.settings.notifyProgress)

  const { scenarios } = useScenarios()

  const completedMissions = useUserProgressStore((state) => state.completedMissions)
  const completedWeeklyTests = useUserProgressStore((state) => state.completedWeeklyTests)
  const unlockedAchievements = useUserProgressStore((state) => state.unlockedAchievements)
  const level = useUserProgressStore((state) => state.level)
  const streakDays = useUserProgressStore((state) => state.streakDays)

  // Очередь: лента показывает одно сообщение и заменяет предыдущее, поэтому
  // события, случившиеся одновременно, разводятся по времени.
  const queue = useRef<string[]>([])
  const draining = useRef(false)

  const enqueue = useCallback((messages: string[]) => {
    queue.current.push(...messages)

    if (draining.current || queue.current.length === 0) {
      return
    }

    draining.current = true

    const drain = () => {
      const next = queue.current.shift()
      if (next === undefined) {
        draining.current = false

        return
      }

      useToastStore.getState().showToast(next)
      window.setTimeout(drain, MESSAGE_GAP_MS)
    }

    drain()
  }, [])

  useEffect(() => {
    const store = useNotificationsStore.getState()
    const scenarioIds = scenarios.map((scenario) => scenario.id)

    // Каталог ещё не пришёл: без него «новых миссий» не отличить от пустого
    // ответа, а на первом запуске нечего запоминать.
    if (scenarioIds.length === 0) {
      return
    }

    // Первый запуск: запоминаем текущее состояние молча. Иначе пользователь
    // получил бы разом весь каталог «новых» миссий и все прежние награды.
    if (isFirstRun(store)) {
      store.markScenariosSeen(scenarioIds)
      store.markAchievementsAnnounced(unlockedAchievements ?? [])
      store.markLevelAnnounced(level ?? 1)

      return
    }

    const activeToday =
      Object.values(completedMissions ?? {}).some((record) => isToday(record.completedAt)) ||
      Object.values(completedWeeklyTests ?? {}).some((record) => isToday(record.completedAt))

    const plan = planNotifications({
      enabled: { daily: notifyDaily, newMissions: notifyNewMissions, progress: notifyProgress },
      announced: store,
      scenarioIds,
      level: level ?? 1,
      unlockedAchievements: unlockedAchievements ?? [],
      streakDays: streakDays ?? 1,
      activeToday,
      today: todayKey(),
    })

    if (plan.markLevel !== null) {
      store.markLevelAnnounced(plan.markLevel)
    }
    if (plan.markAchievements.length > 0) {
      store.markAchievementsAnnounced(plan.markAchievements)
    }
    if (plan.markScenarios.length > 0) {
      store.markScenariosSeen(plan.markScenarios)
    }
    if (plan.markDailyReminder !== null) {
      store.markDailyReminder(plan.markDailyReminder)
    }

    enqueue(plan.messages)
  }, [
    completedMissions,
    completedWeeklyTests,
    enqueue,
    level,
    notifyDaily,
    notifyNewMissions,
    notifyProgress,
    scenarios,
    streakDays,
    unlockedAchievements,
  ])
}
