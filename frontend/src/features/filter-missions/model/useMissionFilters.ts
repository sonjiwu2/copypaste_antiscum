import { useState, useEffect, useMemo, useCallback, Dispatch, SetStateAction } from 'react'
import { Mission, Difficulty, MissionStatus } from '../../../entities/mission'
import { useUserProgressStore } from '../../../entities/user-progress'
import { useToastStore } from '../../toast'

export type RoleFilter = 'buyer' | 'seller' | 'both'

export const CATEGORY_ORDER = [
  'Fake Listings',
  'Payment Scams',
  'Phishing & Links',
  'Account Takeover',
  'Shipping Scams',
  'Impersonation',
]

function readLocalStorageItem(key: string): string | null {
  try {
    return localStorage.getItem(key)
  } catch {
    return null
  }
}

export function missionStatus(mission: Mission, activeMission: string | null): MissionStatus {
  try {
    const userCompleted = useUserProgressStore.getState().completedMissions
    if (userCompleted[mission.id]) return 'completed'
    const raw = readLocalStorageItem('antiscam.completedMissions')
    const completed = JSON.parse(raw ?? '[]') as string[]
    if (completed.includes(mission.id)) return 'completed'
  } catch {
    /* demo storage is optional */
  }
  if (activeMission === mission.id && mission.status !== 'locked' && mission.status !== 'completed')
    return 'in-progress'
  return mission.status
}

export interface UseMissionFiltersOptions {
  allMissions: Mission[]
  initialRole?: RoleFilter
  initialActiveMission?: string | null
  searchRoleParam?: RoleFilter | null
  onRoleParamChange?: (newRole: RoleFilter) => void
  onToast?: (message: string) => void
  onLaunch?: (missionId: string) => void
}

export interface UseMissionFiltersReturn {
  role: RoleFilter
  setRole: (newRole: RoleFilter) => void
  difficulty: Difficulty | 'all'
  setDifficulty: Dispatch<SetStateAction<Difficulty | 'all'>>
  categories: string[]
  toggleCategory: (category: string) => void
  progress: MissionStatus | 'all'
  setProgress: Dispatch<SetStateAction<MissionStatus | 'all'>>
  sort: string
  setSort: Dispatch<SetStateAction<string>>
  viewMore: boolean
  setViewMore: Dispatch<SetStateAction<boolean>>
  selectedMission: Mission | null
  setSelectedMission: Dispatch<SetStateAction<Mission | null>>
  activeMission: string | null
  setActiveMission: Dispatch<SetStateAction<string | null>>
  categoryCounts: Record<string, number>
  progressCounts: Record<string, number>
  visibleMissions: Mission[]
  isFiltering: boolean
  clearFilters: () => void
  beginMission: (targetMission?: Mission | null | unknown) => void
}

export function useMissionFilters({
  allMissions,
  initialRole = 'both',
  initialActiveMission,
  searchRoleParam,
  onRoleParamChange,
  onToast,
  onLaunch,
}: UseMissionFiltersOptions): UseMissionFiltersReturn {
  const [role, setRoleState] = useState<RoleFilter>(searchRoleParam || initialRole)
  const [difficulty, setDifficulty] = useState<Difficulty | 'all'>('all')
  const [categories, setCategories] = useState<string[]>([])
  const [progress, setProgress] = useState<MissionStatus | 'all'>('all')
  const [sort, setSort] = useState('recommended')
  const [viewMore, setViewMore] = useState(false)
  const [selectedMission, setSelectedMission] = useState<Mission | null>(null)
  const [activeMission, setActiveMission] = useState<string | null>(
    () => initialActiveMission ?? readLocalStorageItem('antiscam.activeMission')
  )

  useEffect(() => {
    if (searchRoleParam && ['buyer', 'seller', 'both'].includes(searchRoleParam)) {
      setRoleState(searchRoleParam)
    }
  }, [searchRoleParam])

  const setRole = useCallback(
    (newRole: RoleFilter) => {
      setRoleState(newRole)
      if (onRoleParamChange) onRoleParamChange(newRole)
    },
    [onRoleParamChange]
  )

  const clearFilters = useCallback(() => {
    setRoleState('both')
    if (onRoleParamChange) onRoleParamChange('both')
    setDifficulty('all')
    setCategories([])
    setProgress('all')
    setSort('recommended')
  }, [onRoleParamChange])

  const toggleCategory = useCallback((category: string) => {
    setCategories((current) =>
      current.includes(category)
        ? current.filter((item) => item !== category)
        : [...current, category]
    )
  }, [])

  const beginMission = useCallback(
    (targetMission?: Mission | null | unknown) => {
      const isMission = (obj: unknown): obj is Mission =>
        Boolean(
          obj && typeof obj === 'object' && 'id' in obj && typeof (obj as Mission).id === 'string'
        )

      const missionToStart = isMission(targetMission) ? targetMission : selectedMission
      if (!missionToStart) return
      const missionId = missionToStart.id
      setActiveMission(missionId)
      try {
        localStorage.setItem('antiscam.activeMission', missionId)
      } catch {
        /* local storage is optional */
      }
      if (onToast) {
        onToast(`${missionToStart.title} started — demo progress is saved locally.`)
      } else {
        useToastStore
          .getState()
          .showToast(`${missionToStart.title} started — demo progress is saved locally.`)
      }
      if (onLaunch) {
        onLaunch(missionId)
      }
      setSelectedMission(null)
    },
    [selectedMission, onToast, onLaunch]
  )

  const categoryCounts = useMemo(
    () =>
      Object.fromEntries(
        CATEGORY_ORDER.map((category) => [
          category,
          allMissions.filter((m) => m.category === category).length,
        ])
      ),
    [allMissions]
  )

  const progressCounts = useMemo(
    () => ({
      all: allMissions.length,
      'not-started': allMissions.filter((m) => missionStatus(m, activeMission) === 'not-started')
        .length,
      'in-progress': allMissions.filter((m) => missionStatus(m, activeMission) === 'in-progress')
        .length,
      completed: allMissions.filter((m) => missionStatus(m, activeMission) === 'completed').length,
      locked: allMissions.filter((m) => missionStatus(m, activeMission) === 'locked').length,
    }),
    [allMissions, activeMission]
  )

  const visibleMissions = useMemo(() => {
    const filtered = allMissions.filter((mission) => {
      const roleMatch =
        role === 'both' || mission.role === 'Both' || mission.role.toLowerCase() === role
      const difficultyMatch = difficulty === 'all' || mission.difficulty === difficulty
      const categoryMatch = !categories.length || categories.includes(mission.category)
      const progressMatch = progress === 'all' || missionStatus(mission, activeMission) === progress
      return roleMatch && difficultyMatch && categoryMatch && progressMatch
    })

    return [...filtered].sort((a, b) => {
      if (sort === 'xp') return b.xp - a.xp
      if (sort === 'difficulty')
        return (
          ['Easy', 'Medium', 'Hard'].indexOf(a.difficulty) -
          ['Easy', 'Medium', 'Hard'].indexOf(b.difficulty)
        )
      if (sort === 'title') return a.title.localeCompare(b.title)
      return allMissions.indexOf(a) - allMissions.indexOf(b)
    })
  }, [allMissions, activeMission, categories, difficulty, progress, role, sort])

  const isFiltering =
    role !== 'both' ||
    difficulty !== 'all' ||
    categories.length > 0 ||
    progress !== 'all' ||
    sort !== 'recommended'

  return useMemo(
    () => ({
      role,
      setRole,
      difficulty,
      setDifficulty,
      categories,
      toggleCategory,
      progress,
      setProgress,
      sort,
      setSort,
      viewMore,
      setViewMore,
      selectedMission,
      setSelectedMission,
      activeMission,
      setActiveMission,
      categoryCounts,
      progressCounts,
      visibleMissions,
      isFiltering,
      clearFilters,
      beginMission,
    }),
    [
      role,
      setRole,
      difficulty,
      categories,
      toggleCategory,
      progress,
      sort,
      viewMore,
      selectedMission,
      activeMission,
      categoryCounts,
      progressCounts,
      visibleMissions,
      isFiltering,
      clearFilters,
      beginMission,
    ]
  )
}
