export type IconName =
  | 'star'
  | 'chart'
  | 'book'
  | 'gear'
  | 'flag'
  | 'shield'
  | 'bag'
  | 'store'
  | 'trophy'
  | 'target'
  | 'fire'
  | 'users'
  | 'help'
  | 'headset'
  | 'lock'
  | 'chevron'
  | 'bolt'
  | 'bulb'
  | 'calendar'
  | 'gift'
  | 'message'
  | 'box'
  | 'user'

export function PixelIcon({ name, size = 22 }: { name: IconName; size?: number }) {
  const common = {
    width: size,
    height: size,
    viewBox: '0 0 16 16',
    fill: 'none',
    xmlns: 'http://www.w3.org/2000/svg',
    shapeRendering: 'crispEdges' as const,
    className: 'pixel-icon',
    'aria-hidden': true,
  }

  switch (name) {
    case 'star':
      return (
        <svg {...common}>
          <path d='M6 0h4v3h3v3h3v4h-3v3h-3v3H6v-3H3v-3H0V6h3V3h3z' fill='currentColor' />
          <path d='M7 3h2v3h3v2H9v3H7V8H4V6h3z' fill='#fff' opacity='.62' />
        </svg>
      )
    case 'chart':
      return (
        <svg {...common}>
          <path d='M1 14h14v2H1zM2 8h3v6H2zM7 2h3v12H7zM12 5h3v9h-3z' fill='currentColor' />
          <path d='M3 9h1v4H3zM8 3h1v9H8zM13 6h1v6h-1z' fill='#fff' opacity='.48' />
        </svg>
      )
    case 'book':
      return (
        <svg {...common}>
          <path d='M2 1h11v2h1v12H3v-1H1V2h1z' fill='currentColor' />
          <path d='M4 3h7v2H4zM4 7h7v1H4zM4 10h5v1H4z' fill='#fff' opacity='.88' />
        </svg>
      )
    case 'gear':
      return (
        <svg {...common}>
          <path
            d='M6 0h4v2h2V1h2v2h1v3h1v4h-1v3h-2v2h-3v1H6v-1H3v-2H1v-3H0V6h1V3h2V1h2v1h1z'
            fill='currentColor'
          />
          <path d='M6 5h4v1h1v4h-1v1H6v-1H5V6h1zm1 2v2h2V7z' fill='#fff' opacity='.9' />
        </svg>
      )
    case 'flag':
      return (
        <svg {...common}>
          <path d='M1 0h2v16H1zM3 1h10v2h2v6h-2v2H3z' fill='currentColor' />
          <path d='M5 3h6v2H5z' fill='#fff' opacity='.45' />
        </svg>
      )
    case 'shield':
      return (
        <svg {...common}>
          <path d='M6 0h4v1h4v2h2v6h-1v3h-2v2h-2v1H5v-1H3v-2H1V9H0V3h2V2h4z' fill='currentColor' />
          <path d='M3 4h10v5h-1v2h-2v2H6v-1H4v-2H3z' fill='#fff' opacity='.28' />
          <path d='M4 7h2v2h2V7h4v2h-2v2H7v-1H5V9H4z' fill='#fff' />
        </svg>
      )
    case 'bag':
      return (
        <svg {...common}>
          <path d='M3 5h10v2h2v9H1V7h2z' fill='currentColor' />
          <path d='M5 1h6v1h1v5h-2V3H6v4H4V2h1z' fill='currentColor' />
          <path d='M3 8h10v5H3z' fill='#fff' opacity='.2' />
        </svg>
      )
    case 'store':
      return (
        <svg {...common}>
          <path d='M2 0h12v2h1v5h-1v9H2V7H1V2h1z' fill='currentColor' />
          <path d='M3 2h2v3h2V2h2v3h2V2h2v3h-1v2h-2V6H8v1H6V6H4v1H3z' fill='#fff' opacity='.76' />
          <path d='M5 9h6v5H5z' fill='#fff' opacity='.72' />
        </svg>
      )
    case 'trophy':
      return (
        <svg {...common}>
          <path d='M3 1h10v2h3v5h-2v2h-3v2H9v2h4v2H3v-2h4v-2H5v-2H2V8H0V3h3z' fill='currentColor' />
          <path
            d='M5 3h6v4h-1v2H6V7H5zM1 5h2v2H2V6H1zM13 5h2v1h-1v1h-1z'
            fill='#fff'
            opacity='.62'
          />
        </svg>
      )
    case 'target':
      return (
        <svg {...common}>
          <path d='M5 1h6v1h2v2h1v2h1v6h-1v2h-2v1H4v-1H2v-2H1V6h1V4h2V2h1z' fill='currentColor' />
          <path d='M6 4h4v1h2v2h1v3h-1v2h-2v1H6v-1H4v-2H3V7h1V5h2z' fill='#fff' />
          <path d='M7 6h3v1h1v3h-1v1H7v-1H6V7h1z' fill='currentColor' />
          <path d='M10 0h6v2h-2v2h-2V2h-2z' fill='currentColor' />
        </svg>
      )
    case 'fire':
      return (
        <svg {...common}>
          <path
            d='M8 0h2v3h2v2h2v3h2v4h-1v2h-2v2H4v-1H2v-2H1V9h1V6h2v3h2V5h1V2h1z'
            fill='currentColor'
          />
          <path d='M8 8h2v2h2v3h-1v1H6v-1H5v-2h1V9h2z' fill='#fff' opacity='.75' />
        </svg>
      )
    case 'users':
      return (
        <svg {...common}>
          <path
            d='M3 1h4v1h1v4H7v1H3V6H2V2h1zM10 2h3v1h1v3h-1v1h-3V6H9V3h1zM1 9h8v1h2v6H0v-6h1zM11 9h3v1h2v6h-4v-5h-1z'
            fill='currentColor'
          />
        </svg>
      )
    case 'help':
      return (
        <svg {...common}>
          <path
            d='M5 0h6v1h2v2h2v2h1v6h-1v2h-2v2h-2v1H5v-1H3v-2H1v-2H0V5h1V3h2V1h2z'
            fill='currentColor'
          />
          <path d='M6 3h4v1h2v3h-1v1H9v2H7V7h2V6h1V5H6zM7 12h2v2H7z' fill='#fff' />
        </svg>
      )
    case 'headset':
      return (
        <svg {...common}>
          <path
            d='M5 1h6v1h2v2h1v2h1v7h-2v2H9v-2h4V7h-1V4h-2V3H6v1H4v3H3v6H0V7h1V5h1V3h3z'
            fill='currentColor'
          />
          <path d='M3 8h3v6H2V9h1zM11 8h3v5h-4V9h1z' fill='currentColor' />
        </svg>
      )
    case 'lock':
      return (
        <svg {...common}>
          <path d='M5 0h6v1h2v6h2v9H1V7h2V2h2zm0 7h6V3h-1V2H6v1H5z' fill='currentColor' />
          <path d='M7 10h2v3H7z' fill='#fff' />
        </svg>
      )
    case 'chevron':
      return (
        <svg {...common}>
          <path d='M4 1h3v2h2v2h2v2h2v2h-2v2H9v2H7v2H4v-3h2v-2h2V6H6V4H4z' fill='currentColor' />
        </svg>
      )
    case 'bolt':
      return (
        <svg {...common}>
          <path d='M8 0h6v3h-2v2h3v3h-2v2h-2v2H9v2H7v2H4v-5H1V8h2V6h2V3h3z' fill='currentColor' />
          <path d='M9 2h2v2H9v2H7v2H5V7h2V5h2z' fill='#fff' opacity='.55' />
        </svg>
      )
    case 'bulb':
      return (
        <svg {...common}>
          <path
            d='M6 0h4v1h2v2h2v6h-1v2h-2v2H5v-2H3V9H2V4h1V2h3zM5 14h6v2H5z'
            fill='currentColor'
          />
          <path d='M6 3h3v1h2v3H9V5H6z' fill='#fff' opacity='.72' />
          <path d='M0 5h2v2H0zM14 5h2v2h-2zM1 0h2v2H1zM13 0h2v2h-2z' fill='currentColor' />
        </svg>
      )
    case 'calendar':
      return (
        <svg {...common}>
          <path d='M2 2h2V0h2v2h4V0h2v2h2v2h2v12H0V4h2z' fill='currentColor' />
          <path d='M2 6h12v8H2z' fill='#fff' />
          <path
            d='M4 8h2v2H4zM7 8h2v2H7zM10 8h2v2h-2zM4 11h2v2H4zM7 11h2v2H7zM10 11h2v2h-2z'
            fill='currentColor'
            opacity='.7'
          />
        </svg>
      )
    case 'gift':
      return (
        <svg {...common}>
          <path d='M1 5h14v4h-1v7H2V9H1zM4 1h3v1h1v2h1V2h1V1h3v1h1v3H2V2h2z' fill='currentColor' />
          <path d='M7 5h2v11H7zM2 6h13v2H2z' fill='#fff' opacity='.55' />
        </svg>
      )
    case 'message':
      return (
        <svg {...common}>
          <path d='M2 1h12v1h2v9h-2v2H8l-4 3v-3H2v-2H0V3h2z' fill='currentColor' />
          <path d='M4 5h2v2H4zM7 5h2v2H7zM10 5h2v2h-2z' fill='#fff' />
        </svg>
      )
    case 'box':
      return (
        <svg {...common}>
          <path d='M2 4 8 0l6 4v8l-6 4-6-4z' fill='currentColor' />
          <path d='M4 4 8 2l4 2-4 2zM4 6l3 2v5l-3-2zM9 8l3-2v5l-3 2z' fill='#fff' opacity='.78' />
        </svg>
      )
    case 'user':
      return (
        <svg {...common}>
          <path d='M5 0h6v1h2v6h-2v2h2v1h2v6H1v-6h2V9h2V7H3V1h2z' fill='currentColor' />
          <path d='M5 2h6v4H9v1H7V6H5zM4 11h8v3H4z' fill='#fff' opacity='.9' />
        </svg>
      )
  }
}
