import type { PixelIconName } from '../../../shared/ui/PixelIcon'
import type { Difficulty, MissionRole, MissionStatus, ScamCategory } from './types'

/** Русские подписи для кодов миссии. Интерфейс берёт текст только отсюда. */

export const DIFFICULTY_LABELS: Record<Difficulty, string> = {
  easy: 'Легко',
  medium: 'Средне',
  hard: 'Сложно',
}

export const DIFFICULTY_ORDER: Difficulty[] = ['easy', 'medium', 'hard']

export const DIFFICULTY_ICONS: Record<Difficulty, PixelIconName> = {
  easy: 'difficulty-easy',
  medium: 'difficulty-medium',
  hard: 'difficulty-hard',
}

export const ROLE_LABELS: Record<MissionRole, string> = {
  buyer: 'Покупатель',
  seller: 'Продавец',
  both: 'Оба',
}

export const ROLE_ICONS: Record<MissionRole, PixelIconName> = {
  buyer: 'role-buyer',
  seller: 'role-seller',
  both: 'role-both',
}

export const CATEGORY_ORDER: ScamCategory[] = [
  'fake-listings',
  'payment',
  'phishing',
  'account-takeover',
  'shipping',
  'impersonation',
]

export const CATEGORY_LABELS: Record<ScamCategory, string> = {
  'fake-listings': 'Фейковые объявления',
  payment: 'Платёжные схемы',
  phishing: 'Фишинг и ссылки',
  'account-takeover': 'Угон аккаунтов',
  shipping: 'Доставка и отправка',
  impersonation: 'Подмена личности',
}

export const CATEGORY_ICONS: Record<ScamCategory, PixelIconName> = {
  'fake-listings': 'listing',
  payment: 'payment',
  phishing: 'phone',
  'account-takeover': 'account',
  shipping: 'delivery',
  impersonation: 'shield',
}

export const STATUS_LABELS: Record<MissionStatus, string> = {
  'not-started': 'Не начато',
  'in-progress': 'В процессе',
  completed: 'Завершено',
  locked: 'Заблокировано',
}

/** Подпись кнопки запуска. Зависит от того, начата ли миссия. */
export const STATUS_ACTIONS: Record<MissionStatus, string> = {
  'not-started': 'Начать',
  'in-progress': 'Продолжить',
  completed: 'Пройти снова',
  locked: 'Заблокировано',
}
