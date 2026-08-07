import { useEffect, useRef } from 'react'

export function useButtonSound(): void {
  const audioContext = useRef<AudioContext | null>(null)

  useEffect(() => {
    function playButtonSound(event: MouseEvent) {
      const target = event.target
      if (!(target instanceof Element) || !target.closest('button:not(:disabled)')) return

      const context =
        audioContext.current?.state === 'closed' || !audioContext.current
          ? new AudioContext()
          : audioContext.current
      audioContext.current = context
      if (context.state === 'suspended') void context.resume()

      const startedAt = context.currentTime
      const gain = context.createGain()
      gain.gain.setValueAtTime(0.0001, startedAt)
      gain.gain.exponentialRampToValueAtTime(0.075, startedAt + 0.006)
      gain.gain.exponentialRampToValueAtTime(0.0001, startedAt + 0.095)
      gain.connect(context.destination)

      const firstTone = context.createOscillator()
      firstTone.type = 'square'
      firstTone.frequency.setValueAtTime(620, startedAt)
      firstTone.frequency.setValueAtTime(820, startedAt + 0.045)
      firstTone.connect(gain)
      firstTone.start(startedAt)
      firstTone.stop(startedAt + 0.095)
    }

    document.addEventListener('click', playButtonSound)
    return () => {
      document.removeEventListener('click', playButtonSound)
      const context = audioContext.current
      audioContext.current = null
      if (context) void context.close()
    }
  }, [])
}
