import { beforeEach, describe, expect, it } from 'vitest'
import {
  AVATAR_ORDER,
  DEFAULT_SETTINGS,
  avatarUrl,
  effectsGain,
  musicGain,
  nextAvatar,
  useSettingsStore,
} from '../model/settingsStore'

describe('настройки приложения', () => {
  beforeEach(() => {
    localStorage.clear()
    useSettingsStore.setState({ settings: DEFAULT_SETTINGS })
  })

  it('сохраняет только переданные поля', () => {
    useSettingsStore.getState().applySettings({ theme: 'dark', masterVolume: 30 })

    const { settings } = useSettingsStore.getState()
    expect(settings.theme).toBe('dark')
    expect(settings.masterVolume).toBe(30)
    // Остальные настройки не должны потеряться.
    expect(settings.playerName).toBe(DEFAULT_SETTINGS.playerName)
    expect(settings.notifyDaily).toBe(DEFAULT_SETTINGS.notifyDaily)
  })

  it('возвращает значения по умолчанию, сохраняя имя и аватар', () => {
    useSettingsStore
      .getState()
      .applySettings({ theme: 'dark', playerName: 'Тест', avatar: 'leader-2' })
    useSettingsStore.getState().resetSettings()

    // Сброс предпочтений не должен стирать личность игрока.
    expect(useSettingsStore.getState().settings).toEqual({
      ...DEFAULT_SETTINGS,
      playerName: 'Тест',
      avatar: 'leader-2',
    })
  })

  it('перебирает аватары по кругу', () => {
    const first = AVATAR_ORDER[0]
    let avatar = first

    for (let step = 0; step < AVATAR_ORDER.length; step++) {
      avatar = nextAvatar(avatar)
    }

    expect(avatar).toBe(first)
    expect(avatarUrl(first)).toContain('.png')
  })

  it('считает громкость эффектов как произведение ползунков', () => {
    expect(effectsGain({ ...DEFAULT_SETTINGS, masterVolume: 100, effectsVolume: 100 })).toBe(1)
    expect(effectsGain({ ...DEFAULT_SETTINGS, masterVolume: 50, effectsVolume: 50 })).toBeCloseTo(
      0.25
    )
    // Выключенный общий звук глушит эффекты независимо от их ползунка.
    expect(effectsGain({ ...DEFAULT_SETTINGS, masterVolume: 0, effectsVolume: 100 })).toBe(0)
  })

  it('учитывает общий уровень и выключатель фоновой музыки', () => {
    expect(musicGain({ ...DEFAULT_SETTINGS, masterVolume: 100, musicVolume: 50 })).toBe(0.5)
    expect(
      musicGain({ ...DEFAULT_SETTINGS, masterVolume: 100, musicVolume: 100, muteMenuMusic: true })
    ).toBe(0)
  })
})
