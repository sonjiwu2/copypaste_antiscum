import { useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import { MISSIONS, scenariosToMissions } from '../../../entities/mission'
import { useUserProgressStore } from '../../../entities/user-progress'
import { useScenarios } from '../../../shared/api'
import { useMissionFilters, type RoleFilter } from '../../../features/filter-missions'
import { MissionFiltersSidebar } from '../../../widgets/mission-filters'
import { MissionCatalog } from '../../../widgets/mission-catalog'
import { MissionDetailDialog } from '../../../widgets/mission-dialog'
import '../../../app/missions.css'

export interface MissionsPageProps {
  initialRole?: RoleFilter
  activeMission?: string | null
  onToast?: (message: string) => void
  onLaunch?: (missionId: string) => void
}

export function MissionsPage({
  initialRole = 'both',
  activeMission: propActiveMission,
  onToast,
  onLaunch,
}: MissionsPageProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const searchRoleParam = searchParams.get('role') as RoleFilter | null

  // Получаем сценарии с бэкенда через TanStack Query
  const { scenarios, isLoading, isError, refetch } = useScenarios()
  const completedMissions = useUserProgressStore((state) => state.completedMissions)

  const allMissions = useMemo(() => {
    return scenariosToMissions(scenarios ?? [])
  }, [scenarios, completedMissions])

  const handleRoleParamChange = useCallback(
    (newRole: RoleFilter) => {
      if (newRole === 'both') {
        searchParams.delete('role')
      } else {
        searchParams.set('role', newRole)
      }
      setSearchParams(searchParams, { replace: true })
    },
    [searchParams, setSearchParams]
  )

  const {
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
  } = useMissionFilters({
    allMissions,
    initialRole,
    initialActiveMission: propActiveMission,
    searchRoleParam,
    onRoleParamChange: handleRoleParamChange,
    onToast,
    onLaunch,
  })


  if (isLoading) {
    return (
      <main className='missions-layout' style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <div style={{ color: '#93c5fd', fontSize: '1.2rem' }}>Загрузка сценариев с бэкенда...</div>
      </main>
    )
  }

  if (isError) {
    return (
      <main className='missions-layout' style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', minHeight: '60vh', gap: '1rem' }}>
        <div style={{ color: '#fca5a5', fontSize: '1.2rem' }}>Ошибка загрузки данных с бэкенда.</div>
        <button onClick={() => refetch()} style={{ padding: '0.5rem 1rem', background: '#3b82f6', color: '#fff', borderRadius: '0.375rem', cursor: 'pointer' }}>
          Повторить попытку
        </button>
      </main>
    )
  }

  return (
    <main className='missions-layout'>
      <MissionFiltersSidebar
        role={role}
        setRole={setRole}
        difficulty={difficulty}
        setDifficulty={setDifficulty}
        categories={categories}
        toggleCategory={toggleCategory}
        progress={progress}
        setProgress={setProgress}
        viewMore={viewMore}
        setViewMore={setViewMore}
        categoryCounts={categoryCounts}
        progressCounts={progressCounts}
        clearFilters={clearFilters}
      />

      <MissionCatalog
        visibleMissions={visibleMissions}
        isFiltering={isFiltering}
        activeMission={activeMission}
        sort={sort}
        onSortChange={setSort}
        onSelectMission={setSelectedMission}
        onLaunchMission={beginMission}
        onToast={onToast}
        onClearFilters={clearFilters}
        onSetProgressFilter={setProgress}
      />

      <MissionDetailDialog
        mission={selectedMission}
        activeMission={activeMission}
        onClose={() => setSelectedMission(null)}
        onBegin={beginMission}
      />
    </main>
  )
}
