import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
  type UseMutationOptions,
} from '@tanstack/react-query';
import {
  checkHealth,
  fetchScenarios,
  fetchScenarioById,
  startAttempt,
  restoreAttempt,
  submitChoice,
} from './client';
import type {
  Attempt,
  HealthResponse,
  Scenario,
  ScenarioRole,
  SubmitChoiceParams,
  Transition,
} from './types';

export const queryKeys = {
  health: ['health'] as const,
  scenarios: {
    all: ['scenarios'] as const,
    list: (role?: ScenarioRole) => ['scenarios', 'list', role ?? 'all'] as const,
    detail: (id: string) => ['scenarios', 'detail', id] as const,
  },
  attempts: {
    all: ['attempts'] as const,
    detail: (id?: string) => ['attempts', 'detail', id] as const,
  },
};

/**
 * 1. Запрос проверки доступности бэкенда (Healthcheck)
 */
export function useHealthQuery(
  options?: Omit<UseQueryOptions<HealthResponse, Error>, 'queryKey' | 'queryFn'>
) {
  return useQuery<HealthResponse, Error>({
    queryKey: queryKeys.health,
    queryFn: checkHealth,
    staleTime: 30_000,
    ...options,
  });
}

/**
 * 2. Запрос каталога сценариев с возможной фильтрацией по роли
 */
export function useScenariosQuery(
  role?: ScenarioRole,
  options?: Omit<UseQueryOptions<Scenario[], Error>, 'queryKey' | 'queryFn'>
) {
  return useQuery<Scenario[], Error>({
    queryKey: queryKeys.scenarios.list(role),
    queryFn: () => fetchScenarios(role),
    staleTime: 60_000,
    ...options,
  });
}

/**
 * 3. Запрос подробностей конкретного сценария
 */
export function useScenarioQuery(
  scenarioId: string,
  options?: Omit<UseQueryOptions<Scenario, Error>, 'queryKey' | 'queryFn'>
) {
  return useQuery<Scenario, Error>({
    queryKey: queryKeys.scenarios.detail(scenarioId),
    queryFn: () => fetchScenarioById(scenarioId),
    enabled: Boolean(scenarioId),
    ...options,
  });
}

/**
 * 4. Запрос состояния попытки (восстановление при F5 или переходе)
 */
export function useAttemptQuery(
  attemptId?: string,
  options?: Omit<UseQueryOptions<Attempt, Error>, 'queryKey' | 'queryFn'>
) {
  return useQuery<Attempt, Error>({
    queryKey: queryKeys.attempts.detail(attemptId),
    queryFn: () => restoreAttempt(attemptId!),
    enabled: Boolean(attemptId),
    staleTime: 0,
    ...options,
  });
}

/**
 * 5. Мутация для создания новой попытки сценария
 */
export function useStartAttemptMutation(
  options?: UseMutationOptions<Attempt, Error, string>
) {
  const queryClient = useQueryClient();

  return useMutation<Attempt, Error, string>({
    ...options,
    mutationFn: (scenarioId: string) => startAttempt(scenarioId),
    onSuccess: (data, variables, context, mutation) => {
      queryClient.setQueryData(queryKeys.attempts.detail(data.attemptId), data);
      options?.onSuccess?.(data, variables, context, mutation);
    },
  });
}

/**
 * 6. Мутация для отправки ответа пользователя в шаге сценария
 */
export function useSubmitChoiceMutation(
  options?: UseMutationOptions<Transition, Error, SubmitChoiceParams>
) {
  const queryClient = useQueryClient();

  return useMutation<Transition, Error, SubmitChoiceParams>({
    ...options,
    mutationFn: (params: SubmitChoiceParams) => submitChoice(params),
    onSuccess: (transition, variables, context, mutation) => {
      queryClient.setQueryData<Attempt | undefined>(
        queryKeys.attempts.detail(variables.attemptId),
        (oldAttempt) => {
          if (!oldAttempt) return undefined;
          return {
            ...oldAttempt,
            status: transition.status,
            score: transition.score,
            currentNodeId: transition.currentNodeId,
            revealedNodes: [...oldAttempt.revealedNodes, ...transition.revealedNodes],
            updatedAt: new Date().toISOString(),
          };
        }
      );
      options?.onSuccess?.(transition, variables, context, mutation);
    },
  });
}

