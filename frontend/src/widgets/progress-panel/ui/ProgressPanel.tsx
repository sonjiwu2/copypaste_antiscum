import { useUserProgressStore } from '../../../entities/user-progress'
import { useProgress } from '../../../shared/api'
import { PixelIcon, type PixelIconName } from '../../../shared/ui/PixelIcon'

/** Сколько XP нужно на один уровень. Та же ступень, что в `calculateLevel`. */
const XP_PER_LEVEL = 100

export function ProgressPanel() {
  const totalXp = useUserProgressStore((state) => state.totalXp ?? 0)
  const level = useUserProgressStore((state) => state.level ?? 1)
  const streakDays = useUserProgressStore((state) => state.streakDays ?? 1)
  const unlockedAchievements = useUserProgressStore((state) => state.unlockedAchievements ?? [])
  const { progress, isLoading, isError } = useProgress()

  const xpInLevel = totalXp % XP_PER_LEVEL
  const levelPercent = Math.round((xpInLevel / XP_PER_LEVEL) * 100)

  const completedMissions = progress
    ? progress.summary.completedScenarios
    : isLoading
      ? '…'
      : isError
        ? '—'
        : 0

  const stats: Array<{
    icon: PixelIconName
    value: number | string
    label: string
    tone: string
  }> = [
    { icon: 'xp', value: totalXp, label: 'Всего XP', tone: 'gold' },
    { icon: 'star', value: completedMissions, label: 'Пройдено', tone: 'gold' },
    { icon: 'rewards', value: unlockedAchievements.length, label: 'Награды', tone: 'red' },
    { icon: 'streak', value: streakDays, label: 'Серия (дней)', tone: 'orange' },
  ]

  return (
    <section className='sidebar-panel progress-panel pixel-card' aria-labelledby='progress-title'>
      <h2 id='progress-title'>
        <PixelIcon name='progress' size={22} /> ВАШ ПРОГРЕСС
      </h2>
      <div className='level-row'>
        <strong>Уровень {level}</strong>
        <span>Уровень {level + 1}</span>
      </div>
      <div
        className='progress-track'
        role='progressbar'
        aria-valuemin={0}
        aria-valuemax={XP_PER_LEVEL}
        aria-valuenow={xpInLevel}
      >
        <span style={{ width: `${levelPercent}%` }} />
      </div>
      <div className='progress-caption'>
        {xpInLevel} / {XP_PER_LEVEL} XP (всего: {totalXp} XP)
      </div>
      <div className='progress-stats'>
        {stats.map((stat) => (
          <div className='progress-stat' key={stat.label}>
            <span className={`progress-stat__icon progress-stat__icon--${stat.tone}`}>
              <PixelIcon name={stat.icon} size={25} />
            </span>
            <strong>{stat.value}</strong>
            <small>{stat.label}</small>
          </div>
        ))}
      </div>
    </section>
  )
}
