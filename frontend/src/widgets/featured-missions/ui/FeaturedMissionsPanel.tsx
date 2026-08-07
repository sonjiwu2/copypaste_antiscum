import { FEATURED_SUMMARY } from '../../../entities/mission'
import { useScenarios } from '../../../shared/api'
import { PROVIDED_ASSET_ROOT } from '../../../shared/config/assets'

const ASSETS = [
  'featured-message.png',
  'featured-payment.png',
  'featured-listing.png',
  'featured-security.png',
]

export function FeaturedMissionsPanel({ onAction }: { onAction: (title: string) => void }) {
  const { scenarios } = useScenarios()

  const featuredItems =
    scenarios && scenarios.length > 0
      ? scenarios.slice(0, 6).map((sc, i) => ({
          title: sc.title,
          subtitle: sc.description,
          asset: ASSETS[i % ASSETS.length],
        }))
      : FEATURED_SUMMARY

  return (
    <section className='sidebar-panel featured-panel pixel-card' aria-labelledby='featured-title'>
      <div className='panel-heading'>
        <h2 id='featured-title'>
          <img
            src={`${PROVIDED_ASSET_ROOT}/featured-flag.png`}
            alt=''
            className='heading-asset pixel-art'
          />{' '}
          РЕКОМЕНДУЕМЫЕ МИССИИ
        </h2>
        <button type='button' onClick={() => onAction('Все миссии')}>
          Смотреть все
        </button>
      </div>
      <div className='mission-list'>
        {featuredItems.map((mission) => (
          <button
            className='mission-row'
            key={mission.title}
            type='button'
            onClick={() => onAction(mission.title)}
          >
            <span className='mission-tile'>
              <img src={`${PROVIDED_ASSET_ROOT}/${mission.asset}`} alt='' className='pixel-art' />
            </span>
            <span className='mission-copy'>
              <strong>{mission.title}</strong>
              <small>{mission.subtitle}</small>
            </span>
            <span className='mission-xp'>+10 XP</span>
          </button>
        ))}
      </div>
    </section>
  )
}

