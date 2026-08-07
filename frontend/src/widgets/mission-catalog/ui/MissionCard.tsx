import { MissionArtwork, type Mission } from '../../../entities/mission'
import { missionStatus } from '../../../features/filter-missions'

export interface MissionCardProps {
  mission: Mission
  activeMission: string | null
  onOpen: (mission: Mission) => void
  onLaunch?: (mission: Mission) => void
  onLocked: (mission: Mission) => void
}

export function MissionCard({ mission, activeMission, onOpen, onLaunch, onLocked }: MissionCardProps) {
  const status = missionStatus(mission, activeMission)
  const actionLabel =
    status === 'in-progress'
      ? 'Продолжить'
      : status === 'completed'
        ? 'Пройдено'
        : status === 'locked'
          ? 'Заблокировано'
          : 'Начать'

  return (
    <article
      className={`catalog-card catalog-card--${mission.tone}${
        status === 'locked' ? ' catalog-card--locked' : ''
      }`}
      onClick={() => (status === 'locked' ? onLocked(mission) : onOpen(mission))}
      style={{ cursor: 'pointer' }}
    >
      <div className='catalog-card__top'>
        <MissionArtwork mission={mission} />
        <div className='catalog-card__copy'>
          <h3>{mission.title}</h3>
          <p>{mission.description}</p>
          <div className='catalog-tags'>
            <span className={`catalog-tag catalog-tag--${mission.difficulty.toLowerCase()}`}>
              {mission.difficulty}
            </span>
            <span className='catalog-tag catalog-tag--role'>{mission.role}</span>
          </div>
        </div>
      </div>

      {status === 'in-progress' && (
        <div className='catalog-progress'>
          <span style={{ width: '60%' }} />
          <small>60% Завершено</small>
        </div>
      )}

      {status === 'locked' && <div className='catalog-unlock'>🔒 {mission.unlock}</div>}

      <div className='catalog-card__footer'>
        <strong>
          <span aria-hidden='true'>★</span> +{mission.xp} XP
        </strong>
        <button
          type='button'
          onClick={(e) => {
            e.stopPropagation()
            if (status === 'locked') {
              onLocked(mission)
            } else if (onLaunch) {
              onLaunch(mission)
            } else {
              onOpen(mission)
            }
          }}
        >
          {actionLabel}
        </button>
      </div>
    </article>
  )
}
