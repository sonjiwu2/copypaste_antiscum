import { MissionArtwork, type Mission } from '../../../entities/mission'
import { missionStatus } from '../../../features/filter-missions'

export interface MissionDetailDialogProps {
  mission: Mission | null
  activeMission: string | null
  onClose: () => void
  onBegin: (mission?: Mission | null) => void
}

export function MissionDetailDialog({
  mission,
  activeMission,
  onClose,
  onBegin,
}: MissionDetailDialogProps) {
  if (!mission) return null

  const currentStatus = missionStatus(mission, activeMission)
  const actionButtonText =
    currentStatus === 'completed'
      ? 'Review Mission'
      : currentStatus === 'in-progress'
        ? 'Continue Mission'
        : 'Begin Mission'

  return (
    <div
      className='mission-dialog-backdrop'
      role='presentation'
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          onClose()
        }
      }}
    >
      <section
        className='mission-dialog'
        role='dialog'
        aria-modal='true'
        aria-labelledby='mission-dialog-title'
      >
        <button
          className='mission-dialog__close'
          type='button'
          aria-label='Close mission'
          onClick={onClose}
        >
          ×
        </button>
        <MissionArtwork mission={mission} />
        <h2 id='mission-dialog-title'>{mission.title}</h2>
        <p>{mission.description}</p>
        <div className='mission-dialog__details'>
          <span>{mission.difficulty}</span>
          <span>{mission.role}</span>
          <span>+{mission.xp} XP</span>
        </div>
        <div className='mission-dialog__actions'>
          <button type='button' onClick={onClose}>
            Not now
          </button>
          <button type='button' data-testid='begin-btn' onClick={() => onBegin(mission)}>
            {actionButtonText}
          </button>
        </div>
      </section>
    </div>
  )
}
