import { Routes, Route } from 'react-router-dom'
import { Header } from '../widgets/header'
import { HomePage } from '../pages/home'
import { MissionsPage } from '../pages/missions'
import { ProgressPage } from '../pages/progress'
import { PlaceholderPage } from '../pages/placeholder'
import { useToast, Toast } from '../features/toast'
import { useButtonSound } from '../features/sound'
import { MissionPlayWrapper } from './MissionPlayWrapper'
import { useAppNavigation } from './model/useAppNavigation'

export function App() {
  const { toast, dismiss } = useToast()
  useButtonSound()

  const { currentRoute, handleNavigate, handleLaunchMission, handleExitMission, handleExitToHome } =
    useAppNavigation()

  return (
    <div
      className={`app-shell${currentRoute === 'missions' ? ' app-shell--missions' : ''}${currentRoute === 'play' ? ' app-shell--play' : ''}${currentRoute === 'progress' ? ' app-shell--progress' : ''}`}
    >
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
        <Route
          path='/rules'
          element={<PlaceholderPage route='rules' onBack={handleExitMission} />}
        />
        <Route
          path='/settings'
          element={<PlaceholderPage route='settings' onBack={handleExitMission} />}
        />
        <Route path='*' element={<PlaceholderPage route='progress' onBack={handleExitToHome} />} />
      </Routes>
      {toast && <Toast message={toast} onDismiss={dismiss} />}
    </div>
  )
}
