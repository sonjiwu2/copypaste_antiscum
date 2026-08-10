import { ART_ROOT, ICON_ROOT } from '../../../shared/config/assets'
import { Glyph } from '../../../shared/ui/Glyph'
import { PixelIcon } from '../../../shared/ui/PixelIcon'

export type UtilityCardKind = 'tip' | 'sos' | 'weekly' | 'rewards'

export type UtilityCardProps = {
  kind: UtilityCardKind
  title: string
  description: string
  action: string
  /** Имя файла в выбранной коллекции `public/assets`. */
  art: string
  artCollection?: 'art' | 'icons'
  claimed?: boolean
  pending?: boolean
  onAction: () => void
}

export function UtilityCard({
  kind,
  title,
  description,
  action,
  art,
  artCollection = 'art',
  claimed = false,
  pending = false,
  onAction,
}: UtilityCardProps) {
  const artRoot = artCollection === 'icons' ? ICON_ROOT : ART_ROOT

  return (
    <article
      className={`utility-card utility-card--${kind}${claimed ? ' utility-card--claimed' : ''}${pending ? ' utility-card--pending' : ''} pixel-card`}
    >
      <h2 className='utility-card__title'>
        {kind === 'tip' && <Glyph name='bulb' size={22} />}
        {title}
      </h2>
      <div className='utility-card__body'>
        <div className='utility-card__art'>
          <img src={`${artRoot}/${art}`} alt='' className='pixel-art' />
          {kind === 'sos' && (
            <span className='utility-card__sos-badge' aria-hidden='true'>
              SOS
            </span>
          )}
          {kind === 'rewards' && claimed && (
            <span className='utility-card__claimed-badge' aria-hidden='true'>
              ✓
            </span>
          )}
        </div>
        <p>{description}</p>
      </div>
      <button
        className='utility-card__action'
        type='button'
        disabled={pending}
        onClick={onAction}
      >
        {pending && <PixelIcon name='locked' size={16} />}
        {action}
        {!pending && <Glyph name='chevron' size={17} />}
      </button>
    </article>
  )
}
