import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { PROVIDED_ASSET_ROOT } from '../../../shared/config/assets'

export type RoleCardProps = {
  role: 'buyer' | 'seller'
  title: string
  description: string
  badge: string
  missions: number
  completed: number
  accuracy: string
  onAction: () => void
}

export function RoleCard({
  role,
  title,
  description,
  badge,
  missions,
  completed,
  accuracy,
  onAction,
}: RoleCardProps) {
  const isBuyer = role === 'buyer'
  return (
    <article className={`role-card role-card--${role} pixel-card`}>
      <div className='role-character' aria-hidden='true'>
        <span className='character-shadow' />
        <img
          src={
            isBuyer
              ? '/assets/anti-scam/buyer-character-transparent.png'
              : '/assets/anti-scam/seller-character-transparent.png'
          }
          alt=''
          className={`role-character__image role-character__image--${role} pixel-art`}
          draggable={false}
        />
      </div>
      <img
        src={`${PROVIDED_ASSET_ROOT}/${isBuyer ? 'buyer-role.png' : 'seller-role.png'}`}
        alt=''
        className='role-card__icon pixel-art'
      />
      <div className='role-card__content'>
        <h2>{title}</h2>
        <p>{description}</p>
        <div className='role-badge'>
          <PixelIcon name='shield' size={22} />
          <span>{badge}</span>
        </div>
      </div>
      <div className='role-stats' aria-label={`Статистика: ${title}`}>
        <div>
          <span>Миссий</span>
          <strong>{missions}</strong>
        </div>
        <div>
          <span>Пройдено</span>
          <strong>{completed}</strong>
        </div>
        <div>
          <span>Точность</span>
          <strong>{accuracy}</strong>
        </div>
      </div>
      <button className='role-cta' type='button' onClick={onAction}>
        <span>Продолжить как {isBuyer ? 'Покупатель' : 'Селлер'}</span>
        <PixelIcon name='chevron' size={20} />
      </button>
    </article>
  )
}
