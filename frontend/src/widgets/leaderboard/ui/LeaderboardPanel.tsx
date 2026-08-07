import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { LEADERS } from '../../../entities/leader'
import { PROVIDED_ASSET_ROOT } from '../../../shared/config/assets'

export function LeaderboardPanel({ onAction }: { onAction: () => void }) {
  return (
    <section
      className='sidebar-panel leaderboard-panel pixel-card'
      aria-labelledby='leaderboard-title'
    >
      <div className='panel-heading'>
        <h2 id='leaderboard-title'>
          <PixelIcon name='users' size={22} /> ТАБЛИЦА ЛИДЕРОВ
        </h2>
        <button type='button' onClick={onAction}>
          Смотреть все
        </button>
      </div>
      <div className='leader-list'>
        {LEADERS.map((leader) => (
          <div className='leader-row' key={leader.name}>
            <span className={`rank rank--${leader.rank}`}>{leader.rank}</span>
            <span className='leader-avatar'>
              <img src={`${PROVIDED_ASSET_ROOT}/${leader.asset}`} alt='' className='pixel-art' />
            </span>
            <strong>{leader.name}</strong>
            <span>{leader.xp}</span>
          </div>
        ))}
      </div>
    </section>
  )
}
