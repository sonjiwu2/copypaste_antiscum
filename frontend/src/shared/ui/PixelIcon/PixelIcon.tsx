import { pixelIconUrl, type PixelIconName } from './icons'

export interface PixelIconProps {
  name: PixelIconName
  /** Сторона квадрата в пикселях. Картинка вписывается в него целиком. */
  size?: number
  /**
   * Подпись для читалки экрана. По умолчанию иконка декоративная: рядом с ней
   * всегда есть текст, и второе прочтение того же смысла только мешает.
   */
  label?: string
  className?: string
}

export function PixelIcon({ name, size = 22, label, className }: PixelIconProps) {
  return (
    <img
      src={pixelIconUrl(name)}
      width={size}
      height={size}
      alt={label ?? ''}
      aria-hidden={label ? undefined : true}
      className={className ? `pixel-icon ${className}` : 'pixel-icon'}
      draggable={false}
      decoding='async'
    />
  )
}
