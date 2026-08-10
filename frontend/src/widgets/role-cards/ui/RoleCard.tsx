import { ROLE_ICONS, ROLE_LABELS, type MissionRole } from '../../../entities/mission'
import { SCENE_ROOT } from '../../../shared/config/assets'
import { Glyph } from '../../../shared/ui/Glyph'
import { PixelIcon } from '../../../shared/ui/PixelIcon'

export type RoleCardProps = {
  role: Exclude<MissionRole, 'both'>
  description: string
  badge: string
  missions: number | string
  completed: number | string
  accuracy: string
  onAction: () => void
}

export function RoleCard({
  role,
  description,
  badge,
  missions,
  completed,
  accuracy,
  onAction,
}: RoleCardProps) {
  const title = ROLE_LABELS[role]

  return (
    <article className={`role-card role-card--${role} pixel-card`}>
      <div className='role-character' aria-hidden='true'>
        <span className='character-shadow' />
        <img
          src={`${SCENE_ROOT}/${role}-character.png`}
          alt=''
          className={`role-character__image role-character__image--${role} pixel-art`}
          draggable={false}
        />
      </div>
      <PixelIcon name={ROLE_ICONS[role]} size={48} className='role-card__icon' />
      <div className='role-card__content'>
        <h2>{title.toUpperCase()}</h2>
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
        <span>Продолжить как {title.toLowerCase()}</span>
        <Glyph name='chevron' size={20} />
      </button>
    </article>
  )
}
