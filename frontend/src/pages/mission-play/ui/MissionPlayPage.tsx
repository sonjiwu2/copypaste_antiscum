import { useMissionPlay } from '../model/useMissionPlay'
import { MissionArtwork } from './MissionArtwork'
import { DEMO_STEPS, TOTAL_STEPS, PRACTICAL_TIPS, PROVIDED } from '../lib/missionPlayHelpers'
import { useToastStore } from '../../../features/toast'
import { useUserProgressStore } from '../../../entities/user-progress'
import '../../../app/mission-play.css'

export function MissionPlayPage({
  missionId,
  onExit,
  onToast,
}: {
  missionId: string
  onExit: () => void
  onToast?: (message: string) => void
}) {
  const showToast = useToastStore((state) => state.showToast)
  const handleToast = onToast ?? showToast

  const completedCount = useUserProgressStore((state) => Object.keys(state.completedMissions ?? {}).length)
  const achievementsCount = useUserProgressStore((state) => (state.unlockedAchievements ?? []).length)
  const streakDays = useUserProgressStore((state) => state.streakDays ?? 1)

  const {
    mission,
    progress,
    completed,
    tipsOpen,
    setTipsOpen,
    messages,
    currentStep,
    progressPercent,
    accuracy,
    riskLabel,
    elapsedText,
    chooseResponse,
    restartMission,
  } = useMissionPlay(missionId, handleToast)

  return (
    <main className='mission-play-layout'>
      <aside className='play-sidebar play-sidebar--left'>
        <section className='play-panel mission-progress-card'>
          <div className='play-panel__heading'>
            <span>⚑</span>
            <h2>MISSION PROGRESS</h2>
          </div>
          <div className='mission-progress-card__mission'>
            <span className={`play-mission-art play-mission-art--${mission.tone}`}>
              <MissionArtwork mission={mission} />
            </span>
            <div>
              <strong>{mission.title}</strong>
              <p>{mission.description}</p>
            </div>
            <small>{mission.difficulty}</small>
          </div>
          <div className='play-progress-label'>
            <strong>
              STEP {Math.min(progress.step + 1, TOTAL_STEPS)} OF {TOTAL_STEPS}
            </strong>
            <span>{progressPercent}%</span>
          </div>
          <div className='play-progress-track'>
            <span style={{ width: `${progressPercent}%` }} />
          </div>
        </section>

        <section className='play-panel objective-card'>
          <div className='play-panel__heading'>
            <span>◉</span>
            <h2>CURRENT OBJECTIVE</h2>
          </div>
          <p>{currentStep.objective}</p>
        </section>

        <section
          className={`play-panel risk-card risk-card--${riskLabel.split(' ')[0].toLowerCase()}`}
        >
          <div className='risk-card__icon' aria-hidden='true'>
            ☠
          </div>
          <div>
            <h2>RISK METER</h2>
            <strong>{riskLabel}</strong>
            <p>
              {progress.risk <= 1
                ? 'You are keeping the conversation safe.'
                : progress.risk <= 4
                  ? 'The seller is using suspicious tactics.'
                  : 'Stop and report this conversation.'}
            </p>
          </div>
          <div className='risk-scale' aria-label={`${riskLabel}, ${progress.risk} of 8`}>
            <span className='risk-scale__green' />
            <span className='risk-scale__yellow' />
            <span className='risk-scale__orange' />
            <span className='risk-scale__red' />
            <i style={{ left: `${Math.min(96, (progress.risk / 8) * 100)}%` }} />
          </div>
        </section>

        <section className='play-panel badge-card'>
          <div className='play-panel__heading'>
            <span>♧</span>
            <h2>НАГРАДЫ И УСПЕХИ</h2>
          </div>
          <div className='badge-card__items'>
            <div>
              <b>🛡</b>
              <span>{completedCount} МИССИЙ</span>
            </div>
            <div>
              <b>🏆</b>
              <span>{achievementsCount} НАГРАД</span>
            </div>
            <div>
              <b>🔥</b>
              <span>{streakDays} D</span>
            </div>
          </div>
        </section>

        <button
          className='view-tips-button'
          type='button'
          onClick={() => setTipsOpen((value) => !value)}
        >
          ☼ {tipsOpen ? 'HIDE TIPS' : 'VIEW TIPS'} ({PRACTICAL_TIPS.length})
        </button>
        <button className='exit-mission-button' type='button' onClick={onExit}>
          ← Back to Mission Hub
        </button>
      </aside>

      <section className='conversation-panel' aria-label='Demo scam conversation'>
        <div className='conversation-header'>
          <strong>YOU ARE THE {mission.role === 'Seller' ? 'SELLER' : 'BUYER'}</strong>
          <span>›</span>
          <b>CHAT WITH {mission.role === 'Seller' ? 'BUYER' : 'SELLER'}</b>
          <small>DEMO MODE · #{mission.id.toUpperCase().slice(0, 8)}</small>
        </div>
        <div className='conversation-scroll' aria-live='polite'>
          {messages.map((message) => (
            <div className={`chat-row chat-row--${message.side}`} key={message.id}>
              <span className='chat-avatar'>
                <img
                  src={
                    message.side === 'seller'
                      ? `${PROVIDED}/seller-character.png`
                      : '/assets/anti-scam/buyer-frames/frame-01.png'
                  }
                  alt=''
                  className='pixel-art'
                />
              </span>
              <div className={`chat-bubble${message.tone ? ` chat-bubble--${message.tone}` : ''}`}>
                <p>{message.text}</p>
                <small>
                  {message.time}
                  {message.side === 'buyer' ? ' ✓✓' : ''}
                </small>
              </div>
            </div>
          ))}
        </div>

        <div className='response-area'>
          <div className='response-heading'>
            <span />
            <h2>✣ CHOOSE YOUR RESPONSE ✣</h2>
            <span />
          </div>
          <div className='response-grid'>
            {currentStep.choices.map((choice, index) => (
              <button
                className={`response-choice response-choice--${choice.tone}`}
                type='button'
                key={choice.text}
                onClick={() => chooseResponse(index)}
                disabled={completed}
              >
                <span className='response-choice__icon' aria-hidden='true'>
                  ▰
                </span>
                <strong>{choice.text}</strong>
                <b>›</b>
              </button>
            ))}
          </div>
        </div>
      </section>

      <aside className='play-sidebar play-sidebar--right'>
        <section className='play-panel signals-card'>
          <div className='play-panel__heading play-panel__heading--red'>
            <span>⚑</span>
            <h2>SCAM SIGNALS DETECTED</h2>
            <small>{Math.min(progress.step + 1, TOTAL_STEPS)}</small>
          </div>
          {[...DEMO_STEPS.slice(0, Math.min(progress.step + 1, 3))].map((step, index) => (
            <div className='signal-row' key={step.signal}>
              <b>{index === 0 ? '⚑' : index === 1 ? '⚠' : '☠'}</b>
              <span>{step.signal}</span>
            </div>
          ))}
          <button type='button' onClick={() => setTipsOpen(true)}>
            LEARN MORE <span>›</span>
          </button>
        </section>

        <section className={`play-panel tips-card${tipsOpen ? ' tips-card--open' : ''}`}>
          <div className='play-panel__heading'>
            <span>♙</span>
            <h2>PRACTICAL TIPS</h2>
            <small>{PRACTICAL_TIPS.length}</small>
          </div>
          {PRACTICAL_TIPS.map(([title, copy], index) => (
            <button
              className='tip-row'
              type='button'
              key={title}
              onClick={() => handleToast(`${title}: ${copy}`)}
            >
              <b>?</b>
              <span>
                <strong>{title}</strong>
                <small>{copy}</small>
              </span>
              <em>+{index === 3 ? 15 : 10} XP</em>
            </button>
          ))}
        </section>

        <section className='play-panel session-card'>
          <div className='play-panel__heading'>
            <span>⌁</span>
            <h2>SESSION STATS</h2>
          </div>
          <div className='session-card__stats'>
            <div>
              <span>◴</span>
              <small>TIME ELAPSED</small>
              <strong>{elapsedText}</strong>
            </div>
            <div>
              <span>★</span>
              <small>DEMO XP</small>
              <strong>{progress.xp}</strong>
            </div>
          </div>
          <div className='session-accuracy'>
            <span>Safe choices</span>
            <strong>{accuracy}%</strong>
          </div>
        </section>
      </aside>

      {completed && (
        <div className='mission-complete-backdrop'>
          <section
            className='mission-complete'
            role='dialog'
            aria-modal='true'
            aria-labelledby='mission-complete-title'
          >
            <div className='mission-complete__trophy'>🏆</div>
            <p>DEMO SCENARIO COMPLETE</p>
            <h1 id='mission-complete-title'>MISSION COMPLETE!</h1>
            <span>You identified the warning signs and finished the local training scenario.</span>
            <div className='mission-complete__stats'>
              <div>
                <small>XP EARNED</small>
                <strong>+{progress.xp}</strong>
              </div>
              <div>
                <small>SAFE CHOICES</small>
                <strong>{accuracy}%</strong>
              </div>
              <div>
                <small>RISK LEVEL</small>
                <strong>{riskLabel}</strong>
              </div>
            </div>
            <div className='mission-complete__actions'>
              <button type='button' onClick={restartMission}>
                Play Again
              </button>
              <button type='button' onClick={onExit}>
                Back to Mission Hub
              </button>
            </div>
          </section>
        </div>
      )}
    </main>
  )
}
