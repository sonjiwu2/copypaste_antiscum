import { useMemo } from 'react'
import { PixelIcon } from '../../../shared/ui/PixelIcon'
import { ACHIEVEMENTS, earnedXpFor, useUserProgressStore } from '../../../entities/user-progress'
import { missionXpFor } from '../../../entities/mission'
import { useProgress, useScenarios } from '../../../shared/api'
import './progress.css'

/** Сколько XP нужно на уровень. Та же ступень, что в `calculateLevel`. */
const XP_PER_LEVEL = 100

export interface ProgressPageProps {
  onNavigate?: (route: string) => void
  onLaunch?: (missionId: string) => void
}

function formatDate(isoDate?: string): string {
  if (!isoDate) {
    return '—'
  }

  return new Date(isoDate).toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

export function ProgressPage({ onNavigate, onLaunch }: ProgressPageProps) {
  const totalXp = useUserProgressStore((state) => state.totalXp ?? 0)
  const level = useUserProgressStore((state) => state.level ?? 1)
  const streakDays = useUserProgressStore((state) => state.streakDays ?? 1)
  const unlockedAchievements = useUserProgressStore((state) => state.unlockedAchievements ?? [])

  // Статистика прохождений приходит с бэкенда: он считает её по сохранённым
  // попыткам. Локально остаётся только геймификация — уровень и награды.
  const { progress } = useProgress()
  const { scenarios } = useScenarios()

  const completedCount = progress?.summary.completedScenarios ?? 0
  const xpInCurrentLevel = totalXp % XP_PER_LEVEL
  const levelPercent = Math.round((xpInCurrentLevel / XP_PER_LEVEL) * 100)

  const completedList = useMemo(() => {
    // Награда считается по сложности сценария и лучшему результату, поэтому
    // строка таблицы одинакова на любом устройстве и не зависит от того,
    // проходился ли сценарий в этом браузере.
    const difficultyById = new Map(scenarios.map((item) => [item.id, item.difficulty]))

    return (progress?.scenarioProgress ?? [])
      .filter((item) => item.lastCompletedAt)
      .map((item) => ({
        id: item.scenarioId,
        title: item.title || item.scenarioId,
        bestScore: item.bestScore,
        xpEarned: earnedXpFor(missionXpFor(difficultyById.get(item.scenarioId) ?? ''), item.bestScore),
        attemptsCount: item.attempts,
        completedAt: formatDate(item.lastCompletedAt),
      }))
  }, [progress, scenarios])

  return (
    <main className='progress-page-layout'>
      <header className='progress-header'>
        <div className='progress-header__title'>
          <PixelIcon name='progress' size={36} />
          <div>
            <h1>ТАБЛИЦА ПРОГРЕССА</h1>
            <p>Ваши достижения, уровень и история прохождения тренировочных миссий</p>
          </div>
        </div>
        <div className='progress-header__actions'>
          {onNavigate && (
            <button type='button' className='btn-primary' onClick={() => onNavigate('MISSIONS')}>
              <PixelIcon name='missions' size={18} /> К миссиям
            </button>
          )}
        </div>
      </header>

      <section className='progress-dashboard'>
        <div className='progress-card'>
          <div className='progress-card__heading'>
            <PixelIcon name='xp' size={22} /> ИГРОВОЙ УРОВЕНЬ
          </div>
          <div className='level-banner'>
            <div className='level-banner__info'>
              <div>
                <span className='level-banner__badge'>Уровень {level}</span>
                <span className='level-banner__rank'>
                  {level === 1 ? 'Стажёр антискама' : level < 5 ? 'Агент защиты' : 'Элитный эксперт'}
                </span>
              </div>
              <span className='level-banner__next'>Цель: Уровень {level + 1}</span>
            </div>

            <div className='progress-page-track'>
              <span className='progress-page-track__fill' style={{ width: `${levelPercent}%` }} />
              <span className='progress-page-track__text'>{levelPercent}% ({xpInCurrentLevel} / {XP_PER_LEVEL} XP)</span>
            </div>

            <div className='level-details-grid'>
              <div className='level-detail-item'>
                <small>Прогресс уровня</small>
                <strong>{xpInCurrentLevel} / {XP_PER_LEVEL} XP</strong>
              </div>
              <div className='level-detail-item'>
                <small>До {level + 1} уровня осталось</small>
                <strong>{XP_PER_LEVEL - xpInCurrentLevel} XP</strong>
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
            <PixelIcon name='progress' size={22} /> СТАТИСТИКА
          </div>
          <div className='stats-grid-4'>
            <div className='stat-box'>
              <span className='stat-box__icon stat-box__icon--gold'>
                <PixelIcon name='xp' size={24} />
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
                <PixelIcon name='rewards' size={24} />
              </span>
              <strong>{unlockedAchievements.length}</strong>
              <span>Награды</span>
            </div>
            <div className='stat-box'>
              <span className='stat-box__icon stat-box__icon--orange'>
                <PixelIcon name='streak' size={24} />
              </span>
              <strong>{streakDays}</strong>
              <span>Серия (дней)</span>
            </div>
          </div>
        </div>
      </section>

      <section className='achievements-section'>
        <div className='progress-card__heading'>
          <PixelIcon name='trophy' size={22} /> ДОСТИЖЕНИЯ И НАГРАДЫ
        </div>
        <div className='achievements-grid'>
          {ACHIEVEMENTS.map((ach) => {
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
          <PixelIcon name='rules' size={22} /> ТАБЛИЦА ПРОЙДЕННЫХ МИССИЙ
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
                        className='btn-primary btn-primary--compact'
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
            <PixelIcon name='missions' size={48} />
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
