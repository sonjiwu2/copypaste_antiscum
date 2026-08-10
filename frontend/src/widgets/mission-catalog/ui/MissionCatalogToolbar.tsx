import type { MissionSort } from '../../../features/filter-missions'

const SORT_OPTIONS: Array<[MissionSort, string]> = [
  ['recommended', 'Рекомендуемые'],
  ['difficulty', 'По сложности'],
  ['xp', 'По награде'],
  ['title', 'По названию'],
]

/** Правильная форма слова «миссия» для числа найденных карточек. */
function missionsFoundLabel(total: number): string {
  const lastTwo = total % 100
  const last = total % 10

  if (lastTwo >= 11 && lastTwo <= 14) {
    return 'миссий'
  }

  if (last === 1) {
    return 'миссия'
  }

  return last >= 2 && last <= 4 ? 'миссии' : 'миссий'
}

export interface MissionCatalogToolbarProps {
  totalFound: number
  sort: MissionSort
  onSortChange: (sort: MissionSort) => void
}

export function MissionCatalogToolbar({
  totalFound,
  sort,
  onSortChange,
}: MissionCatalogToolbarProps) {
  return (
    <div className='catalog-toolbar'>
      <span className='catalog-toolbar__count'>
        Найдено: {totalFound} {missionsFoundLabel(totalFound)}
      </span>
      <label>
        Сортировка:
        <select value={sort} onChange={(event) => onSortChange(event.target.value as MissionSort)}>
          {SORT_OPTIONS.map(([value, label]) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </select>
      </label>
    </div>
  )
}
