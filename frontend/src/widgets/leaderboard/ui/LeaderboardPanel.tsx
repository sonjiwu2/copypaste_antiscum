import { avatarUrl, type AvatarId } from '../../../entities/settings'
import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { useLeaderboardQuery } from '../../../shared/api'

export function LeaderboardPanel() {
  const leaderboard = useLeaderboardQuery()

  const leaders = leaderboard.data?.leaders ?? []
  const current = leaderboard.data?.current

  return (
    <section
      className='sidebar-panel leaderboard-panel pixel-card'
      aria-labelledby='leaderboard-title'
    >
      <div className='panel-heading'>
        <h2 id='leaderboard-title'>
          <PixelIcon name='leaders' size={22} /> ТАБЛИЦА ЛИДЕРОВ
        </h2>
        <span className='leaderboard-status leaderboard-status--live'>ОНЛАЙН</span>
      </div>
      {leaderboard.isLoading && <p className='leaderboard-message'>Загружаем рейтинг…</p>}
      {leaderboard.isError && <p className='leaderboard-message'>Рейтинг временно недоступен.</p>}
      {!leaderboard.isLoading && !leaderboard.isError && leaders.length === 0 && (
        <div className='leaderboard-placeholder'>
          <span className='leaderboard-placeholder__icon' aria-hidden='true'>
            <PixelIcon name='leaders' size={32} />
          </span>
          <div>
            <strong>Станьте первым лидером</strong>
            <p>Завершите миссию — результат сразу появится в общем рейтинге.</p>
          </div>
        </div>
      )}
      {leaders.length > 0 && (
        <div className='leader-list'>
          {leaders.map((leader) => (
            <div
              className={`leader-row${leader.currentPlayer ? ' leader-row--current' : ''}`}
              key={leader.rank}
            >
              <span className={`rank rank--${leader.rank}`}>{leader.rank}</span>
              <span className='leader-avatar'>
                <img src={avatarUrl(leader.avatar as AvatarId)} alt='' className='pixel-art' />
              </span>
              <span className='leader-row__copy'>
                <strong>{leader.displayName}</strong>
                <small>
                  {leader.completedScenarios} мисс. · {leader.averageScore}%
                </small>
              </span>
              <span>{leader.rating} RP</span>
            </div>
          ))}
        </div>
      )}
      {current && current.rank > leaders.length && (
        <div className='leaderboard-current'>
          Ваше место: <strong>#{current.rank}</strong> · {current.rating} RP
        </div>
      )}
    </section>
  )
}
