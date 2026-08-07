export type AvatarTone = 'blue' | 'green' | 'orange'

const OUTLINE = '#111827'
const SKIN = '#ffc29b'
const HAIR = '#21191a'

export function PixelAvatar({ tone = 'blue' }: { tone?: AvatarTone }) {
  const palette = {
    blue: ['#176dcd', '#48a7ff', '#eef5ff'],
    green: ['#087c58', '#34bc86', '#20a06f'],
    orange: ['#c45e20', '#ff9a4e', '#ef8734'],
  }[tone]

  return (
    <svg
      className='pixel-avatar'
      viewBox='0 0 32 32'
      shapeRendering='crispEdges'
      aria-hidden='true'
    >
      <path d='M7 3h18v2h3v9H4V7h3z' fill={OUTLINE} />
      <path d='M9 3h13v2h3v5H5V7h4z' fill={palette[0]} />
      <path d='M11 4h9v2h4v2H8V6h3z' fill={palette[1]} />
      <path d='M5 11h22v4h2v9h-4v4H8v-2H4V15h1z' fill={HAIR} />
      <path d='M9 12h14v2h3v9h-3v3H8V15h1z' fill={SKIN} />
      <path d='M12 16h3v4h-3zM20 16h3v4h-3z' fill={OUTLINE} />
      <path d='M16 22h6v2h-6z' fill='#d94b3e' />
      <path d='M6 27h20v5H6z' fill={palette[2]} />
      <path d='M14 27h4v5h-4z' fill='#fff' opacity='.86' />
    </svg>
  )
}
