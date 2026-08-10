import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
  type UseMutationOptions,
} from '@tanstack/react-query'
import {
  fetchProgress,
  fetchScenarios,
  fetchScenarioById,
  startAttempt,
  restoreAttempt,
  submitChoice,
  ensureWeeklyTest,
  checkWeeklyTestAnswer,
  submitWeeklyTest,
} from './client'
import type {
  Attempt,
  Progress,
  Scenario,
  ScenarioRole,
  SubmitChoiceParams,
  Transition,
  WeeklyTest,
  SubmitWeeklyTestParams,
  CheckWeeklyTestAnswerParams,
  CheckWeeklyTestAnswerResult,
} from './types'

export const queryKeys = {
  progress: ['progress'] as const,
  scenarios: {
    all: ['scenarios'] as const,
    list: (role?: ScenarioRole) => ['scenarios', 'list', role ?? 'all'] as const,
    detail: (id: string) => ['scenarios', 'detail', id] as const,
  },
  attempts: {
    all: ['attempts'] as const,
    detail: (id?: string) => ['attempts', 'detail', id] as const,
  },
  weeklyTest: ['weekly-test', 'current'] as const,
}

export function useScenariosQuery(
  role?: ScenarioRole,
  options?: Omit<UseQueryOptions<Scenario[], Error>, 'queryKey' | 'queryFn'>
) {
  return useQuery<Scenario[], Error>({
    queryKey: queryKeys.scenarios.list(role),
    queryFn: () => fetchScenarios(role),
    staleTime: 60_000,
    ...options,
  })
}

export function useScenarioQuery(
  scenarioId: string,
  options?: Omit<UseQueryOptions<Scenario, Error>, 'queryKey' | 'queryFn'>
) {
  return useQuery<Scenario, Error>({
    queryKey: queryKeys.scenarios.detail(scenarioId),
    queryFn: () => fetchScenarioById(scenarioId),
    enabled: Boolean(scenarioId),
    ...options,
  })
}

/** Состояние попытки: тем же запросом экран восстанавливается после перезагрузки. */
export function useAttemptQuery(
  attemptId?: string,
  options?: Omit<UseQueryOptions<Attempt, Error>, 'queryKey' | 'queryFn'>
) {
  return useQuery<Attempt, Error>({
    queryKey: queryKeys.attempts.detail(attemptId),
    queryFn: () => restoreAttempt(attemptId as string),
    enabled: Boolean(attemptId),
    staleTime: 0,
    // Потерянная или чужая попытка не чинится повтором: её обрабатывает вызывающий код.
    retry: false,
    ...options,
  })
}

/** Прогресс профиля. Считается на сервере по сохранённым попыткам. */
export function useProgressQuery(
  options?: Omit<UseQueryOptions<Progress, Error>, 'queryKey' | 'queryFn'>
) {
  return useQuery<Progress, Error>({
    queryKey: queryKeys.progress,
    queryFn: fetchProgress,
    staleTime: 10_000,
    ...options,
  })
}

export function useStartAttemptMutation(options?: UseMutationOptions<Attempt, Error, string>) {
  const queryClient = useQueryClient()

  return useMutation<Attempt, Error, string>({
    ...options,
    mutationFn: (scenarioId: string) => startAttempt(scenarioId),
    onSuccess: (attempt, variables, context, mutation) => {
      queryClient.setQueryData(queryKeys.attempts.detail(attempt.attemptId), attempt)
      // Новая попытка появляется в списке активных на экране прогресса.
      void queryClient.invalidateQueries({ queryKey: queryKeys.progress })
      options?.onSuccess?.(attempt, variables, context, mutation)
    },
  })
}

export function useSubmitChoiceMutation(
  options?: UseMutationOptions<Transition, Error, SubmitChoiceParams>
) {
  const queryClient = useQueryClient()

  return useMutation<Transition, Error, SubmitChoiceParams>({
    ...options,
    mutationFn: (params: SubmitChoiceParams) => submitChoice(params),
    onSuccess: (transition, variables, context, mutation) => {
      queryClient.setQueryData<Attempt>(queryKeys.attempts.detail(variables.attemptId), (cached) =>
        cached ? applyTransition(cached, transition) : cached
      )

      if (transition.status === 'completed') {
        void queryClient.invalidateQueries({ queryKey: queryKeys.progress })
      }

      options?.onSuccess?.(transition, variables, context, mutation)
    },
  })
}

export function useWeeklyTestQuery(
  options?: Omit<UseQueryOptions<WeeklyTest, Error>, 'queryKey' | 'queryFn'>
) {
  return useQuery<WeeklyTest, Error>({
    queryKey: queryKeys.weeklyTest,
    queryFn: ensureWeeklyTest,
    staleTime: 60_000,
    retry: false,
    ...options,
  })
}

export function useSubmitWeeklyTestMutation(
  options?: UseMutationOptions<WeeklyTest, Error, SubmitWeeklyTestParams>
) {
  const queryClient = useQueryClient()

  return useMutation<WeeklyTest, Error, SubmitWeeklyTestParams>({
    ...options,
    mutationFn: submitWeeklyTest,
    onSuccess: (test, variables, context, mutation) => {
      queryClient.setQueryData(queryKeys.weeklyTest, test)
      options?.onSuccess?.(test, variables, context, mutation)
    },
  })
}

export function useCheckWeeklyTestAnswerMutation(
  options?: UseMutationOptions<CheckWeeklyTestAnswerResult, Error, CheckWeeklyTestAnswerParams>
) {
  return useMutation<CheckWeeklyTestAnswerResult, Error, CheckWeeklyTestAnswerParams>({
    ...options,
    mutationFn: checkWeeklyTestAnswer,
  })
}

/**
 * Переносит результат шага в состояние попытки.
 *
 * Шаг возвращает только раскрытые им узлы, поэтому история дополняется,
 * а не заменяется.
 */
function applyTransition(attempt: Attempt, transition: Transition): Attempt {
  return {
    ...attempt,
    status: transition.status,
    score: transition.score,
    outcome: transition.outcome,
    currentNodeId: transition.currentNodeId,
    revealedNodes: [...attempt.revealedNodes, ...transition.revealedNodes],
    decisions: [
      ...attempt.decisions,
      {
        nodeId: transition.acceptedChoice.nodeId,
        choiceId: transition.acceptedChoice.choiceId,
        label: transition.acceptedChoice.label,
        playerReply: transition.acceptedChoice.playerReply,
        consequence: transition.consequence,
      },
    ],
    updatedAt: new Date().toISOString(),
    completedAt: transition.completedAt,
  }
}
