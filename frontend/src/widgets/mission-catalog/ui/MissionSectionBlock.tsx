import type { Mission } from '../../../entities/mission'
import { PixelIcon, type PixelIconName } from '../../../shared/ui/PixelIcon'
import { MissionCard } from './MissionCard'

export interface MissionSectionBlockProps {
  icon: PixelIconName
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
  // Пустой блок не показывается: заголовок без карточек выглядит как ошибка.
  if (missions.length === 0) {
    return null
  }

  return (
    <section className='catalog-section'>
      <div className='catalog-section__heading'>
        <h2>
          <PixelIcon name={icon} size={20} />
          {title}
        </h2>
        <button type='button' onClick={onViewAll}>
          Смотреть все
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
