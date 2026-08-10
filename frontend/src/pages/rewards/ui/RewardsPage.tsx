import { currentLocalDateKey, useUserProgressStore } from '../../../entities/user-progress'
import { useToastStore } from '../../../features/toast'
import { ART_ROOT } from '../../../shared/config/assets'
import { Glyph } from '../../../shared/ui/Glyph'
import { PixelIcon } from '../../../shared/ui/PixelIcon'
import './rewards.css'

const DAILY_REWARD_XP = 25

export function RewardsPage({ onBack }: { onBack: () => void }) {
  const totalXp = useUserProgressStore((state) => state.totalXp)
  const lastDailyRewardDate = useUserProgressStore((state) => state.lastDailyRewardDate)
  const claimDailyReward = useUserProgressStore((state) => state.claimDailyReward)
  const showToast = useToastStore((state) => state.showToast)
  const claimedToday = lastDailyRewardDate === currentLocalDateKey()

  const handleClaim = () => {
    if (!claimDailyReward(DAILY_REWARD_XP)) {
      return
    }

    showToast(`Ежедневная награда получена: +${DAILY_REWARD_XP} XP`)
  }

  return (
    <main className='rewards-page'>
      <section
        className={`rewards-panel${claimedToday ? ' rewards-panel--claimed' : ''}`}
        aria-labelledby='rewards-title'
      >
        <header className='rewards-panel__header'>
          <button type='button' onClick={onBack} aria-label='На главную'>
            ←
          </button>
          <h1 id='rewards-title'>НАГРАДЫ</h1>
          <span aria-hidden='true' />
        </header>

        <div className='rewards-gift-stage' aria-hidden='true'>
          <span className='rewards-spark rewards-spark--one'>✦</span>
          <span className='rewards-spark rewards-spark--two'>✦</span>
          <span className='rewards-spark rewards-spark--three'>✦</span>
          <img src={`${ART_ROOT}/rewards-gift.png`} alt='' className='pixel-art' />
          {claimedToday && <span className='rewards-gift-stage__check'>✓</span>}
        </div>

        <p className='rewards-panel__lead'>
          {claimedToday ? (
            <>
              Награда на сегодня собрана.
              <br />Возвращайся завтра за новой!
            </>
          ) : (
            <>
              Копи <b>XP</b>, открывай
              <br />бейджи и особые награды.
            </>
          )}
        </p>

        <div className='daily-reward-card' aria-live='polite'>
          <small>{claimedToday ? 'ПОЛУЧЕНО СЕГОДНЯ' : 'ЕЖЕДНЕВНАЯ НАГРАДА'}</small>
          <strong>+{DAILY_REWARD_XP} XP</strong>
          <PixelIcon name='xp' size={68} label='Опыт' />
        </div>

        <div className='rewards-total'>
          Текущий баланс: <b>{totalXp} XP</b>
        </div>

        <button
          className='rewards-claim-button'
          type='button'
          disabled={claimedToday}
          onClick={handleClaim}
        >
          <span>{claimedToday ? 'НАГРАДА СОБРАНА' : 'ЗАБРАТЬ НАГРАДУ'}</span>
          {claimedToday ? <span aria-hidden='true'>✓</span> : <Glyph name='chevron' size={20} />}
        </button>
      </section>
    </main>
  )
}
