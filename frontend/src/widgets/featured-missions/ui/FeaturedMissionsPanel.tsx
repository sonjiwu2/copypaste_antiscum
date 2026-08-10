import { scenariosToMissions } from '../../../entities/mission'
import { useScenarios } from '../../../shared/api'
import { PixelIcon } from '../../../shared/ui/PixelIcon'

/** Сколько сценариев показывает панель на главной странице. */
const FEATURED_LIMIT = 5

export function FeaturedMissionsPanel({ onAction }: { onAction: () => void }) {
  const { scenarios } = useScenarios()
  const featured = scenariosToMissions(scenarios).slice(0, FEATURED_LIMIT)

  return (
    <section className='sidebar-panel featured-panel pixel-card' aria-labelledby='featured-title'>
      <div className='panel-heading'>
        <h2 id='featured-title'>
          <PixelIcon name='featured' size={22} className='heading-asset' /> РЕКОМЕНДУЕМЫЕ МИССИИ
        </h2>
        <button type='button' onClick={onAction}>
          Смотреть все
        </button>
      </div>
      <div className='mission-list'>
        {featured.map((mission) => (
          <button
            className='mission-row'
            key={mission.id}
            type='button'
            onClick={onAction}
          >
            <span className={`mission-tile mission-tile--${mission.tone}`}>
              <PixelIcon name={mission.icon} size={52} />
            </span>
            <span className='mission-copy'>
              <strong>{mission.title}</strong>
              <small>{mission.description}</small>
            </span>
            <span className='mission-xp'>+{mission.xp} XP</span>
          </button>
        ))}
      </div>
    </section>
  )
}
