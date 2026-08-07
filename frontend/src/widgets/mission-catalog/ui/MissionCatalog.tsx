import type { Mission, MissionStatus } from '../../../entities/mission'
import { missionStatus } from '../../../features/filter-missions'
import { useToastStore } from '../../../features/toast'
import { MissionCatalogToolbar } from './MissionCatalogToolbar'
import { MissionSectionBlock } from './MissionSectionBlock'

export interface MissionCatalogProps {
  visibleMissions: Mission[]
  isFiltering: boolean
  activeMission: string | null
  sort: string
  onSortChange: (sort: string) => void
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
  const sectionMissions = (section: Mission['section']) =>
    visibleMissions.filter((mission) => mission.section === section)

  const handleLocked = (mission: Mission) => {
    if (onToast) {
      onToast(mission.unlock ?? 'Эта миссия пока недоступна.')
    } else {
      useToastStore.getState().showToast(mission.unlock ?? 'Эта миссия пока недоступна.')
    }
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
            icon='⌕'
            title='РЕЗУЛЬТАТЫ ФИЛЬТРАЦИИ'
            missions={visibleMissions}
            activeMission={activeMission}
            onOpen={onSelectMission}
            onLaunch={onLaunchMission}
            onLocked={handleLocked}
            onViewAll={onClearFilters}
          />
        ) : (
          <>
            <MissionSectionBlock
              icon='🔖'
              title='РЕКОМЕНДУЕМЫЕ МИССИИ'
              missions={sectionMissions('featured')}
              activeMission={activeMission}
              onOpen={onSelectMission}
              onLaunch={onLaunchMission}
              onLocked={handleLocked}
              onViewAll={onClearFilters}
            />
            <MissionSectionBlock
              icon='▰'
              title='НОВЫЕ МИССИИ'
              missions={sectionMissions('new')}
              activeMission={activeMission}
              onOpen={onSelectMission}
              onLaunch={onLaunchMission}
              onLocked={handleLocked}
              onViewAll={onClearFilters}
            />
            <div className='catalog-bottom-sections'>
              <MissionSectionBlock
                icon='◈'
                title='В ПРОЦЕССЕ'
                missions={sectionMissions('progress')}
                activeMission={activeMission}
                onOpen={onSelectMission}
                onLaunch={onLaunchMission}
                onLocked={handleLocked}
                onViewAll={() => onSetProgressFilter('in-progress')}
              />
              <MissionSectionBlock
                icon='🔒'
                title='ЗАБЛОКИРОВАННЫЕ МИССИИ'
                missions={sectionMissions('archive').filter(
                  (mission) => missionStatus(mission, activeMission) !== 'completed'
                )}
                activeMission={activeMission}
                onOpen={onSelectMission}
                onLaunch={onLaunchMission}
                onLocked={handleLocked}
                onViewAll={() => onSetProgressFilter('locked')}
              />
            </div>
          </>
        )}

        {!visibleMissions.length && (
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
