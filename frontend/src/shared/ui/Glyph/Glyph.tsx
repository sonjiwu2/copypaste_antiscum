/**
 * Мелкие векторные знаки, для которых нет пиксельного ассета.
 *
 * Это служебные элементы управления: стрелка «дальше», лампочка совета,
 * действия в настройках. Они нарисованы по той же сетке 16×16 и с
 * `crispEdges`, поэтому в общем ряду выглядят как остальной пиксель-арт.
 * Для всего, что есть в наборе иконок, используется `PixelIcon`.
 */
export type GlyphName = 'chevron' | 'bulb' | 'trash' | 'warning' | 'pencil' | 'key'

export interface GlyphProps {
  name: GlyphName
  size?: number
  className?: string
}

export function Glyph({ name, size = 20, className }: GlyphProps) {
  const shared = {
    width: size,
    height: size,
    viewBox: '0 0 16 16',
    fill: 'none',
    xmlns: 'http://www.w3.org/2000/svg',
    shapeRendering: 'crispEdges' as const,
    className: className ? `pixel-icon ${className}` : 'pixel-icon',
    'aria-hidden': true,
  }

  switch (name) {
    case 'chevron':
      return (
        <svg {...shared}>
          <path d='M4 1h3v2h2v2h2v2h2v2h-2v2H9v2H7v2H4v-3h2v-2h2V6H6V4H4z' fill='currentColor' />
        </svg>
      )

    case 'bulb':
      return (
        <svg {...shared}>
          <path
            d='M6 0h4v1h2v2h2v6h-1v2h-2v2H5v-2H3V9H2V4h1V2h3zM5 14h6v2H5z'
            fill='currentColor'
          />
          <path d='M6 3h3v1h2v3H9V5H6z' fill='#fff' opacity='.72' />
          <path d='M0 5h2v2H0zM14 5h2v2h-2zM1 0h2v2H1zM13 0h2v2h-2z' fill='currentColor' />
        </svg>
      )

    case 'trash':
      return (
        <svg {...shared}>
          <path d='M6 0h4v2h4v2H2V2h4z' fill='currentColor' />
          <path d='M3 5h10v11H3z' fill='currentColor' />
          <path d='M5 7h2v7H5zM9 7h2v7H9z' fill='#fff' opacity='.8' />
        </svg>
      )

    case 'warning':
      return (
        <svg {...shared}>
          <path d='M8 1 15 14H1z' fill='currentColor' />
          <path d='M7 6h2v4H7zM7 11h2v2H7z' fill='#fff' />
        </svg>
      )

    case 'pencil':
      return (
        <svg {...shared}>
          <path d='M11 2h1v1h1v1h1v1h-1v1h-1V5h-1V4h-1V3h1z' fill='currentColor' />
          <path d='M4 9h1V8h1V7h1V6h1V5h1v1h1v1h1v1h-1v1h-1v1H7v1H6v1H5v1H4z' fill='currentColor' />
          <path d='M2 12h2v2H2v1H1v-2h1z' fill='currentColor' />
        </svg>
      )

    case 'key':
      return (
        <svg {...shared}>
          <path d='M6 1h4v1h1v3h-1v1H6V5H5V2h1z' fill='currentColor' />
          <path d='M7 3h2v1H7z' fill='#fff' />
          <path d='M7 6h2v10H7z' fill='currentColor' />
          <path d='M9 8h3v2H9zM9 11h2v2H9z' fill='currentColor' />
        </svg>
      )
  }
}
