import { useProgressQuery, useScenariosQuery } from './queries'
import type { ScenarioRole } from './types'

/** Каталог сценариев с бэкенда. */
export function useScenarios(role?: ScenarioRole) {
  const query = useScenariosQuery(role)

  return {
    scenarios: query.data ?? [],
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
    refetch: query.refetch,
  }
}

/** Прогресс анонимного профиля с бэкенда. */
export function useProgress() {
  const query = useProgressQuery()

  return {
    progress: query.data,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
    refetch: query.refetch,
  }
}
