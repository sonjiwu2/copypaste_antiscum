import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { ART_ROOT, ICON_ROOT } from '../../../shared/config/assets'
import type { AppSettings, AvatarId, SettingsState } from './types'

/**
 * Настройки хранятся только в браузере.
 *
 * Бэкенд их не принимает: это оформление и предпочтения одного устройства,
 * а не часть прогресса, который сервер считает по попыткам.
 */
export const DEFAULT_SETTINGS: AppSettings = {
  language: 'ru',

  showDailyTip: true,

  theme: 'light',

  masterVolume: 80,
  effectsVolume: 70,
  musicVolume: 40,
  muteMenuMusic: false,

  notifyDaily: true,
  notifyNewMissions: true,
  notifyProgress: true,

  playerName: 'Алекс',
  avatar: 'profile',
}

/** Стандартные предпочтения не должны стирать личность игрока. */
export function defaultSettingsForProfile(
  profile: Pick<AppSettings, 'playerName' | 'avatar'>
): AppSettings {
  return {
    ...DEFAULT_SETTINGS,
    playerName: profile.playerName,
    avatar: profile.avatar,
  }
}

/** Порядок аватаров в карточке аккаунта: кнопка перебирает список по кругу. */
export const AVATAR_ORDER: AvatarId[] = ['profile', 'leader-1', 'leader-2', 'leader-3']

const AVATAR_FILES: Record<AvatarId, string> = {
  profile: `${ICON_ROOT}/profile-avatar.png`,
  'leader-1': `${ART_ROOT}/leader-avatar-1.png`,
  'leader-2': `${ART_ROOT}/leader-avatar-2.png`,
  'leader-3': `${ART_ROOT}/leader-avatar-3.png`,
}

export function avatarUrl(avatar: AvatarId): string {
  return AVATAR_FILES[avatar] ?? AVATAR_FILES.profile
}

export function nextAvatar(current: AvatarId): AvatarId {
  const index = AVATAR_ORDER.indexOf(current)

  return AVATAR_ORDER[(index + 1) % AVATAR_ORDER.length]
}

/** Громкость эффекта с учётом общего ползунка. Возвращает 0–1. */
export function effectsGain(settings: AppSettings): number {
  return (settings.masterVolume / 100) * (settings.effectsVolume / 100)
}

/** Громкость фоновой музыки с учётом общего ползунка и выключателя. */
export function musicGain(settings: AppSettings): number {
  if (settings.muteMenuMusic) {
    return 0
  }

  return (settings.masterVolume / 100) * (settings.musicVolume / 100)
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set, get) => ({
      settings: DEFAULT_SETTINGS,
      registeredAt: new Date().toISOString(),

      applySettings: (changes) => set({ settings: { ...get().settings, ...changes } }),

      resetSettings: () => set({ settings: defaultSettingsForProfile(get().settings) }),
    }),
    {
      name: 'antiscam.settings.v1',
      storage: createJSONStorage(() => localStorage),
      // Новые поля появляются со значением по умолчанию, а не как undefined:
      // иначе после обновления приложения тумблер остался бы без состояния.
      // Удалённые настройки при этом отбрасываются, чтобы в браузере не
      // копились ключи функций, которых в приложении больше нет.
      merge: (persisted, current) => {
        const saved = persisted as Partial<SettingsState> | undefined
        const settings = { ...DEFAULT_SETTINGS }

        for (const key of Object.keys(settings) as Array<keyof AppSettings>) {
          const value = saved?.settings?.[key]
          if (value !== undefined) {
            settings[key] = value as never
          }
        }

        return { ...current, ...saved, settings }
      },
    }
  )
)
