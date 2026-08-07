import type { Difficulty, MissionStatus } from '../../../entities/mission'
import { PROVIDED_ASSET_ROOT as ASSET_ROOT } from '../../../shared/config/assets'
import { CATEGORY_ORDER, type RoleFilter } from '../../../features/filter-missions'

export interface MissionFiltersSidebarProps {
  role: RoleFilter
  setRole: (role: RoleFilter) => void
  difficulty: Difficulty | 'all'
  setDifficulty: React.Dispatch<React.SetStateAction<Difficulty | 'all'>>
  categories: string[]
  toggleCategory: (category: string) => void
  progress: MissionStatus | 'all'
  setProgress: (progress: MissionStatus | 'all') => void
  viewMore: boolean
  setViewMore: React.Dispatch<React.SetStateAction<boolean>>
  categoryCounts: Record<string, number>
  progressCounts: Record<MissionStatus | 'all', number>
  clearFilters: () => void
}

const ROLES: RoleFilter[] = ['buyer', 'seller', 'both']
const DIFFICULTIES: Difficulty[] = ['Easy', 'Medium', 'Hard']
const PROGRESS_OPTIONS: Array<[MissionStatus | 'all', string]> = [
  ['all', 'All Missions'],
  ['not-started', 'Not Started'],
  ['in-progress', 'In Progress'],
  ['completed', 'Completed'],
  ['locked', 'Locked'],
]

function getCategoryIcon(category: string): string {
  switch (category) {
    case 'Fake Listings':
      return '▣'
    case 'Payment Scams':
      return '◆'
    case 'Phishing & Links':
      return '↝'
    case 'Account Takeover':
      return '⬡'
    default:
      return '▰'
  }
}

export function MissionFiltersSidebar({
  role,
  setRole,
  difficulty,
  setDifficulty,
  categories,
  toggleCategory,
  progress,
  setProgress,
  viewMore,
  setViewMore,
  categoryCounts,
  progressCounts,
  clearFilters,
}: MissionFiltersSidebarProps) {
  const visibleCategories = CATEGORY_ORDER.slice(0, viewMore ? CATEGORY_ORDER.length : 5)

  return (
    <aside className='mission-filters' aria-label='Mission filters'>
      <div className='filter-title'>
        <h2>
          <span aria-hidden='true'>▽</span> FILTER MISSIONS
        </h2>
        <button type='button' onClick={clearFilters}>
          Clear all
        </button>
      </div>

      <fieldset className='filter-group'>
        <legend>ROLE</legend>
        <div className='role-filter-row'>
          {ROLES.map((value) => (
            <button
              key={value}
              type='button'
              className={role === value ? 'is-active' : ''}
              aria-pressed={role === value}
              onClick={() => setRole(value)}
            >
              <img
                src={
                  value === 'buyer'
                    ? `${ASSET_ROOT}/buyer-role.png`
                    : value === 'seller'
                      ? `${ASSET_ROOT}/seller-role.png`
                      : '/assets/anti-scam/brand-shield.png'
                }
                alt=''
                className='pixel-art'
              />
              {value[0].toUpperCase() + value.slice(1)}
            </button>
          ))}
        </div>
      </fieldset>

      <fieldset className='filter-group'>
        <legend>DIFFICULTY</legend>
        <div className='difficulty-filter-row'>
          {DIFFICULTIES.map((value) => (
            <button
              key={value}
              type='button'
              className={difficulty === value ? `is-active is-${value.toLowerCase()}` : ''}
              aria-pressed={difficulty === value}
              onClick={() => setDifficulty((current) => (current === value ? 'all' : value))}
            >
              <span aria-hidden='true'>
                {value === 'Easy' ? '🙂' : value === 'Medium' ? '😐' : '😡'}
              </span>
              {value}
            </button>
          ))}
        </div>
      </fieldset>

      <fieldset className='filter-group filter-group--list'>
        <legend>SCAM TYPE</legend>
        {visibleCategories.map((category) => (
          <label key={category}>
            <input
              type='checkbox'
              checked={categories.includes(category)}
              onChange={() => toggleCategory(category)}
            />
            <span className='filter-category-icon' aria-hidden='true'>
              {getCategoryIcon(category)}
            </span>
            <span>{category}</span>
            <small>{categoryCounts[category]}</small>
          </label>
        ))}
        <button
          className='view-more-filter'
          type='button'
          onClick={() => setViewMore((value) => !value)}
        >
          {viewMore ? 'View less' : 'View more'}⌄
        </button>
      </fieldset>

      <fieldset className='filter-group filter-group--list progress-filter'>
        <legend>PROGRESS</legend>
        {PROGRESS_OPTIONS.map(([value, label]) => (
          <label key={value} className={progress === value ? 'is-active' : ''}>
            <input
              type='radio'
              name='progress'
              checked={progress === value}
              onChange={() => setProgress(value)}
            />
            <span>{label}</span>
            <small>{progressCounts[value]}</small>
          </label>
        ))}
      </fieldset>

      <button className='reset-filters' type='button' onClick={clearFilters}>
        ↻ Reset Filters
      </button>
    </aside>
  )
}
