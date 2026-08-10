import {
  DIFFICULTY_LABELS,
  MissionArtwork,
  ROLE_LABELS,
  STATUS_ACTIONS,
  type Mission,
} from '../../../entities/mission'
import { missionStatus } from '../../../features/filter-missions'
import { PixelIcon } from '../../../shared/ui/PixelIcon'

export interface MissionCardProps {
  mission: Mission
  activeMission: string | null
  onOpen: (mission: Mission) => void
  onLaunch?: (mission: Mission) => void
  onLocked: (mission: Mission) => void
}

export function MissionCard({
  mission,
  activeMission,
  onOpen,
  onLaunch,
  onLocked,
}: MissionCardProps) {
  const status = missionStatus(mission, activeMission)
  const isLocked = status === 'locked'

  return (
    <article
      className={`catalog-card catalog-card--${mission.tone}${isLocked ? ' catalog-card--locked' : ''}`}
    >
      <div className='catalog-card__top'>
        <MissionArtwork mission={mission} />
        <div className='catalog-card__copy'>
          {/*
            Кнопка внутри заголовка растянута по всей карточке через ::after,
            поэтому нажатие в любом месте открывает описание, а в фокус попадает
            один понятный элемент вместо кликабельного блока.
          */}
          <h3>
            <button
              className='catalog-card__open'
              type='button'
              onClick={() => (isLocked ? onLocked(mission) : onOpen(mission))}
            >
              {mission.title}
            </button>
          </h3>
          <p>{mission.description}</p>
          <div className='catalog-tags'>
            <span className={`catalog-tag catalog-tag--${mission.difficulty}`}>
              {DIFFICULTY_LABELS[mission.difficulty]}
            </span>
            <span className='catalog-tag catalog-tag--role'>{ROLE_LABELS[mission.role]}</span>
          </div>
        </div>
      </div>

      {status === 'in-progress' && (
        <p className='catalog-progress'>
          <PixelIcon name='in-progress' size={14} />
          Миссия начата
        </p>
      )}

      {isLocked && mission.unlock && (
        <p className='catalog-unlock'>
          <PixelIcon name='locked' size={16} />
          {mission.unlock}
        </p>
      )}

      <div className='catalog-card__footer'>
        <strong className='catalog-card__xp'>
          <PixelIcon name='xp' size={18} />+{mission.xp} XP
        </strong>
        <button
          className='catalog-card__launch'
          type='button'
          disabled={isLocked}
          onClick={() => (onLaunch ?? onOpen)(mission)}
        >
          {STATUS_ACTIONS[status]}
        </button>
      </div>
    </article>
  )
}
