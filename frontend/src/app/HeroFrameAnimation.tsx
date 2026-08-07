import { useEffect, useState } from 'react'
import { preloadImage } from '../shared/lib/preloadImage'

const HERO_FRAMES = Array.from(
  { length: 16 },
  (_, index) =>
    `/assets/anti-scam/hero-frames/frame-${String(index + 1).padStart(2, '0')}.png?v=20260804-16f`
)

// Longer calm holds slow the scene without stretching the blink. The authored
// motion still advances decisively, while one complete loop now takes 3.705s.
const FRAME_DURATIONS = [
  760, 210, 185, 105, 195, 300, 145, 130, 115, 110, 110, 110, 110, 130, 170, 820,
]
const FRAME_END_TIMES = FRAME_DURATIONS.reduce<number[]>((timeline, duration) => {
  timeline.push((timeline.at(-1) ?? 0) + duration)
  return timeline
}, [])
const LOOP_DURATION = FRAME_END_TIMES.at(-1) ?? 0

export function HeroFrameAnimation() {
  const [ready, setReady] = useState(false)
  const [frame, setFrame] = useState(0)
  const [previousFrame, setPreviousFrame] = useState<number | null>(null)

  useEffect(() => {
    let cancelled = false
    Promise.all(HERO_FRAMES.map(preloadImage)).then(() => {
      if (!cancelled) setReady(true)
    })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!ready || window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

    let animationFrame = 0
    let displayedFrame = 0
    const startedAt = performance.now()

    const play = (now: number) => {
      const elapsed = (now - startedAt) % LOOP_DURATION
      const nextFrame = FRAME_END_TIMES.findIndex((frameEnd) => elapsed < frameEnd)

      if (nextFrame !== displayedFrame && nextFrame !== -1) {
        setPreviousFrame(displayedFrame)
        displayedFrame = nextFrame
        setFrame(nextFrame)
      }

      animationFrame = window.requestAnimationFrame(play)
    }

    animationFrame = window.requestAnimationFrame(play)
    return () => window.cancelAnimationFrame(animationFrame)
  }, [ready])

  return (
    <div
      className={`hero-frame-player${ready ? ' hero-frame-player--ready' : ''}`}
      aria-hidden='true'
    >
      {HERO_FRAMES.map((source, index) => (
        <img
          key={source}
          src={source}
          alt=''
          decoding='async'
          draggable={false}
          className={`hero-frame pixel-art${
            index === frame
              ? ' hero-frame--active'
              : index === previousFrame
                ? ' hero-frame--previous'
                : ''
          }`}
        />
      ))}
    </div>
  )
}
