import {
  DIFFICULTY_LABELS,
  MissionArtwork,
  type Mission,
} from '../../../entities/mission'
import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { RISK_HINTS, RISK_LABELS, type RiskGrade } from '../lib/missionPlayHelpers'

export interface MissionSidebarProps {
  mission: Mission
  step: number
  totalSteps: number
  progressPercent: number
  objective: string
  riskGrade: RiskGrade
  riskMarkerPercent: number
  completedCount: number
  achievementsCount: number
  streakDays: number
  onExit: () => void
}

/** Левая колонка экрана прохождения: где мы в сценарии и чем он рискует. */
export function MissionSidebar({
  mission,
  step,
  totalSteps,
  progressPercent,
  objective,
  riskGrade,
  riskMarkerPercent,
  completedCount,
  achievementsCount,
  streakDays,
  onExit,
}: MissionSidebarProps) {
  return (
    <aside className='play-sidebar play-sidebar--left'>
      <section className='play-panel mission-progress-card'>
        <div className='play-panel__heading'>
          <PixelIcon name='missions' size={18} />
          <h2>ХОД МИССИИ</h2>
        </div>
        <div className='mission-progress-card__mission'>
          <MissionArtwork mission={mission} size={38} />
          <div>
            <strong>{mission.title}</strong>
            <p>{mission.description}</p>
          </div>
          <small>{DIFFICULTY_LABELS[mission.difficulty]}</small>
        </div>
        <div className='play-progress-label'>
          <strong>
            ШАГ {step} ИЗ {totalSteps}
          </strong>
          <span>{progressPercent}%</span>
        </div>
        <div
          className='play-progress-track'
          role='progressbar'
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={progressPercent}
        >
          <span style={{ width: `${progressPercent}%` }} />
        </div>
      </section>

      <section className='play-panel objective-card'>
        <div className='play-panel__heading'>
          <PixelIcon name='star' size={18} />
          <h2>ТЕКУЩАЯ ЗАДАЧА</h2>
        </div>
        <p>{objective}</p>
      </section>

      <section className={`play-panel risk-card risk-card--${riskGrade}`}>
        <div className='risk-card__copy'>
          <h2>УРОВЕНЬ РИСКА</h2>
          <strong>{RISK_LABELS[riskGrade]}</strong>
          <p>{RISK_HINTS[riskGrade]}</p>
        </div>
        <div className='risk-scale' aria-label={RISK_LABELS[riskGrade]}>
          <span className='risk-scale__green' />
          <span className='risk-scale__orange' />
          <span className='risk-scale__yellow' />
          <span className='risk-scale__red' />
          <i style={{ left: `${riskMarkerPercent}%` }} />
        </div>
      </section>

      <section className='play-panel badge-card'>
        <div className='play-panel__heading'>
          <PixelIcon name='trophy' size={18} />
          <h2>НАГРАДЫ И УСПЕХИ</h2>
        </div>
        <div className='badge-card__items'>
          <div>
            <PixelIcon name='shield' size={26} />
            <span>{completedCount} миссий</span>
          </div>
          <div>
            <PixelIcon name='rewards' size={26} />
            <span>{achievementsCount} награды</span>
          </div>
          <div>
            <PixelIcon name='streak' size={26} />
            <span>серия {streakDays}</span>
          </div>
        </div>
      </section>

      <button className='exit-mission-button' type='button' onClick={onExit}>
        ← К списку миссий
      </button>
    </aside>
  )
}
