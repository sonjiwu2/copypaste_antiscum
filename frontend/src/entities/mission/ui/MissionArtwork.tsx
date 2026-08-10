import { PixelIcon } from '../../../shared/ui/PixelIcon'
import type { Mission } from '../model/types'

export interface MissionArtworkProps {
  mission: Mission
  size?: number
}

/** Плитка миссии: иконка темы сценария на подложке своего цвета. */
export function MissionArtwork({ mission, size = 42 }: MissionArtworkProps) {
  return (
    <span className={`catalog-art catalog-art--${mission.tone}`}>
      <PixelIcon name={mission.icon} size={size} />
    </span>
  )
}
