import { useCallback, useMemo } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import type { AppRoute } from '../../widgets/header'

/** Адрес раздела по команде из интерфейса. */
const ROUTE_PATHS: Record<string, string> = {
  HOME: '/',
  MISSIONS: '/missions',
  PROGRESS: '/progress',
  RULES: '/rules',
  SETTINGS: '/settings',
  WEEKLY_TEST: '/weekly-test',
  REWARDS: '/rewards',
}

export function currentRouteFromPath(pathname: string): AppRoute {
  const path = pathname.slice(1)

  if (path.startsWith('mission/')) {
    return 'play'
  }

  if (
    path === 'missions' ||
    path === 'progress' ||
    path === 'rules' ||
    path === 'settings' ||
    path === 'weekly-test' ||
    path === 'rewards'
  ) {
    return path
  }

  return 'home'
}

export function useAppNavigation() {
  const navigate = useNavigate()
  const location = useLocation()

  const currentRoute = useMemo(() => currentRouteFromPath(location.pathname), [location.pathname])

  const handleNavigate = useCallback(
    (command: string, role?: 'buyer' | 'seller' | 'both') => {
      const path = ROUTE_PATHS[command]
      if (!path) {
        return
      }

      // Роль попадает в адрес, чтобы каталог открывался уже отфильтрованным.
      const target =
        command === 'MISSIONS' && role && role !== 'both' ? `${path}?role=${role}` : path

      navigate(target)
      window.scrollTo({ top: 0, behavior: 'smooth' })
    },
    [navigate]
  )

  const handleLaunchMission = useCallback(
    (missionId: string) => {
      navigate(`/mission/${encodeURIComponent(missionId)}`)
      window.scrollTo({ top: 0 })
    },
    [navigate]
  )

  const handleExitMission = useCallback(() => navigate('/missions'), [navigate])
  const handleExitToHome = useCallback(() => navigate('/'), [navigate])

  return {
    currentRoute,
    handleNavigate,
    handleLaunchMission,
    handleExitMission,
    handleExitToHome,
  }
}
