export interface MissionCatalogToolbarProps {
  totalFound: number
  sort: string
  onSortChange: (sort: string) => void
}

export function MissionCatalogToolbar({
  totalFound,
  sort,
  onSortChange,
}: MissionCatalogToolbarProps) {
  return (
    <div className='catalog-toolbar'>
      <span>{totalFound} missions found</span>
      <label>
        Sort by:
        <select value={sort} onChange={(event) => onSortChange(event.target.value)}>
          <option value='recommended'>Recommended</option>
          <option value='difficulty'>Difficulty</option>
          <option value='xp'>XP reward</option>
          <option value='title'>Title</option>
        </select>
      </label>
    </div>
  )
}
