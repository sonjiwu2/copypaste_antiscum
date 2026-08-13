import { avatarUrl, useSettingsStore } from '../../../entities/settings'
import { useUserProgressStore } from '../../../entities/user-progress'
import { SCENE_ROOT } from '../../../shared/config/assets'
import { PixelIcon, type PixelIconName } from '../../../shared/ui/PixelIcon'
import { useHeaderProfileMenu } from '../model/useHeaderProfileMenu'

export type AppRoute =
  | 'home'
  | 'missions'
  | 'play'
  | 'progress'
  | 'rules'
  | 'settings'
  | 'weekly-test'
  | 'rewards'

/** `route` совпадает с адресом страницы, `command` — с ключом в useAppNavigation. */
const NAV_ITEMS: Array<{ route: AppRoute; command: string; label: string; icon: PixelIconName }> = [
  { route: 'missions', command: 'MISSIONS', label: 'МИССИИ', icon: 'missions' },
  { route: 'progress', command: 'PROGRESS', label: 'ПРОГРЕСС', icon: 'progress' },
  { route: 'rules', command: 'RULES', label: 'ПРАВИЛА', icon: 'rules' },
  { route: 'settings', command: 'SETTINGS', label: 'НАСТРОЙКИ', icon: 'settings' },
]

export function Header({
  onNavigate,
  currentRoute,
  onLogout,
  loggingOut = false,
}: {
  onNavigate: (label: string) => void
  currentRoute: AppRoute
  onLogout: () => void
  loggingOut?: boolean
}) {
  const { profileOpen, profileRef, toggleProfile, closeMenu } = useHeaderProfileMenu()
  const level = useUserProgressStore((state) => state.level)
  const totalXp = useUserProgressStore((state) => state.totalXp)
  const playerName = useSettingsStore((state) => state.settings.playerName)
  const avatar = useSettingsStore((state) => state.settings.avatar)

  // Во время прохождения подсвечен раздел миссий: экран сценария открыт из него.
  const activeRoute = currentRoute === 'play' ? 'missions' : currentRoute

  return (
    <header className='app-header'>
      <button
        className='brand'
        type='button'
        aria-label='Антискам-тренажёр — главная'
        onClick={() => onNavigate('HOME')}
      >
        <img src={`${SCENE_ROOT}/brand-shield.png`} alt='' className='brand__shield pixel-art' />
        <span className='brand__copy'>
          <strong>АНТИСКАМ</strong>
          <b>ТРЕНАЖЁР</b>
        </span>
      </button>

      <nav className='primary-nav' aria-label='Основная навигация'>
        {NAV_ITEMS.map((item) => {
          const isActive = activeRoute === item.route

          return (
            <button
              key={item.route}
              type='button'
              className={`nav-link${isActive ? ' nav-link--active' : ''}`}
              aria-current={isActive ? 'page' : undefined}
              onClick={() => onNavigate(item.command)}
            >
              <PixelIcon name={item.icon} size={28} />
              <span>{item.label}</span>
            </button>
          )
        })}
      </nav>

      <div className='header-actions'>
        <button
          className='energy'
          type='button'
          onClick={() => onNavigate('PROGRESS')}
          aria-label={`${totalXp} очков опыта`}
        >
          <PixelIcon name='xp' size={23} />
          <strong>{totalXp} XP</strong>
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
              <img src={avatarUrl(avatar)} alt='' className='pixel-art' />
            </span>
            <span className='profile__copy'>
              <strong>{playerName || 'Без имени'}</strong>
              <small>Уровень {level}</small>
            </span>
            <span className='profile__chevron' aria-hidden='true'>
              ⌄
            </span>
          </button>
          {profileOpen && (
            <div className='profile-menu' role='menu'>
              <button type='button' role='menuitem' onClick={() => onNavigate('PROGRESS')}>
                Мой прогресс
              </button>
              <button type='button' role='menuitem' onClick={() => onNavigate('SETTINGS')}>
                Настройки
              </button>
              <button type='button' role='menuitem' onClick={closeMenu}>
                Закрыть меню
              </button>
              <button
                type='button'
                role='menuitem'
                className='profile-menu__logout'
                disabled={loggingOut}
                onClick={onLogout}
              >
                {loggingOut ? 'Выходим…' : 'Выйти из аккаунта'}
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
