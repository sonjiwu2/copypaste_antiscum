/** Аватары, доступные в настройках профиля. */
export type AvatarId = 'profile' | 'leader-1' | 'leader-2' | 'leader-3'

/** Оформление приложения. Светлая тема — исходный вид. */
export type Theme = 'light' | 'dark'

export interface AppSettings {
  /** Пока в интерфейсе только русский. Поле оставлено под будущие языки. */
  language: 'ru'

  // Общие
  showDailyTip: boolean

  // Интерфейс
  theme: Theme

  // Звук: 0–100
  masterVolume: number
  effectsVolume: number
  musicVolume: number
  muteMenuMusic: boolean

  // Уведомления
  notifyDaily: boolean
  notifyNewMissions: boolean
  notifyProgress: boolean

  // Профиль
  playerName: string
  avatar: AvatarId
}

export interface SettingsState {
  settings: AppSettings
  /** Дата первого запуска. Показывается в карточке аккаунта. */
  registeredAt: string
  /** Сохраняет изменённые поля. Остальные остаются как были. */
  applySettings: (changes: Partial<AppSettings>) => void
  /** Возвращает предпочтения по умолчанию, сохраняя имя и аватар игрока. */
  resetSettings: () => void
}
