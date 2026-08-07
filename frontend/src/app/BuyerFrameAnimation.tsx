import { useEffect, useState, type CSSProperties } from 'react'
import { preloadImage } from '../shared/lib/preloadImage'

const FRAME_DURATIONS = [
  540, 180, 150, 110, 150, 190, 160, 140, 130, 120, 140, 170, 210, 240, 280, 520,
]

const FRAME_PLACEMENTS = [
  { scale: 1, x: 0, y: 0 },
  { scale: 1.00923, x: 0.3, y: 2.4 },
  { scale: 1.00923, x: 0.1, y: 2.4 },
  { scale: 1.00661, x: 0.5, y: 1.6 },
  { scale: 1.00487, x: 0.9, y: 1.2 },
  { scale: 1.00314, x: 1.2, y: 0.6 },
  { scale: 0.99969, x: 1.6, y: -0.1 },
  { scale: 0.99969, x: 1.6, y: -0.1 },
  { scale: 0.96982, x: -0.7, y: -14.2 },
  { scale: 0.97785, x: -0.3, y: -11.9 },
  { scale: 0.96578, x: 0.1, y: -14.7 },
  { scale: 0.98889, x: 1, y: -8.6 },
  { scale: 0.99399, x: 0.9, y: -6.8 },
  { scale: 0.99058, x: 0.8, y: -7.9 },
  { scale: 1.01402, x: -0.1, y: -2.4 },
  { scale: 0.98973, x: 1, y: -8.1 },
]

const BUYER_FRAMES = FRAME_PLACEMENTS.map((placement, index) => ({
  ...placement,
  source: `/assets/anti-scam/buyer-frames/frame-${String(index + 1).padStart(2, '0')}.png?v=20260805-16f`,
}))

const FRAME_END_TIMES = FRAME_DURATIONS.reduce<number[]>((timeline, duration) => {
  timeline.push((timeline.at(-1) ?? 0) + duration)
  return timeline
}, [])
const LOOP_DURATION = FRAME_END_TIMES.at(-1) ?? 0

type BuyerFrameStyle = CSSProperties & {
  '--buyer-frame-scale': number
  '--buyer-frame-x': string
  '--buyer-frame-y': string
}

export function BuyerFrameAnimation() {
  const [ready, setReady] = useState(false)
  const [frame, setFrame] = useState(0)
  const [previousFrame, setPreviousFrame] = useState<number | null>(null)

  useEffect(() => {
    let cancelled = false
    Promise.all(BUYER_FRAMES.map(({ source }) => preloadImage(source))).then(() => {
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
    <div className='buyer-frame-player' aria-hidden='true'>
      {BUYER_FRAMES.map(({ source, scale, x, y }, index) => {
        const style: BuyerFrameStyle = {
          '--buyer-frame-scale': scale,
          '--buyer-frame-x': `${x}px`,
          '--buyer-frame-y': `${y}px`,
        }

        return (
          <img
            key={source}
            src={source}
            alt=''
            decoding='async'
            draggable={false}
            style={style}
            className={`buyer-character-frame pixel-art${
              index === frame
                ? ' buyer-character-frame--active'
                : index === previousFrame
                  ? ' buyer-character-frame--previous'
                  : ''
            }`}
          />
        )
      })}
    </div>
  )
}
