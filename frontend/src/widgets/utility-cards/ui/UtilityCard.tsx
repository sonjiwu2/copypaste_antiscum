import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { PROVIDED_ASSET_ROOT } from '../../../shared/config/assets'

export type UtilityCardProps = {
  kind: 'tip' | 'daily' | 'weekly' | 'rewards'
  title: string
  description: string
  action: string
  asset: string
  onAction: () => void
}

export function UtilityCard({
  kind,
  title,
  description,
  action,
  asset,
  onAction,
}: UtilityCardProps) {
  return (
    <article className={`utility-card utility-card--${kind} pixel-card`}>
      <div className='utility-card__art'>
        <img src={`${PROVIDED_ASSET_ROOT}/${asset}`} alt='' className='pixel-art' />
      </div>
      <div className='utility-card__content'>
        <h2>
          {kind === 'tip' && <PixelIcon name='bulb' size={22} />}
          {title}
        </h2>
        <p>{description}</p>
        <button type='button' onClick={onAction}>
          {action}
          <PixelIcon name='chevron' size={17} />
        </button>
      </div>
    </article>
  )
}
