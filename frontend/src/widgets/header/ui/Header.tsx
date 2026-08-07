import { PixelIcon, IconName } from '../../../shared/ui/PixelIcon'
import { PROVIDED_ASSET_ROOT } from '../../../shared/config/assets'
import { useHeaderProfileMenu } from '../model/useHeaderProfileMenu'
import { useUserProgressStore } from '../../../entities/user-progress'

const NAV_ITEMS: Array<{ key: string; label: string; icon: IconName }> = [
  { key: 'MISSIONS', label: 'МИССИИ', icon: 'star' },
  { key: 'PROGRESS', label: 'ПРОГРЕСС', icon: 'chart' },
  { key: 'RULES', label: 'ПРАВИЛА', icon: 'book' },
  { key: 'SETTINGS', label: 'НАСТРОЙКИ', icon: 'gear' },
]

export type AppRoute = 'home' | 'missions' | 'play' | 'progress' | 'rules' | 'settings'

export function Header({
  onNavigate,
  currentRoute,
}: {
  onNavigate: (label: string) => void
  currentRoute: AppRoute
}) {
  const { profileOpen, profileRef, toggleProfile, closeMenu } = useHeaderProfileMenu()
  const level = useUserProgressStore((state) => state.level)
  const totalXp = useUserProgressStore((state) => state.totalXp)

  return (
    <header className='app-header'>
      <button
        className='brand'
        type='button'
        aria-label='Anti-Scam Trainer главная'
        onClick={() => onNavigate('HOME')}
      >
        <img src='/assets/anti-scam/brand-shield.png' alt='' className='brand__shield pixel-art' />
        <span className='brand__copy'>
          <strong>ANTI-SCAM</strong>
          <b>TRAINER</b>
        </span>
      </button>

      <nav className='primary-nav' aria-label='Основная навигация'>
        {NAV_ITEMS.map((item) => {
          const isActive =
            (currentRoute === 'play' ? 'missions' : currentRoute) === item.key.toLowerCase()
          return (
            <button
              key={item.key}
              type='button'
              className={`nav-link${isActive ? ' nav-link--active' : ''}`}
              aria-current={isActive ? 'page' : undefined}
              onClick={() => onNavigate(item.key)}
            >
              <PixelIcon name={item.icon} size={23} />
              <span>{item.label}</span>
            </button>
          )
        })}
      </nav>

      <div className='header-actions'>
        <button
          className='energy'
          type='button'
          onClick={() => onNavigate('ENERGY')}
          aria-label={`${totalXp} очков XP`}
        >
          <PixelIcon name='bolt' size={23} />
          <strong>{totalXp}</strong>
        </button>
        <div className='profile' ref={profileRef}>
          <button
            className='profile__button'
            type='button'
            aria-haspopup='menu'
            aria-expanded={profileOpen}
            onClick={toggleProfile}
          >
            <span className='profile__avatar'>
              <img src={`${PROVIDED_ASSET_ROOT}/profile-avatar.png`} alt='' className='pixel-art' />
            </span>
            <span className='profile__copy'>
              <strong>Алекс</strong>
              <small>Уровень {level}</small>
            </span>
            <span className='profile__chevron' aria-hidden='true'>
              ⌄
            </span>
          </button>
          {profileOpen && (
            <div className='profile-menu' role='menu'>
              <button type='button' role='menuitem' onClick={() => onNavigate('PROFILE')}>
                Просмотреть прогресс
              </button>
              <button type='button' role='menuitem' onClick={() => onNavigate('SETTINGS')}>
                Настройки
              </button>
              <button type='button' role='menuitem' onClick={closeMenu}>
                Закрыть меню
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}

