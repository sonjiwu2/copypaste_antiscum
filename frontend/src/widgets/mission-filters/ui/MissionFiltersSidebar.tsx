import type { Dispatch, SetStateAction } from 'react'
import {
  CATEGORY_ICONS,
  CATEGORY_LABELS,
  CATEGORY_ORDER,
  DIFFICULTY_ICONS,
  DIFFICULTY_LABELS,
  DIFFICULTY_ORDER,
  ROLE_ICONS,
  ROLE_LABELS,
  STATUS_LABELS,
  type Difficulty,
  type MissionRole,
  type MissionStatus,
  type ScamCategory,
} from '../../../entities/mission'
import { PixelIcon } from '../../../shared/ui/PixelIcon'

/** Сколько типов скама видно до нажатия «Показать больше». */
const COLLAPSED_CATEGORIES = 5

const ROLES: MissionRole[] = ['buyer', 'seller', 'both']

const PROGRESS_OPTIONS: Array<MissionStatus | 'all'> = [
  'all',
  'not-started',
  'in-progress',
  'completed',
  'locked',
]

function progressLabel(value: MissionStatus | 'all'): string {
  return value === 'all' ? 'Все миссии' : STATUS_LABELS[value]
}

export interface MissionFiltersSidebarProps {
  role: MissionRole
  setRole: (role: MissionRole) => void
  difficulty: Difficulty | 'all'
  setDifficulty: Dispatch<SetStateAction<Difficulty | 'all'>>
  categories: ScamCategory[]
  toggleCategory: (category: ScamCategory) => void
  progress: MissionStatus | 'all'
  setProgress: (progress: MissionStatus | 'all') => void
  viewMore: boolean
  setViewMore: Dispatch<SetStateAction<boolean>>
  categoryCounts: Record<ScamCategory, number>
  progressCounts: Record<MissionStatus | 'all', number>
  clearFilters: () => void
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
  const visibleCategories = viewMore
    ? CATEGORY_ORDER
    : CATEGORY_ORDER.slice(0, COLLAPSED_CATEGORIES)

  return (
    <aside className='mission-filters' aria-label='Фильтры миссий'>
      <div className='filter-title'>
        <h2>
          <PixelIcon name='filter' size={18} />
          ФИЛЬТРЫ
        </h2>
        <button type='button' onClick={clearFilters}>
          Сбросить всё
        </button>
      </div>

      <fieldset className='filter-group'>
        <legend>РОЛЬ</legend>
        <div className='role-filter-row'>
          {ROLES.map((value) => (
            <button
              key={value}
              type='button'
              className={role === value ? 'is-active' : ''}
              aria-pressed={role === value}
              onClick={() => setRole(value)}
            >
              <PixelIcon name={ROLE_ICONS[value]} size={20} />
              {ROLE_LABELS[value]}
            </button>
          ))}
        </div>
      </fieldset>

      <fieldset className='filter-group'>
        <legend>СЛОЖНОСТЬ</legend>
        <div className='difficulty-filter-row'>
          {DIFFICULTY_ORDER.map((value) => (
            <button
              key={value}
              type='button'
              className={difficulty === value ? `is-active is-${value}` : ''}
              aria-pressed={difficulty === value}
              // Повторное нажатие снимает фильтр: это переключатель, а не радио.
              onClick={() => setDifficulty((current) => (current === value ? 'all' : value))}
            >
              <PixelIcon name={DIFFICULTY_ICONS[value]} size={18} />
              {DIFFICULTY_LABELS[value]}
            </button>
          ))}
        </div>
      </fieldset>

      <fieldset className='filter-group filter-group--list'>
        <legend>ТИП СКАМА</legend>
        {visibleCategories.map((category) => (
          <label key={category}>
            <input
              type='checkbox'
              checked={categories.includes(category)}
              onChange={() => toggleCategory(category)}
            />
            <PixelIcon name={CATEGORY_ICONS[category]} size={18} />
            <span>{CATEGORY_LABELS[category]}</span>
            <small>{categoryCounts[category]}</small>
          </label>
        ))}
        {CATEGORY_ORDER.length > COLLAPSED_CATEGORIES && (
          <button
            className='view-more-filter'
            type='button'
            aria-expanded={viewMore}
            onClick={() => setViewMore((value) => !value)}
          >
            {viewMore ? 'Показать меньше' : 'Показать больше'}
          </button>
        )}
      </fieldset>

      <fieldset className='filter-group filter-group--list progress-filter'>
        <legend>ПРОГРЕСС</legend>
        {PROGRESS_OPTIONS.map((value) => (
          <label key={value} className={progress === value ? 'is-active' : ''}>
            <input
              type='radio'
              name='progress'
              checked={progress === value}
              onChange={() => setProgress(value)}
            />
            <span>{progressLabel(value)}</span>
            <small>{progressCounts[value]}</small>
          </label>
        ))}
      </fieldset>

      <button className='reset-filters' type='button' onClick={clearFilters}>
        <PixelIcon name='reset' size={16} />
        Сбросить фильтры
      </button>
    </aside>
  )
}
