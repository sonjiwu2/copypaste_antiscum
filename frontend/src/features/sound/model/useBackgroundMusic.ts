import { useEffect, useRef } from 'react'
import { musicGain, useSettingsStore } from '../../../entities/settings'
import { AUDIO_ROOT } from '../../../shared/config/assets'

const BACKGROUND_TRACK = `${AUDIO_ROOT}/pixel-dawn.mp3`

/**
 * Фоновая музыка приложения.
 *
 * Браузеры запрещают воспроизведение со звуком до первого действия
 * пользователя. Поэтому хук сначала пробует запустить дорожку сам, а при
 * блокировке ждёт первое нажатие мыши или клавиши. Настройки громкости
 * применяются сразу после сохранения на странице настроек.
 */
export function useBackgroundMusic(): void {
  const settings = useSettingsStore((state) => state.settings)
  const audioRef = useRef<HTMLAudioElement | null>(null)
  const interactionUnlockedRef = useRef(false)

  useEffect(() => {
    const audio = new Audio(BACKGROUND_TRACK)
    audio.loop = true
    audio.preload = 'auto'
    audio.volume = musicGain(useSettingsStore.getState().settings)
    audioRef.current = audio

    const removeUnlockListeners = () => {
      document.removeEventListener('pointerdown', unlockPlayback)
      document.removeEventListener('keydown', unlockPlayback)
    }

    const startPlayback = async () => {
      if (musicGain(useSettingsStore.getState().settings) <= 0) {
        return
      }

      try {
        await audio.play()
        interactionUnlockedRef.current = true
        removeUnlockListeners()
      } catch {
        // Автовоспроизведение заблокировано. Первое действие пользователя
        // повторит попытку через обработчики ниже.
      }
    }

    function unlockPlayback() {
      interactionUnlockedRef.current = true
      removeUnlockListeners()
      void startPlayback()
    }

    document.addEventListener('pointerdown', unlockPlayback)
    document.addEventListener('keydown', unlockPlayback)
    void startPlayback()

    return () => {
      removeUnlockListeners()
      audio.pause()
      audio.src = ''
      audioRef.current = null
      interactionUnlockedRef.current = false
    }
  }, [])

  useEffect(() => {
    const audio = audioRef.current
    if (!audio) {
      return
    }

    const gain = musicGain(settings)
    audio.volume = Math.min(1, Math.max(0, gain))

    if (gain <= 0) {
      audio.pause()
    } else if (interactionUnlockedRef.current && audio.paused) {
      void audio.play().catch(() => undefined)
    }
  }, [settings])
}
