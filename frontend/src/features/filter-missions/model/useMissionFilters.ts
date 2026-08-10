import { useCallback, useMemo, useState, type Dispatch, type SetStateAction } from 'react'
import {
  CATEGORY_ORDER,
  type Difficulty,
  type Mission,
  type MissionRole,
  type MissionStatus,
  type ScamCategory,
} from '../../../entities/mission'
import { useUserProgressStore } from '../../../entities/user-progress'
import { useToastStore } from '../../toast'

/** Порядок сортировки каталога. */
export type MissionSort = 'recommended' | 'difficulty' | 'xp' | 'title'

const DEFAULT_SORT: MissionSort = 'recommended'

/** Ключ, по которому начатая миссия переживает перезагрузку страницы. */
const ACTIVE_MISSION_KEY = 'antiscam.activeMission'

const DIFFICULTY_WEIGHT: Record<Difficulty, number> = { easy: 0, medium: 1, hard: 2 }

function readLocalStorageItem(key: string): string | null {
  try {
    return localStorage.getItem(key)
  } catch {
    // Приватный режим браузера запрещает хранилище: фильтры просто начнут с нуля.
    return null
  }
}

function isMissionRole(value: string | null | undefined): value is MissionRole {
  return value === 'buyer' || value === 'seller' || value === 'both'
}

/**
 * Состояние миссии для каталога.
 *
 * Пройденные сценарии знает только store прогресса, поэтому статус из карточки
 * дополняется им, а активная миссия перекрывает статус «не начато».
 */
export function missionStatus(mission: Mission, activeMission: string | null): MissionStatus {
  if (useUserProgressStore.getState().completedMissions[mission.id]) {
    return 'completed'
  }

  if (mission.status === 'locked' || mission.status === 'completed') {
    return mission.status
  }

  return activeMission === mission.id ? 'in-progress' : mission.status
}

export interface UseMissionFiltersOptions {
  allMissions: Mission[]
  initialRole?: MissionRole
  initialActiveMission?: string | null
  /** Роль из адреса страницы. Если задана, она важнее выбора в сайдбаре. */
  searchRoleParam?: MissionRole | null
  onRoleParamChange?: (role: MissionRole) => void
  onToast?: (message: string) => void
  onLaunch?: (missionId: string) => void
}

export interface UseMissionFiltersReturn {
  role: MissionRole
  setRole: (role: MissionRole) => void
  difficulty: Difficulty | 'all'
  setDifficulty: Dispatch<SetStateAction<Difficulty | 'all'>>
  categories: ScamCategory[]
  toggleCategory: (category: ScamCategory) => void
  progress: MissionStatus | 'all'
  setProgress: Dispatch<SetStateAction<MissionStatus | 'all'>>
  sort: MissionSort
  setSort: Dispatch<SetStateAction<MissionSort>>
  viewMore: boolean
  setViewMore: Dispatch<SetStateAction<boolean>>
  selectedMission: Mission | null
  setSelectedMission: Dispatch<SetStateAction<Mission | null>>
  activeMission: string | null
  categoryCounts: Record<ScamCategory, number>
  progressCounts: Record<MissionStatus | 'all', number>
  visibleMissions: Mission[]
  isFiltering: boolean
  clearFilters: () => void
  beginMission: (mission?: Mission | null) => void
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
  const [selectedRole, setSelectedRole] = useState<MissionRole>(initialRole)
  const [difficulty, setDifficulty] = useState<Difficulty | 'all'>('all')
  const [categories, setCategories] = useState<ScamCategory[]>([])
  const [progress, setProgress] = useState<MissionStatus | 'all'>('all')
  const [sort, setSort] = useState<MissionSort>(DEFAULT_SORT)
  const [viewMore, setViewMore] = useState(false)
  const [selectedMission, setSelectedMission] = useState<Mission | null>(null)
  const [activeMission, setActiveMission] = useState<string | null>(
    () => initialActiveMission ?? readLocalStorageItem(ACTIVE_MISSION_KEY)
  )

  // Ссылка с ?role= задаёт фильтр: адрес страницы важнее последнего выбора,
  // иначе открытая ссылка показывала бы не то, что в ней написано.
  const role: MissionRole = isMissionRole(searchRoleParam) ? searchRoleParam : selectedRole

  const setRole = useCallback(
    (newRole: MissionRole) => {
      setSelectedRole(newRole)
      onRoleParamChange?.(newRole)
    },
    [onRoleParamChange]
  )

  const clearFilters = useCallback(() => {
    setSelectedRole('both')
    onRoleParamChange?.('both')
    setDifficulty('all')
    setCategories([])
    setProgress('all')
    setSort(DEFAULT_SORT)
  }, [onRoleParamChange])

  const toggleCategory = useCallback((category: ScamCategory) => {
    setCategories((current) =>
      current.includes(category)
        ? current.filter((item) => item !== category)
        : [...current, category]
    )
  }, [])

  const beginMission = useCallback(
    (mission?: Mission | null) => {
      const target = mission ?? selectedMission
      if (!target) {
        return
      }

      setActiveMission(target.id)
      try {
        localStorage.setItem(ACTIVE_MISSION_KEY, target.id)
      } catch {
        // Без хранилища миссия останется активной только до перезагрузки.
      }

      const notify = onToast ?? useToastStore.getState().showToast
      notify(`Миссия «${target.title}» начата.`)

      onLaunch?.(target.id)
      setSelectedMission(null)
    },
    [selectedMission, onToast, onLaunch]
  )

  const categoryCounts = useMemo(
    () =>
      Object.fromEntries(
        CATEGORY_ORDER.map((category) => [
          category,
          allMissions.filter((mission) => mission.category === category).length,
        ])
      ) as Record<ScamCategory, number>,
    [allMissions]
  )

  const progressCounts = useMemo(() => {
    const statuses = allMissions.map((mission) => missionStatus(mission, activeMission))
    const countOf = (status: MissionStatus) => statuses.filter((item) => item === status).length

    return {
      all: allMissions.length,
      'not-started': countOf('not-started'),
      'in-progress': countOf('in-progress'),
      completed: countOf('completed'),
      locked: countOf('locked'),
    }
  }, [allMissions, activeMission])

  const visibleMissions = useMemo(() => {
    const matching = allMissions.filter((mission) => {
      const roleMatch = role === 'both' || mission.role === 'both' || mission.role === role
      const difficultyMatch = difficulty === 'all' || mission.difficulty === difficulty
      const categoryMatch = categories.length === 0 || categories.includes(mission.category)
      const progressMatch = progress === 'all' || missionStatus(mission, activeMission) === progress

      return roleMatch && difficultyMatch && categoryMatch && progressMatch
    })

    // «Рекомендуемые» сохраняют порядок каталога, поэтому сортируется копия.
    return [...matching].sort((first, second) => {
      switch (sort) {
        case 'xp':
          return second.xp - first.xp
        case 'difficulty':
          return DIFFICULTY_WEIGHT[first.difficulty] - DIFFICULTY_WEIGHT[second.difficulty]
        case 'title':
          return first.title.localeCompare(second.title, 'ru')
        default:
          return 0
      }
    })
  }, [allMissions, activeMission, categories, difficulty, progress, role, sort])

  const isFiltering =
    role !== 'both' ||
    difficulty !== 'all' ||
    categories.length > 0 ||
    progress !== 'all' ||
    sort !== DEFAULT_SORT

  return {
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
    categoryCounts,
    progressCounts,
    visibleMissions,
    isFiltering,
    clearFilters,
    beginMission,
  }
}
