import { useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import { scenariosToMissions, type MissionRole } from '../../../entities/mission'
import { useUserProgressStore } from '../../../entities/user-progress'
import { useScenarios } from '../../../shared/api'
import { useMissionFilters } from '../../../features/filter-missions'
import { MissionFiltersSidebar } from '../../../widgets/mission-filters'
import { MissionCatalog } from '../../../widgets/mission-catalog'
import { MissionDetailDialog } from '../../../widgets/mission-dialog'
import { CatalogError, CatalogLoading } from './CatalogState'
import './missions.css'

export interface MissionsPageProps {
  initialRole?: MissionRole
  activeMission?: string | null
  onToast?: (message: string) => void
  onLaunch?: (missionId: string) => void
}

function isMissionRole(value: string | null): value is MissionRole {
  return value === 'buyer' || value === 'seller' || value === 'both'
}

export function MissionsPage({
  initialRole = 'both',
  activeMission: activeMissionProp,
  onToast,
  onLaunch,
}: MissionsPageProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const roleParam = searchParams.get('role')

  const { scenarios, isLoading, isError, refetch } = useScenarios()
  const completedMissions = useUserProgressStore((state) => state.completedMissions)

  const allMissions = useMemo(
    () => scenariosToMissions(scenarios, new Set(Object.keys(completedMissions ?? {}))),
    [scenarios, completedMissions]
  )

  // Роль остаётся в адресе, чтобы ссылку на отфильтрованный каталог можно было
  // переслать. `both` — состояние по умолчанию, поэтому в адресе не хранится.
  const handleRoleParamChange = useCallback(
    (role: MissionRole) => {
      const next = new URLSearchParams(searchParams)

      if (role === 'both') {
        next.delete('role')
      } else {
        next.set('role', role)
      }

      setSearchParams(next, { replace: true })
    },
    [searchParams, setSearchParams]
  )

  const filters = useMissionFilters({
    allMissions,
    initialRole,
    initialActiveMission: activeMissionProp,
    searchRoleParam: isMissionRole(roleParam) ? roleParam : null,
    onRoleParamChange: handleRoleParamChange,
    onToast,
    onLaunch,
  })

  if (isLoading) {
    return <CatalogLoading />
  }

  if (isError) {
    return <CatalogError onRetry={() => void refetch()} />
  }

  return (
    <main className='missions-layout'>
      <MissionFiltersSidebar
        role={filters.role}
        setRole={filters.setRole}
        difficulty={filters.difficulty}
        setDifficulty={filters.setDifficulty}
        categories={filters.categories}
        toggleCategory={filters.toggleCategory}
        progress={filters.progress}
        setProgress={filters.setProgress}
        viewMore={filters.viewMore}
        setViewMore={filters.setViewMore}
        categoryCounts={filters.categoryCounts}
        progressCounts={filters.progressCounts}
        clearFilters={filters.clearFilters}
      />

      <MissionCatalog
        visibleMissions={filters.visibleMissions}
        isFiltering={filters.isFiltering}
        activeMission={filters.activeMission}
        sort={filters.sort}
        onSortChange={filters.setSort}
        onSelectMission={filters.setSelectedMission}
        onLaunchMission={filters.beginMission}
        onToast={onToast}
        onClearFilters={filters.clearFilters}
        onSetProgressFilter={filters.setProgress}
      />

      <MissionDetailDialog
        mission={filters.selectedMission}
        activeMission={filters.activeMission}
        onClose={() => filters.setSelectedMission(null)}
        onBegin={filters.beginMission}
      />
    </main>
  )
}
