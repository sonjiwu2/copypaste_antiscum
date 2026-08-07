import { useState, useCallback } from 'react';
import {
  useScenariosQuery,
  useScenarioQuery,
  useAttemptQuery,
  useStartAttemptMutation,
  useSubmitChoiceMutation,
} from './queries';
import type { ScenarioRole } from './types';

/**
 * Хук для работы с каталогом сценариев через TanStack Query
 */
export function useScenarios(role?: ScenarioRole) {
  const query = useScenariosQuery(role);

  return {
    scenarios: query.data ?? [],
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
    refetch: query.refetch,
  };
}

/**
 * Хук для получения информации о сценарии по ID через TanStack Query
 */
export function useScenarioDetail(scenarioId: string) {
  const query = useScenarioQuery(scenarioId);

  return {
    scenario: query.data,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
  };
}

/**
 * Хук для прохождения сценария (старт попытки, отправка выборов) через TanStack Query
 */
export function useScenarioRunner(initialAttemptId?: string) {
  const [currentAttemptId, setCurrentAttemptId] = useState<string | undefined>(
    initialAttemptId
  );

  // Запрос восстановления/актуализации попытки в TanStack Query
  const attemptQuery = useAttemptQuery(currentAttemptId);

  // Мутация старта попытки
  const startMutation = useStartAttemptMutation({
    onSuccess: (attempt) => {
      setCurrentAttemptId(attempt.attemptId);
    },
  });

  // Мутация отправки выбора пользователя
  const submitMutation = useSubmitChoiceMutation();

  const start = useCallback(
    async (scenarioId: string) => {
      return startMutation.mutateAsync(scenarioId);
    },
    [startMutation]
  );

  const choose = useCallback(
    async (nodeId: string, choiceId: string) => {
      if (!currentAttemptId) {
        throw new Error('Попытка не начата');
      }
      return submitMutation.mutateAsync({
        attemptId: currentAttemptId,
        nodeId,
        choiceId,
      });
    },
    [currentAttemptId, submitMutation]
  );

  return {
    attemptId: currentAttemptId,
    attempt: attemptQuery.data,
    isLoading: attemptQuery.isLoading || startMutation.isPending,
    isSubmitting: submitMutation.isPending,
    error: attemptQuery.error || startMutation.error || submitMutation.error,
    start,
    choose,
    reset: () => setCurrentAttemptId(undefined),
  };
}
