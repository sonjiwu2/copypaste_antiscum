import { useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'
import { Header } from '../widgets/header'
import { HomePage } from '../pages/home'
import { MissionsPage } from '../pages/missions'
import { ProgressPage } from '../pages/progress'
import { SettingsPage } from '../pages/settings'
import { RulesPage } from '../pages/rules'
import { WeeklyTestPage } from '../pages/weekly-test'
import { RewardsPage } from '../pages/rewards'
import { PlaceholderPage } from '../pages/placeholder'
import { useToast, Toast } from '../features/toast'
import { useBackgroundMusic, useButtonSound } from '../features/sound'
import { useAppNotifications } from '../features/notifications'
import { useSettingsStore } from '../entities/settings'
import { MissionPlayWrapper } from './MissionPlayWrapper'
import { useAppNavigation } from './model/useAppNavigation'

/**
 * Разделы со своей раскладкой. Главная страница помещается в экран целиком,
 * а остальные скроллятся, поэтому получают свой класс.
 */
const SCROLLABLE_ROUTES = [
  'missions',
  'play',
  'progress',
  'rules',
  'settings',
  'weekly-test',
  'rewards',
] as const

export function App() {
  const { toast, dismiss } = useToast()
  useButtonSound()
  useBackgroundMusic()
  useAppNotifications()

  const theme = useSettingsStore((state) => state.settings.theme)

  // Тема ставится на корневой элемент, а не на оболочку: фон страницы за
  // пределами приложения берётся из html и иначе остался бы светлым.
  // Первую установку делает inline-скрипт в index.html, здесь — переключение.
  useEffect(() => {
    document.documentElement.dataset.theme = theme

    const themeColor = document.querySelector('meta[name="theme-color"]')
    themeColor?.setAttribute('content', theme === 'dark' ? '#1c2630' : '#f8fbff')
  }, [theme])

  const { currentRoute, handleNavigate, handleLaunchMission, handleExitMission, handleExitToHome } =
    useAppNavigation()

  return (
    <div className={appShellClass(currentRoute)}>
      <Header onNavigate={handleNavigate} currentRoute={currentRoute} />
      <Routes>
        <Route path='/' element={<HomePage onNavigate={handleNavigate} />} />
        <Route path='/missions' element={<MissionsPage onLaunch={handleLaunchMission} />} />
        <Route
          path='/mission/:missionId'
          element={<MissionPlayWrapper onExit={handleExitMission} />}
        />
        <Route
          path='/progress'
          element={<ProgressPage onNavigate={handleNavigate} onLaunch={handleLaunchMission} />}
        />
        <Route path='/settings' element={<SettingsPage />} />
        <Route path='/rules' element={<RulesPage onNavigate={handleNavigate} />} />
        <Route path='/weekly-test' element={<WeeklyTestPage onBack={handleExitToHome} />} />
        <Route path='/rewards' element={<RewardsPage onBack={handleExitToHome} />} />
        <Route path='*' element={<PlaceholderPage route='progress' onBack={handleExitToHome} />} />
      </Routes>
      {toast && <Toast message={toast} onDismiss={dismiss} />}
    </div>
  )
}

/**
 * Класс оболочки: раскладка текущего раздела.
 * Через класс, а не inline-стили, потому что правила живут в CSS страниц.
 */
function appShellClass(
  currentRoute: ReturnType<typeof useAppNavigation>['currentRoute']
): string {
  const classes = ['app-shell']

  if (SCROLLABLE_ROUTES.some((route) => route === currentRoute)) {
    classes.push(`app-shell--${currentRoute}`)
  }

  return classes.join(' ')
}
