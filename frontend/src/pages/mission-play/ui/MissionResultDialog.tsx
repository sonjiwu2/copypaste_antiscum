import type { AttemptNodeOutcome } from '../../../shared/api'
import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { RISK_LABELS, type RiskGrade } from '../lib/missionPlayHelpers'

export interface MissionResultDialogProps {
  outcome?: AttemptNodeOutcome
  xpEarned: number
  score: number
  riskGrade: RiskGrade
  onRestart: () => void
  onExit: () => void
}

/**
 * Итог сценария.
 *
 * Неудача объясняет ошибку и предлагает пройти заново: экран учит, а не
 * наказывает, поэтому оформление здесь спокойнее, чем красная рамка.
 */
export function MissionResultDialog({
  outcome,
  xpEarned,
  score,
  riskGrade,
  onRestart,
  onExit,
}: MissionResultDialogProps) {
  const isSafe = outcome?.type === 'safe'

  return (
    <div className='mission-complete-backdrop'>
      <section
        className={`mission-complete mission-complete--${isSafe ? 'safe' : 'unsafe'}`}
        role='dialog'
        aria-modal='true'
        aria-labelledby='mission-complete-title'
      >
        <PixelIcon
          name={isSafe ? 'trophy' : 'shield'}
          size={72}
          className='mission-complete__art'
        />
        <h1 id='mission-complete-title'>
          {isSafe ? 'Сделка прошла безопасно' : 'Сценарий завершён'}
        </h1>
        <p>{outcome?.title ?? ''}</p>
        <span>{outcome?.explanation ?? ''}</span>
        <div className='mission-complete__stats'>
          <div>
            <small>Опыт</small>
            <strong>+{xpEarned}</strong>
          </div>
          <div>
            <small>Безопасные решения</small>
            <strong>{score}%</strong>
          </div>
          <div>
            <small>Уровень риска</small>
            <strong>{RISK_LABELS[riskGrade]}</strong>
          </div>
        </div>
        <div className='mission-complete__actions'>
          <button type='button' onClick={onRestart}>
            Пройти снова
          </button>
          <button type='button' onClick={onExit}>
            К списку миссий
          </button>
        </div>
      </section>
    </div>
  )
}
