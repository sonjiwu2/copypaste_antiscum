import type { Mission } from '../model/types'
import { PROVIDED_ASSET_ROOT as ASSET_ROOT } from '../../../shared/config/assets'

export function MissionArtwork({ mission }: { mission: Mission }) {
  return (
    <span className={`catalog-art catalog-art--${mission.tone}`} aria-hidden='true'>
      {mission.asset ? (
        <img src={`${ASSET_ROOT}/${mission.asset}`} alt='' className='pixel-art' />
      ) : (
        <span>{mission.glyph}</span>
      )}
    </span>
  )
}
