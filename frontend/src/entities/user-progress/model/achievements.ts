import type { PixelIconName } from '../../../shared/ui/PixelIcon'

export interface Achievement {
  id: string
  title: string
  description: string
  icon: PixelIconName
}

/**
 * Список наград одинаков для страницы прогресса и для уведомлений.
 *
 * Раньше он лежал внутри страницы, и уведомление о полученной награде не могло
 * назвать её по имени, не дублируя тексты.
 */
export const ACHIEVEMENTS: Achievement[] = [
  {
    id: 'first_mission',
    title: 'Первый шаг',
    description: 'Успешно завершите вашу первую миссию по безопасности.',
    icon: 'star',
  },
  {
    id: 'five_missions',
    title: 'Опытный защитник',
    description: 'Пройдите 5 миссий и закрепите навыки распознавания мошенников.',
    icon: 'shield',
  },
  {
    id: 'perfect_score',
    title: 'Мастер безопасности',
    description: 'Завершите любую миссию со 100% результатом без ошибок.',
    icon: 'trophy',
  },
  {
    id: 'streak_3',
    title: 'Страж порядка',
    description: 'Сохраняйте активность 3 дня подряд.',
    icon: 'streak',
  },
]

const TITLE_BY_ID = new Map(ACHIEVEMENTS.map((item) => [item.id, item.title]))

/** Название награды для уведомления. Неизвестный код возвращается как есть. */
export function achievementTitle(id: string): string {
  return TITLE_BY_ID.get(id) ?? id
}
