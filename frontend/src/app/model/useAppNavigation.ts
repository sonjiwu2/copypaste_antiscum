import { useCallback, useMemo } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { AppRoute } from '../../widgets/header'
import { useToastStore } from '../../features/toast'

export function currentRouteFromPath(pathname: string): AppRoute {
  const path = pathname.substring(1)
  if (path.startsWith('mission/')) return 'play'
  if (path === 'missions' || path === 'progress' || path === 'rules' || path === 'settings')
    return path as AppRoute
  return 'home'
}

export function useAppNavigation(onToast?: (msg: string) => void) {
  const navigate = useNavigate()
  const location = useLocation()

  const currentRoute = useMemo(() => currentRouteFromPath(location.pathname), [location.pathname])

  const handleNavigate = useCallback(
    (label: string, roleOverride?: 'buyer' | 'seller' | 'both') => {
      const routeMap: Record<string, string> = {
        HOME: '/',
        MISSIONS:
          roleOverride && roleOverride !== 'both' ? `/missions?role=${roleOverride}` : '/missions',
        PROGRESS: '/progress',
        RULES: '/rules',
        SETTINGS: '/settings',
        PROFILE: '/progress',
      }

      const target = routeMap[label]
      if (target) {
        navigate(target)
        window.scrollTo({ top: 0, behavior: 'smooth' })
      } else {
        if (onToast) {
          onToast(`${label} selected.`)
        } else {
          useToastStore.getState().showToast(`${label} selected.`)
        }
      }
    },
    [navigate, onToast]
  )

  const handleLaunchMission = useCallback(
    (missionId: string) => {
      navigate(`/mission/${encodeURIComponent(missionId)}`)
      window.scrollTo({ top: 0 })
    },
    [navigate]
  )

  const handleExitMission = useCallback(() => {
    navigate('/missions')
  }, [navigate])

  const handleExitToHome = useCallback(() => {
    navigate('/')
  }, [navigate])

  return useMemo(
    () => ({
      currentRoute,
      handleNavigate,
      handleLaunchMission,
      handleExitMission,
      handleExitToHome,
    }),
    [currentRoute, handleNavigate, handleLaunchMission, handleExitMission, handleExitToHome]
  )
}
