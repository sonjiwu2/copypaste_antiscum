import { ART_ROOT, ICON_ROOT } from '../../config/assets'

/**
 * Реестр пиксельных иконок.
 *
 * Имя описывает роль иконки в интерфейсе, а не то, что на ней нарисовано: так
 * замена картинки не требует правок в компонентах. Часть значков приходит из
 * папки иллюстраций — эти рисунки появились раньше набора иконок.
 */
const ICON_FILES = {
  // Навигация, профиль, награды
  missions: `${ICON_ROOT}/mission-sword.png`,
  progress: `${ICON_ROOT}/nav-progress.png`,
  rules: `${ICON_ROOT}/nav-rules.png`,
  settings: `${ICON_ROOT}/nav-settings.png`,
  profile: `${ICON_ROOT}/profile-avatar.png`,
  xp: `${ICON_ROOT}/xp-token.png`,
  star: `${ICON_ROOT}/star.png`,
  trophy: `${ART_ROOT}/daily-trophy.png`,
  rewards: `${ART_ROOT}/rewards-gift.png`,
  streak: `${ICON_ROOT}/sparkles.png`,
  leaders: `${ICON_ROOT}/account-user.png`,

  // Роль и сложность
  'role-buyer': `${ICON_ROOT}/role-buyer-bag.png`,
  'role-seller': `${ICON_ROOT}/role-seller-store.png`,
  'role-both': `${ICON_ROOT}/shield.png`,
  'difficulty-easy': `${ICON_ROOT}/difficulty-easy.png`,
  'difficulty-medium': `${ICON_ROOT}/difficulty-medium.png`,
  'difficulty-hard': `${ICON_ROOT}/difficulty-hard.png`,

  // Каталог миссий
  filter: `${ICON_ROOT}/filter-funnel.png`,
  featured: `${ART_ROOT}/featured-flag.png`,
  new: `${ICON_ROOT}/sparkles.png`,
  'in-progress': `${ICON_ROOT}/refresh.png`,
  locked: `${ART_ROOT}/tip-lock.png`,
  reset: `${ICON_ROOT}/refresh.png`,
  close: `${ICON_ROOT}/close-cross.png`,

  // Тематика сценариев
  delivery: `${ICON_ROOT}/mission-delivery-truck.png`,
  phone: `${ICON_ROOT}/mission-phone.png`,
  console: `${ICON_ROOT}/mission-gamepad.png`,
  monitor: `${ICON_ROOT}/mission-monitor.png`,
  courier: `${ICON_ROOT}/mission-courier-scooter.png`,
  payment: `${ICON_ROOT}/mission-money-bag.png`,
  account: `${ICON_ROOT}/account-user.png`,
  listing: `${ICON_ROOT}/mission-desktop.png`,
  earbuds: `${ICON_ROOT}/mission-earbuds-case.png`,
  gpu: `${ICON_ROOT}/mission-gpu-card.png`,
  'laptop-lock': `${ICON_ROOT}/mission-laptop-lock.png`,
  handheld: `${ICON_ROOT}/mission-handheld.png`,
  'email-hook': `${ICON_ROOT}/mission-fake-email.png`,
  qr: `${ICON_ROOT}/mission-qr-code.png`,
  sms: `${ICON_ROOT}/mission-sms-code.png`,
  camera: `${ICON_ROOT}/mission-camera.png`,

  // Прохождение сценария
  shield: `${ICON_ROOT}/shield.png`,
  events: `${ICON_ROOT}/panel-eye.png`,
  signals: `${ICON_ROOT}/panel-magnifier.png`,
  tips: `${ICON_ROOT}/panel-checklist.png`,
  session: `${ICON_ROOT}/panel-stopwatch.png`,

  // Настройки
  'sound-on': `${ICON_ROOT}/sound-on.png`,
  'sound-off': `${ICON_ROOT}/sound-off.png`,
  music: `${ICON_ROOT}/music-note.png`,
  bell: `${ICON_ROOT}/bell.png`,
  save: `${ICON_ROOT}/save-floppy.png`,
  database: `${ICON_ROOT}/database.png`,
} as const

export type PixelIconName = keyof typeof ICON_FILES

export function pixelIconUrl(name: PixelIconName): string {
  return ICON_FILES[name]
}
