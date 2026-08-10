import { useMemo } from 'react'
import { useProgress, useScenarios, type ScenarioRole } from '../../../shared/api'
import { RoleCard } from './RoleCard'

type RoleStats = {
  missions: number
  completed: number
  accuracy: string
}

function statsForRole(
  role: ScenarioRole,
  scenarios: ReturnType<typeof useScenarios>['scenarios'],
  progress: ReturnType<typeof useProgress>['progress']
): RoleStats {
  const scenarioRoleById = new Map(scenarios.map((scenario) => [scenario.id, scenario.role]))
  const completed = (progress?.scenarioProgress ?? []).filter(
    (item) => item.attempts > 0 && (item.role ?? scenarioRoleById.get(item.scenarioId)) === role
  )

  const averageScore = completed.length
    ? Math.round(completed.reduce((sum, item) => sum + item.bestScore, 0) / completed.length)
    : null

  return {
    missions: scenarios.filter((scenario) => scenario.role === role).length,
    completed: completed.length,
    accuracy: averageScore === null ? '—' : `${averageScore}%`,
  }
}

export function RoleCardsGrid({
  onNavigate,
}: {
  onNavigate: (label: string, role: 'buyer' | 'seller') => void
}) {
  const { scenarios, isLoading: scenariosLoading, isError: scenariosError } = useScenarios()
  const { progress, isLoading: progressLoading, isError: progressError } = useProgress()

  const stats = useMemo(
    () => ({
      buyer: statsForRole('buyer', scenarios, progress),
      seller: statsForRole('seller', scenarios, progress),
    }),
    [progress, scenarios]
  )

  const missionsValue = (role: ScenarioRole) =>
    scenariosLoading ? '…' : scenariosError ? '—' : stats[role].missions
  const completedValue = (role: ScenarioRole) =>
    progressLoading ? '…' : progressError ? '—' : stats[role].completed
  const accuracyValue = (role: ScenarioRole) =>
    progressLoading ? '…' : progressError ? '—' : stats[role].accuracy

  return (
    <div className='role-grid'>
      <RoleCard
        role='buyer'
        description='Узнай, как покупать безопасно и избегать частых схем обмана.'
        badge='Проверенный покупатель'
        missions={missionsValue('buyer')}
        completed={completedValue('buyer')}
        accuracy={accuracyValue('buyer')}
        onAction={() => onNavigate('MISSIONS', 'buyer')}
      />
      <RoleCard
        role='seller'
        description='Узнай, как продавать безопасно и защищать свои объявления.'
        badge='Проверенный продавец'
        missions={missionsValue('seller')}
        completed={completedValue('seller')}
        accuracy={accuracyValue('seller')}
        onAction={() => onNavigate('MISSIONS', 'seller')}
      />
    </div>
  )
}
