import { PixelIcon, IconName } from '../../../shared/ui/PixelIcon'
import { useUserProgressStore } from '../../../entities/user-progress'

export function ProgressPanel() {
  const totalXp = useUserProgressStore((state) => state.totalXp ?? 0)
  const level = useUserProgressStore((state) => state.level ?? 1)
  const streakDays = useUserProgressStore((state) => state.streakDays ?? 1)
  const completedMissions = useUserProgressStore((state) => state.completedMissions ?? {})
  const unlockedAchievements = useUserProgressStore((state) => state.unlockedAchievements ?? [])

  const completedCount = Object.keys(completedMissions).length
  const xpInCurrentLevel = totalXp % 100
  const xpTarget = 100
  const progressPercent = Math.min(100, Math.round((xpInCurrentLevel / xpTarget) * 100))

  const stats: Array<{
    icon: IconName
    value: string
    label: string
    tone: string
  }> = [
    { icon: 'star', value: String(totalXp), label: 'Total XP', tone: 'gold' },
    { icon: 'trophy', value: String(completedCount), label: 'Пройдено', tone: 'gold' },
    { icon: 'target', value: String(unlockedAchievements.length), label: 'Награды', tone: 'red' },
    { icon: 'fire', value: String(streakDays), label: 'Стрик (дней)', tone: 'orange' },
  ]

  return (
    <section className='sidebar-panel progress-panel pixel-card' aria-labelledby='progress-title'>
      <h2 id='progress-title'>
        <PixelIcon name='chart' size={22} /> ВАШ ПРОГРЕСС
      </h2>
      <div className='level-row'>
        <strong>Уровень {level}</strong>
        <span>Уровень {level + 1}</span>
      </div>
      <div
        className='progress-track'
        role='progressbar'
        aria-valuemin={0}
        aria-valuemax={xpTarget}
        aria-valuenow={xpInCurrentLevel}
      >
        <span style={{ width: `${progressPercent}%` }} />
      </div>
      <div className='progress-caption'>
        {xpInCurrentLevel} / {xpTarget} XP (Всего: {totalXp} XP)
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
