import type { Mission } from '../../../entities/mission'
import { MissionCard } from './MissionCard'

export interface MissionSectionBlockProps {
  icon: string
  title: string
  missions: Mission[]
  activeMission: string | null
  onOpen: (mission: Mission) => void
  onLaunch?: (mission: Mission) => void
  onLocked: (mission: Mission) => void
  onViewAll: () => void
}

export function MissionSectionBlock({
  icon,
  title,
  missions,
  activeMission,
  onOpen,
  onLaunch,
  onLocked,
  onViewAll,
}: MissionSectionBlockProps) {
  if (!missions.length) return null
  return (
    <section className='catalog-section'>
      <div className='catalog-section__heading'>
        <h2>
          <span aria-hidden='true'>{icon}</span>
          {title}
        </h2>
        <button type='button' onClick={onViewAll}>
          Посмотреть все
        </button>
      </div>
      <div className='catalog-grid'>
        {missions.map((mission) => (
          <MissionCard
            key={mission.id}
            mission={mission}
            activeMission={activeMission}
            onOpen={onOpen}
            onLaunch={onLaunch}
            onLocked={onLocked}
          />
        ))}
      </div>
    </section>
  )
}
