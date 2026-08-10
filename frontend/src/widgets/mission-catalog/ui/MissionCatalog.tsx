import type { Mission, MissionSection, MissionStatus } from '../../../entities/mission'
import { missionStatus, type MissionSort } from '../../../features/filter-missions'
import { useToastStore } from '../../../features/toast'
import type { PixelIconName } from '../../../shared/ui/PixelIcon'
import { MissionCatalogToolbar } from './MissionCatalogToolbar'
import { MissionSectionBlock } from './MissionSectionBlock'

/** Блоки каталога в порядке показа на странице. */
const SECTIONS: Array<{
  section: MissionSection
  icon: PixelIconName
  title: string
  /** Фильтр, который открывает «Смотреть все» для этого блока. */
  viewAllFilter: MissionStatus | 'all'
}> = [
  { section: 'featured', icon: 'featured', title: 'РЕКОМЕНДУЕМЫЕ МИССИИ', viewAllFilter: 'all' },
  { section: 'new', icon: 'new', title: 'НОВЫЕ МИССИИ', viewAllFilter: 'all' },
  { section: 'progress', icon: 'in-progress', title: 'В ПРОЦЕССЕ', viewAllFilter: 'in-progress' },
  { section: 'archive', icon: 'locked', title: 'ЗАБЛОКИРОВАННЫЕ', viewAllFilter: 'locked' },
]

export interface MissionCatalogProps {
  visibleMissions: Mission[]
  isFiltering: boolean
  activeMission: string | null
  sort: MissionSort
  onSortChange: (sort: MissionSort) => void
  onSelectMission: (mission: Mission) => void
  onLaunchMission?: (mission: Mission) => void
  onToast?: (message: string) => void
  onClearFilters: () => void
  onSetProgressFilter: (progress: MissionStatus | 'all') => void
}

export function MissionCatalog({
  visibleMissions,
  isFiltering,
  activeMission,
  sort,
  onSortChange,
  onSelectMission,
  onLaunchMission,
  onToast,
  onClearFilters,
  onSetProgressFilter,
}: MissionCatalogProps) {
  const handleLocked = (mission: Mission) => {
    const message = mission.unlock ?? 'Эта миссия пока недоступна.'
    ;(onToast ?? useToastStore.getState().showToast)(message)
  }

  const missionsOf = (section: MissionSection) => {
    const inSection = visibleMissions.filter((mission) => mission.section === section)

    // Пройденную миссию не показываем в заблокированных: статус уже изменился.
    return section === 'archive'
      ? inSection.filter((mission) => missionStatus(mission, activeMission) !== 'completed')
      : inSection
  }

  return (
    <div className='mission-catalog'>
      <MissionCatalogToolbar
        totalFound={visibleMissions.length}
        sort={sort}
        onSortChange={onSortChange}
      />

      <div className='catalog-sections'>
        {isFiltering ? (
          <MissionSectionBlock
            icon='filter'
            title='РЕЗУЛЬТАТЫ ФИЛЬТРАЦИИ'
            missions={visibleMissions}
            activeMission={activeMission}
            onOpen={onSelectMission}
            onLaunch={onLaunchMission}
            onLocked={handleLocked}
            onViewAll={onClearFilters}
          />
        ) : (
          SECTIONS.map(({ section, icon, title, viewAllFilter }) => (
            <MissionSectionBlock
              key={section}
              icon={icon}
              title={title}
              missions={missionsOf(section)}
              activeMission={activeMission}
              onOpen={onSelectMission}
              onLaunch={onLaunchMission}
              onLocked={handleLocked}
              onViewAll={
                viewAllFilter === 'all' ? onClearFilters : () => onSetProgressFilter(viewAllFilter)
              }
            />
          ))
        )}

        {visibleMissions.length === 0 && (
          <div className='catalog-empty'>
            <strong>Миссии по выбранным фильтрам не найдены</strong>
            <p>Попробуйте выбрать другую роль или сложность.</p>
            <button type='button' onClick={onClearFilters}>
              Сбросить фильтры
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
