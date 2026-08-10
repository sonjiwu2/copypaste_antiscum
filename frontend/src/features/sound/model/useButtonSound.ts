import { useEffect, useRef } from 'react'
import { effectsGain, useSettingsStore } from '../../../entities/settings'

/** Пиковая громкость щелчка при ползунках на максимуме. */
const PEAK_GAIN = 0.075

/**
 * Короткий щелчок при нажатии любой кнопки.
 *
 * Слушатель один на весь документ: вешать обработчик на каждую кнопку —
 * это десятки подписок и правка каждого компонента.
 */
export function useButtonSound(): void {
  const audioContext = useRef<AudioContext | null>(null)

  useEffect(() => {
    function playButtonSound(event: MouseEvent) {
      const target = event.target
      if (!(target instanceof Element) || !target.closest('button:not(:disabled)')) {
        return
      }

      // Громкость читаем из store на каждый щелчок: иначе смена настройки
      // требовала бы переподписки на события.
      const gainLevel = effectsGain(useSettingsStore.getState().settings) * PEAK_GAIN
      if (gainLevel <= 0) {
        return
      }

      const context =
        !audioContext.current || audioContext.current.state === 'closed'
          ? new AudioContext()
          : audioContext.current
      audioContext.current = context

      if (context.state === 'suspended') {
        void context.resume()
      }

      const startedAt = context.currentTime
      const gain = context.createGain()
      // Экспоненциальная кривая не принимает нуль, поэтому крайние точки взяты
      // чуть выше: щелчок должен нарастать и гаснуть без клика на границах.
      gain.gain.setValueAtTime(0.0001, startedAt)
      gain.gain.exponentialRampToValueAtTime(gainLevel, startedAt + 0.006)
      gain.gain.exponentialRampToValueAtTime(0.0001, startedAt + 0.095)
      gain.connect(context.destination)

      const tone = context.createOscillator()
      tone.type = 'square'
      tone.frequency.setValueAtTime(620, startedAt)
      tone.frequency.setValueAtTime(820, startedAt + 0.045)
      tone.connect(gain)
      tone.start(startedAt)
      tone.stop(startedAt + 0.095)
    }

    document.addEventListener('click', playButtonSound)

    return () => {
      document.removeEventListener('click', playButtonSound)
      const context = audioContext.current
      audioContext.current = null
      void context?.close()
    }
  }, [])
}
