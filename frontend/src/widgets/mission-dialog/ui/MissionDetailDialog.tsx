import { useEffect, useRef } from 'react'
import {
  DIFFICULTY_LABELS,
  MissionArtwork,
  ROLE_LABELS,
  type Mission,
} from '../../../entities/mission'
import { missionStatus } from '../../../features/filter-missions'
import { PixelIcon } from '../../../shared/ui/PixelIcon'

const ACTION_LABELS = {
  'not-started': 'Начать миссию',
  'in-progress': 'Продолжить миссию',
  completed: 'Пройти снова',
  locked: 'Миссия заблокирована',
} as const

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
  const closeRef = useRef<HTMLButtonElement>(null)

  // Escape закрывает окно, а фокус уходит внутрь: иначе клавиатура остаётся
  // на карточке за подложкой.
  useEffect(() => {
    if (!mission) {
      return
    }

    closeRef.current?.focus()

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [mission, onClose])

  if (!mission) {
    return null
  }

  const status = missionStatus(mission, activeMission)

  return (
    <div
      className='mission-dialog-backdrop'
      role='presentation'
      onClick={(event) => {
        if (event.target === event.currentTarget) {
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
          aria-label='Закрыть'
          ref={closeRef}
          onClick={onClose}
        >
          <PixelIcon name='close' size={16} />
        </button>
        <MissionArtwork mission={mission} size={54} />
        <h2 id='mission-dialog-title'>{mission.title}</h2>
        <p>{mission.description}</p>
        <div className='mission-dialog__details'>
          <span>{DIFFICULTY_LABELS[mission.difficulty]}</span>
          <span>{ROLE_LABELS[mission.role]}</span>
          <span>+{mission.xp} XP</span>
        </div>
        <div className='mission-dialog__actions'>
          <button type='button' onClick={onClose}>
            Не сейчас
          </button>
          <button
            type='button'
            data-testid='begin-btn'
            disabled={status === 'locked'}
            onClick={() => onBegin(mission)}
          >
            {ACTION_LABELS[status]}
          </button>
        </div>
      </section>
    </div>
  )
}
