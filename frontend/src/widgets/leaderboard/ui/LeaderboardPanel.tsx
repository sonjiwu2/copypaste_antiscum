import { PixelIcon } from '../../../shared/ui/PixelIcon'

export function LeaderboardPanel() {
  return (
    <section
      className='sidebar-panel leaderboard-panel pixel-card'
      aria-labelledby='leaderboard-title'
    >
      <div className='panel-heading'>
        <h2 id='leaderboard-title'>
          <PixelIcon name='leaders' size={22} /> ТАБЛИЦА ЛИДЕРОВ
        </h2>
        <span className='leaderboard-status'>СКОРО</span>
      </div>
      <div className='leaderboard-placeholder'>
        <span className='leaderboard-placeholder__icon' aria-hidden='true'>
          <PixelIcon name='leaders' size={32} />
        </span>
        <div>
          <strong>Мультиплеер в разработке</strong>
          <p>Рейтинг появится после подключения аккаунтов и общего сервера игроков.</p>
        </div>
      </div>
    </section>
  )
}
