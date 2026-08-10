import type { Consequence } from '../../../shared/api'
import {
  DIFFICULTY_ICONS,
  type Difficulty,
} from '../../../entities/mission'
import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { PRACTICAL_TIPS, type WorldEvent } from '../lib/missionPlayHelpers'

/** Иконка разбора выбора. Степень опасности совпадает со шкалой сложности. */
const SEVERITY_ICON: Record<Consequence['severity'], Difficulty> = {
  safe: 'easy',
  warning: 'medium',
  dangerous: 'hard',
}

export interface InsightsSidebarProps {
  /** Факты обстановки по ходу сделки. В переписке их нет: это не чьи-то слова. */
  events: WorldEvent[]
  /** Разборы уже сделанных выборов. До выбора список пуст. */
  signals: Consequence[]
  elapsedText: string
  xpEarned: number
  score: number
  onShowTip: (message: string) => void
}

/** Правая колонка: что произошло, чему научил выбор и как идёт попытка. */
export function InsightsSidebar({
  events,
  signals,
  elapsedText,
  xpEarned,
  score,
  onShowTip,
}: InsightsSidebarProps) {
  return (
    <aside className='play-sidebar play-sidebar--right'>
      <section className='play-panel events-card'>
        <div className='play-panel__heading'>
          <PixelIcon name='events' size={18} />
          <h2>ЧТО ПРОИСХОДИТ</h2>
          <small>{events.length}</small>
        </div>
        {events.length === 0 ? (
          <p className='events-card__empty'>
            Здесь появятся факты сделки: экран приложения, банк, доставка.
          </p>
        ) : (
          events.map((event) => (
            <p className='event-row' key={event.id}>
              {event.text}
            </p>
          ))
        )}
      </section>

      <section className='play-panel signals-card'>
        <div className='play-panel__heading play-panel__heading--red'>
          <PixelIcon name='signals' size={18} />
          <h2>ЗАМЕЧЕННЫЕ ПРИЗНАКИ</h2>
          <small>{signals.length}</small>
        </div>
        {signals.length === 0 ? (
          <p className='signals-card__empty'>
            Здесь появится разбор после первого ответа в диалоге.
          </p>
        ) : (
          signals.map((signal) => (
            <div className={`signal-row signal-row--${signal.severity}`} key={signal.title}>
              <PixelIcon name={DIFFICULTY_ICONS[SEVERITY_ICON[signal.severity]]} size={18} />
              <span>
                <strong>{signal.title}</strong>
                <small>{signal.explanation}</small>
                {/* Правило «на будущее» — то, что должно остаться после разбора. */}
                {signal.realWorldRule && <em>{signal.realWorldRule}</em>}
              </span>
            </div>
          ))
        )}
      </section>

      <section className='play-panel tips-card'>
        <div className='play-panel__heading'>
          <PixelIcon name='tips' size={18} />
          <h2>ПРАКТИЧЕСКИЕ СОВЕТЫ</h2>
          <small>{PRACTICAL_TIPS.length}</small>
        </div>
        {PRACTICAL_TIPS.map((tip) => (
          <button
            className='tip-row'
            type='button'
            key={tip.title}
            onClick={() => onShowTip(`${tip.title}: ${tip.advice}`)}
          >
            <span>
              <strong>{tip.title}</strong>
              <small>{tip.advice}</small>
            </span>
          </button>
        ))}
      </section>

      <section className='play-panel session-card'>
        <div className='play-panel__heading'>
          <PixelIcon name='session' size={18} />
          <h2>ЭТА ПОПЫТКА</h2>
        </div>
        <div className='session-card__stats'>
          <div className='session-stat session-stat--time'>
            <span className='session-stat__copy'>
              <small>Время</small>
              <strong>{elapsedText}</strong>
            </span>
          </div>
          <div className='session-stat'>
            <PixelIcon name='xp' size={20} />
            <span className='session-stat__copy'>
              <small>Опыт</small>
              <strong>{xpEarned}</strong>
            </span>
          </div>
        </div>
        <div className='session-accuracy'>
          <span>Безопасные решения</span>
          <strong>{score}%</strong>
        </div>
      </section>
    </aside>
  )
}
