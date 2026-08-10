import { describe, expect, it } from 'vitest'
import { isFirstRun, planNotifications, type NotificationInput } from './rules'

const TODAY = '2026-08-09'

function input(overrides: Partial<NotificationInput> = {}): NotificationInput {
  return {
    enabled: { daily: true, newMissions: true, progress: true },
    announced: {
      lastDailyReminder: null,
      seenScenarioIds: ['buyer-fake-delivery'],
      announcedAchievements: [],
      announcedLevel: 1,
    },
    scenarioIds: ['buyer-fake-delivery'],
    level: 1,
    unlockedAchievements: [],
    streakDays: 1,
    activeToday: true,
    today: TODAY,
    ...overrides,
  }
}

describe('правила уведомлений', () => {
  it('первый запуск определяется по неотмеченному уровню', () => {
    expect(isFirstRun({ ...input().announced, announcedLevel: 0 })).toBe(true)
    expect(isFirstRun(input().announced)).toBe(false)
  })

  it('молчит, когда ничего не произошло', () => {
    expect(planNotifications(input()).messages).toEqual([])
  })

  it('сообщает о новых миссиях и запоминает только новые', () => {
    const plan = planNotifications(
      input({ scenarioIds: ['buyer-fake-delivery', 'seller-qr-payment-trap', 'buyer-switch-prepayment'] })
    )

    expect(plan.messages).toEqual(['В каталоге новых миссий: 2. Загляните в раздел «Миссии».'])
    expect(plan.markScenarios).toEqual(['seller-qr-payment-trap', 'buyer-switch-prepayment'])
  })

  it('единственную новую миссию называет в единственном числе', () => {
    const plan = planNotifications(
      input({ scenarioIds: ['buyer-fake-delivery', 'seller-qr-payment-trap'] })
    )

    expect(plan.messages).toEqual(['В каталоге новая миссия. Загляните в раздел «Миссии».'])
  })

  it('сообщает о новом уровне и награде по названию', () => {
    const plan = planNotifications(input({ level: 2, unlockedAchievements: ['first_mission'] }))

    expect(plan.messages).toEqual([
      'Новый уровень: 2. Открыты миссии посложнее.',
      'Награда получена: Первый шаг.',
    ])
    expect(plan.markLevel).toBe(2)
    expect(plan.markAchievements).toEqual(['first_mission'])
  })

  it('напоминает о дне только при отсутствии активности', () => {
    expect(planNotifications(input({ activeToday: false })).messages).toEqual([
      'Сегодня ещё не было миссий. Пройдите одну и начните серию.',
    ])
    expect(planNotifications(input({ activeToday: true })).messages).toEqual([])
  })

  it('упоминает длину серии, когда она уже идёт', () => {
    const plan = planNotifications(input({ activeToday: false, streakDays: 4 }))

    expect(plan.messages[0]).toContain('серия из 4 дней')
  })

  it('не повторяет напоминание в тот же день', () => {
    const plan = planNotifications(
      input({
        activeToday: false,
        announced: { ...input().announced, lastDailyReminder: TODAY },
      })
    )

    expect(plan.messages).toEqual([])
    expect(plan.markDailyReminder).toBeNull()
  })

  it('выключенный тумблер молчит, но событие считает учтённым', () => {
    const plan = planNotifications(
      input({
        enabled: { daily: false, newMissions: false, progress: false },
        scenarioIds: ['buyer-fake-delivery', 'seller-qr-payment-trap'],
        level: 3,
        unlockedAchievements: ['first_mission'],
        activeToday: false,
      })
    )

    expect(plan.messages).toEqual([])
    // Иначе после включения тумблера вывалилась бы вся пропущенная история.
    expect(plan.markScenarios).toEqual(['seller-qr-payment-trap'])
    expect(plan.markLevel).toBe(3)
    expect(plan.markAchievements).toEqual(['first_mission'])
    // Напоминание не показывали — значит и отмечать его нечем.
    expect(plan.markDailyReminder).toBeNull()
  })
})
