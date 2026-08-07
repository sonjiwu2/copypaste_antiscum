import { type Mission } from '../../../entities/mission'
import { PROVIDED } from '../lib/missionPlayHelpers'

export function MissionArtwork({ mission }: { mission: Mission }) {
  if (mission.asset)
    return <img src={`${PROVIDED}/${mission.asset}`} alt='' className='pixel-art' />
  return <span aria-hidden='true'>{mission.glyph ?? '⚑'}</span>
}
