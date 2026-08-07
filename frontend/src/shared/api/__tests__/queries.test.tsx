import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  useHealthQuery,
  useScenariosQuery,
  useScenarioQuery,
  useAttemptQuery,
  useStartAttemptMutation,
  useSubmitChoiceMutation,
} from '../queries';
import { useScenarios, useScenarioRunner } from '../hooks';
import * as client from '../client';

vi.mock('../client', async () => {
  const actual = await vi.importActual<typeof import('../client')>('../client');
  return {
    ...actual,
    checkHealth: vi.fn(),
    fetchScenarios: vi.fn(),
    fetchScenarioById: vi.fn(),
    startAttempt: vi.fn(),
    restoreAttempt: vi.fn(),
    submitChoice: vi.fn(),
  };
});

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe('TanStack Query Hooks', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('useHealthQuery fetches health status', async () => {
    vi.mocked(client.checkHealth).mockResolvedValueOnce({ status: 'ok' });

    const { result } = renderHook(() => useHealthQuery(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({ status: 'ok' });
    expect(client.checkHealth).toHaveBeenCalledTimes(1);
  });

  it('useScenariosQuery fetches scenarios list', async () => {
    const mockScenarios = [
      {
        id: 'buyer-fake-delivery',
        version: 1,
        slug: 'buyer-fake-delivery',
        role: 'buyer' as const,
        title: 'Fake Delivery',
        description: 'Test description',
        difficulty: 'medium' as const,
        estimatedMinutes: 4,
      },
    ];

    vi.mocked(client.fetchScenarios).mockResolvedValueOnce(mockScenarios);

    const { result } = renderHook(() => useScenariosQuery('buyer'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockScenarios);
    expect(client.fetchScenarios).toHaveBeenCalledWith('buyer');
  });

  it('useScenarioQuery fetches single scenario detail', async () => {
    const mockScenario = {
      id: 'buyer-fake-delivery',
      version: 1,
      slug: 'buyer-fake-delivery',
      role: 'buyer' as const,
      title: 'Fake Delivery',
      description: 'Test description',
      difficulty: 'medium' as const,
      estimatedMinutes: 4,
    };

    vi.mocked(client.fetchScenarioById).mockResolvedValueOnce(mockScenario);

    const { result } = renderHook(() => useScenarioQuery('buyer-fake-delivery'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockScenario);
  });

  it('useAttemptQuery restores attempt when attemptId is provided', async () => {
    const mockAttempt = {
      attemptId: 'att-999',
      status: 'in_progress' as const,
      score: 100,
      currentNodeId: 'node-1',
      revealedNodes: [],
      startedAt: '2026-08-07T00:00:00Z',
      updatedAt: '2026-08-07T00:00:00Z',
    };

    vi.mocked(client.restoreAttempt).mockResolvedValueOnce(mockAttempt);

    const { result } = renderHook(() => useAttemptQuery('att-999'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockAttempt);
  });

  it('useStartAttemptMutation starts a new attempt', async () => {
    const mockAttempt = {
      attemptId: 'att-100',
      status: 'in_progress' as const,
      score: 100,
      currentNodeId: 'node-1',
      revealedNodes: [],
      startedAt: '2026-08-07T00:00:00Z',
      updatedAt: '2026-08-07T00:00:00Z',
    };

    vi.mocked(client.startAttempt).mockResolvedValueOnce(mockAttempt);

    const { result } = renderHook(() => useStartAttemptMutation(), {
      wrapper: createWrapper(),
    });

    result.current.mutate('buyer-fake-delivery');

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockAttempt);
  });

  it('useScenarioRunner hook orchestrates start and choose', async () => {
    const mockAttempt = {
      attemptId: 'att-555',
      status: 'in_progress' as const,
      score: 100,
      currentNodeId: 'node-start',
      revealedNodes: [],
      startedAt: '2026-08-07T00:00:00Z',
      updatedAt: '2026-08-07T00:00:00Z',
    };

    const mockTransition = {
      attemptId: 'att-555',
      status: 'in_progress' as const,
      score: 90,
      consequence: {
        severity: 'safe' as const,
        title: 'Good choice',
        explanation: 'You stayed on platform',
      },
      revealedNodes: [],
      currentNodeId: 'node-2',
    };

    vi.mocked(client.startAttempt).mockResolvedValueOnce(mockAttempt);
    vi.mocked(client.submitChoice).mockResolvedValueOnce(mockTransition);

    const { result } = renderHook(() => useScenarioRunner(), {
      wrapper: createWrapper(),
    });

    await waitFor(async () => {
      await result.current.start('buyer-fake-delivery');
    });

    await waitFor(() => {
      expect(result.current.attemptId).toBe('att-555');
    });

    await waitFor(async () => {
      await result.current.choose('node-start', 'choice-stay');
    });

    expect(client.submitChoice).toHaveBeenCalledWith({
      attemptId: 'att-555',
      nodeId: 'node-start',
      choiceId: 'choice-stay',
    });
  });

});
