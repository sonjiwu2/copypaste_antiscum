import { achievementTitle } from '../../../entities/user-progress'

/** Что уже показывали пользователю. */
export interface AnnouncedState {
  lastDailyReminder: string | null
  seenScenarioIds: string[]
  announcedAchievements: string[]
  announcedLevel: number
}

/** Состояние приложения на момент проверки. */
export interface NotificationInput {
  enabled: { daily: boolean; newMissions: boolean; progress: boolean }
  announced: AnnouncedState
  scenarioIds: string[]
  level: number
  unlockedAchievements: string[]
  streakDays: number
  /** Была ли сегодня пройдена миссия или недельный тест. */
  activeToday: boolean
  /** Сегодняшняя дата в формате ГГГГ-ММ-ДД. */
  today: string
}

/** Сообщения к показу и то, что нужно запомнить как показанное. */
export interface NotificationPlan {
  messages: string[]
  markScenarios: string[]
  markAchievements: string[]
  markLevel: number | null
  markDailyReminder: string | null
}

const EMPTY_PLAN: NotificationPlan = {
  messages: [],
  markScenarios: [],
  markAchievements: [],
  markLevel: null,
  markDailyReminder: null,
}

/** Первый ли это запуск: до него ничего не объявляли. */
export function isFirstRun(announced: AnnouncedState): boolean {
  return announced.announcedLevel === 0
}

/**
 * Решает, о чём сообщить.
 *
 * Функция чистая: она не трогает ни ленту подсказок, ни хранилище, поэтому
 * правила уведомлений можно проверить тестами без браузера.
 *
 * Выключенный тумблер не копит долг — событие всё равно помечается учтённым,
 * иначе после включения на пользователя вывалилась бы вся пропущенная история.
 */
export function planNotifications(input: NotificationInput): NotificationPlan {
  const { announced, enabled } = input
  const plan: NotificationPlan = { ...EMPTY_PLAN, messages: [] }

  if (input.level > announced.announcedLevel) {
    if (enabled.progress) {
      plan.messages.push(`Новый уровень: ${input.level}. Открыты миссии посложнее.`)
    }
    plan.markLevel = input.level
  }

  const freshAchievements = input.unlockedAchievements.filter(
    (id) => !announced.announcedAchievements.includes(id)
  )
  if (freshAchievements.length > 0) {
    if (enabled.progress) {
      for (const id of freshAchievements) {
        plan.messages.push(`Награда получена: ${achievementTitle(id)}.`)
      }
    }
    plan.markAchievements = freshAchievements
  }

  const addedScenarios = input.scenarioIds.filter((id) => !announced.seenScenarioIds.includes(id))
  if (addedScenarios.length > 0) {
    if (enabled.newMissions) {
      plan.messages.push(
        addedScenarios.length === 1
          ? 'В каталоге новая миссия. Загляните в раздел «Миссии».'
          : `В каталоге новых миссий: ${addedScenarios.length}. Загляните в раздел «Миссии».`
      )
    }
    plan.markScenarios = addedScenarios
  }

  if (enabled.daily && !input.activeToday && announced.lastDailyReminder !== input.today) {
    plan.messages.push(
      input.streakDays > 1
        ? `Сегодня ещё не было миссий. Пройдите одну, чтобы серия из ${input.streakDays} дней не прервалась.`
        : 'Сегодня ещё не было миссий. Пройдите одну и начните серию.'
    )
    plan.markDailyReminder = input.today
  }

  return plan
}
