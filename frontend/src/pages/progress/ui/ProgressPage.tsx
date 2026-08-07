import { useMemo } from 'react'
import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { useUserProgressStore } from '../../../entities/user-progress'
import { useScenarios } from '../../../shared/api'
import '../../../app/progress.css'

export interface ProgressPageProps {
  onNavigate?: (route: string) => void
  onLaunch?: (missionId: string) => void
}

const ALL_ACHIEVEMENTS = [
  {
    id: 'first_mission',
    title: 'Первый шаг',
    description: 'Успешно завершите вашу первую миссию по безопасности.',
    icon: 'star' as const,
  },
  {
    id: 'five_missions',
    title: 'Опытный защитник',
    description: 'Пройдите 5 миссий и закрепите навыки распознавания мошенников.',
    icon: 'target' as const,
  },
  {
    id: 'perfect_score',
    title: 'Мастер безопасности',
    description: 'Завершите любую миссию со 100% результатом без ошибок.',
    icon: 'shield' as const,
  },
  {
    id: 'streak_3',
    title: 'Страж порядка',
    description: 'Сохраняйте активность 3 дня подряд.',
    icon: 'fire' as const,
  },
]

export function ProgressPage({ onNavigate, onLaunch }: ProgressPageProps) {
  const totalXp = useUserProgressStore((state) => state.totalXp ?? 0)
  const level = useUserProgressStore((state) => state.level ?? 1)
  const streakDays = useUserProgressStore((state) => state.streakDays ?? 1)
  const completedMissions = useUserProgressStore((state) => state.completedMissions ?? {})
  const unlockedAchievements = useUserProgressStore((state) => state.unlockedAchievements ?? [])

  const { scenarios } = useScenarios()

  const completedCount = Object.keys(completedMissions).length
  const xpInCurrentLevel = totalXp % 100
  const xpTarget = 100
  const progressPercent = Math.min(100, Math.round((xpInCurrentLevel / xpTarget) * 100))

  const completedList = useMemo(() => {
    return Object.entries(completedMissions).map(([id, record]) => {
      const matchedScenario = scenarios?.find((s) => String(s.id) === String(id) || s.slug === id)
      const title = matchedScenario?.title || matchedScenario?.slug || `Миссия #${id}`
      return {
        id,
        title,
        bestScore: record.bestScore,
        xpEarned: record.xpEarned,
        attemptsCount: record.attemptsCount,
        completedAt: record.completedAt
          ? new Date(record.completedAt).toLocaleDateString('ru-RU', {
              day: '2-digit',
              month: '2-digit',
              year: 'numeric',
            })
          : '—',
      }
    })
  }, [completedMissions, scenarios])

  return (
    <main className='progress-page-layout'>
      <header className='progress-header'>
        <div className='progress-header__title'>
          <PixelIcon name='chart' size={36} />
          <div>
            <h1>ТАБЛИЦА ПРОГРЕССА</h1>
            <p>Ваши достижения, уровень и история прохождения тренировочных миссий</p>
          </div>
        </div>
        <div className='progress-header__actions'>
          {onNavigate && (
            <button type='button' className='btn-primary' onClick={() => onNavigate('MISSIONS')}>
              <PixelIcon name='star' size={18} /> К миссиям
            </button>
          )}
        </div>
      </header>

      <section className='progress-dashboard'>
        <div className='progress-card'>
          <div className='progress-card__heading'>
            <PixelIcon name='bolt' size={22} /> ИГРОВОЙ УРОВЕНЬ
          </div>
          <div className='level-banner'>
            <div className='level-banner__info'>
              <div>
                <span className='level-banner__badge'>Уровень {level}</span>
                <span className='level-banner__rank' style={{ marginLeft: '10px' }}>
                  {level === 1 ? 'Стажёр антискама' : level < 5 ? 'Агент защиты' : 'Элитный эксперт'}
                </span>
              </div>
              <span className='level-banner__next'>Цель: Уровень {level + 1}</span>
            </div>

            <div className='progress-page-track'>
              <span className='progress-page-track__fill' style={{ width: `${progressPercent}%` }} />
              <span className='progress-page-track__text'>{progressPercent}% ({xpInCurrentLevel} / {xpTarget} XP)</span>
            </div>

            <div className='level-details-grid'>
              <div className='level-detail-item'>
                <small>Прогресс уровня</small>
                <strong>{xpInCurrentLevel} / {xpTarget} XP</strong>
              </div>
              <div className='level-detail-item'>
                <small>До {level + 1} уровня осталось</small>
                <strong>{xpTarget - xpInCurrentLevel} XP</strong>
              </div>
            </div>

            <div className='level-reward-box'>
              <PixelIcon name='trophy' size={20} />
              <span>
                Уровень {level + 1} открывает доступ к новым миссиям высокого уровня сложности!
              </span>
            </div>
          </div>
        </div>


        <div className='progress-card'>
          <div className='progress-card__heading'>
            <PixelIcon name='target' size={22} /> СТАТИСТИКА
          </div>
          <div className='stats-grid-4'>
            <div className='stat-box'>
              <span className='stat-box__icon stat-box__icon--gold'>
                <PixelIcon name='star' size={24} />
              </span>
              <strong>{totalXp}</strong>
              <span>Всего XP</span>
            </div>
            <div className='stat-box'>
              <span className='stat-box__icon stat-box__icon--blue'>
                <PixelIcon name='trophy' size={24} />
              </span>
              <strong>{completedCount}</strong>
              <span>Пройдено</span>
            </div>
            <div className='stat-box'>
              <span className='stat-box__icon stat-box__icon--red'>
                <PixelIcon name='target' size={24} />
              </span>
              <strong>{unlockedAchievements.length}</strong>
              <span>Награды</span>
            </div>
            <div className='stat-box'>
              <span className='stat-box__icon stat-box__icon--orange'>
                <PixelIcon name='fire' size={24} />
              </span>
              <strong>{streakDays}</strong>
              <span>Стрик (дней)</span>
            </div>
          </div>
        </div>
      </section>

      <section className='achievements-section'>
        <div className='progress-card__heading'>
          <PixelIcon name='trophy' size={22} /> ДОСТИЖЕНИЯ И НАГРАДЫ
        </div>
        <div className='achievements-grid'>
          {ALL_ACHIEVEMENTS.map((ach) => {
            const isUnlocked = unlockedAchievements.includes(ach.id)
            return (
              <div
                key={ach.id}
                className={`achievement-card${isUnlocked ? ' achievement-card--unlocked' : ''}`}
              >
                <div className='achievement-card__icon'>
                  <PixelIcon name={ach.icon} size={22} />
                </div>
                <div className='achievement-card__info'>
                  <h4>{ach.title}</h4>
                  <p>{ach.description}</p>
                  <span
                    className={`achievement-badge${isUnlocked ? ' achievement-badge--unlocked' : ''}`}
                  >
                    {isUnlocked ? 'Разблокировано' : 'Заблокировано'}
                  </span>
                </div>
              </div>
            )
          })}
        </div>
      </section>

      <section className='missions-table-section'>
        <div className='progress-card__heading'>
          <PixelIcon name='book' size={22} /> ТАБЛИЦА ПРОЙДЕННЫХ МИССИЙ
        </div>
        {completedList.length > 0 ? (
          <table className='progress-table'>
            <thead>
              <tr>
                <th>Миссия</th>
                <th>Лучший результат</th>
                <th>Заработано XP</th>
                <th>Попытки</th>
                <th>Дата</th>
                <th>Действие</th>
              </tr>
            </thead>
            <tbody>
              {completedList.map((item) => (
                <tr key={item.id}>
                  <td>
                    <strong>{item.title}</strong>
                  </td>
                  <td>
                    <span
                      className={`score-pill ${
                        item.bestScore >= 90
                          ? 'score-pill--high'
                          : item.bestScore >= 60
                            ? 'score-pill--medium'
                            : 'score-pill--low'
                      }`}
                    >
                      {item.bestScore}%
                    </span>
                  </td>
                  <td>+{item.xpEarned} XP</td>
                  <td>{item.attemptsCount}</td>
                  <td>{item.completedAt}</td>
                  <td>
                    {onLaunch && (
                      <button
                        type='button'
                        className='btn-primary'
                        style={{ padding: '4px 10px', fontSize: '12px' }}
                        onClick={() => onLaunch(item.id)}
                      >
                        Перепройти
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div className='empty-progress-state'>
            <PixelIcon name='target' size={48} />
            <p>Вы пока не завершили ни одной тренировочной миссии.</p>
            {onNavigate && (
              <button
                type='button'
                className='btn-primary'
                onClick={() => onNavigate('MISSIONS')}
              >
                Пройти первую миссию
              </button>
            )}
          </div>
        )}
      </section>
    </main>
  )
}
