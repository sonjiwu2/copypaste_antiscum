import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  checkHealth,
  fetchScenarios,
  fetchScenarioById,
  startAttempt,
  restoreAttempt,
  submitChoice,
  ApiError,
} from '../client';

describe('API Client', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('checkHealth calls GET /healthz', async () => {
    const mockFetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({ status: 'ok' }),
    } as Response);

    const res = await checkHealth();
    expect(res).toEqual({ status: 'ok' });
    expect(mockFetch).toHaveBeenCalledWith('/healthz', expect.any(Object));
  });

  it('fetchScenarios calls GET /api/v1/scenarios with optional role filter', async () => {
    const mockFetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({
        scenarios: [
          {
            id: 'buyer-fake-delivery',
            version: 1,
            slug: 'buyer-fake-delivery',
            role: 'buyer',
            title: 'Test Scenario',
            description: 'Description',
            difficulty: 'easy',
            estimatedMinutes: 3,
          },
        ],
      }),
    } as Response);

    const list = await fetchScenarios('buyer');
    expect(list).toHaveLength(1);
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/scenarios?role=buyer', expect.any(Object));
  });

  it('fetchScenarioById calls GET /api/v1/scenarios/{id}', async () => {
    const mockFetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({
        id: 'buyer-fake-delivery',
        title: 'Test',
      }),
    } as Response);

    const scenario = await fetchScenarioById('buyer-fake-delivery');
    expect(scenario.id).toBe('buyer-fake-delivery');
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/scenarios/buyer-fake-delivery', expect.any(Object));
  });

  it('startAttempt calls POST /api/v1/attempts', async () => {
    const mockFetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({
        attemptId: 'att-123',
        status: 'in_progress',
        score: 100,
      }),
    } as Response);

    const attempt = await startAttempt('buyer-fake-delivery');
    expect(attempt.attemptId).toBe('att-123');
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/attempts', expect.objectContaining({
      method: 'POST',
      body: JSON.stringify({ scenarioId: 'buyer-fake-delivery' }),
    }));
  });

  it('restoreAttempt calls GET /api/v1/attempts/{attemptId}', async () => {
    const mockFetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({
        attemptId: 'att-123',
        status: 'in_progress',
      }),
    } as Response);

    const attempt = await restoreAttempt('att-123');
    expect(attempt.attemptId).toBe('att-123');
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/attempts/att-123', expect.any(Object));
  });

  it('submitChoice sends choice payload with idempotencyKey', async () => {
    const mockFetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({
        attemptId: 'att-123',
        status: 'in_progress',
        score: 90,
      }),
    } as Response);

    await submitChoice({
      attemptId: 'att-123',
      nodeId: 'node-1',
      choiceId: 'choice-1',
      idempotencyKey: 'test-key-uuid',
    });

    expect(mockFetch).toHaveBeenCalledWith('/api/v1/attempts/att-123/choices', expect.objectContaining({
      method: 'POST',
      body: JSON.stringify({
        nodeId: 'node-1',
        choiceId: 'choice-1',
        idempotencyKey: 'test-key-uuid',
      }),
    }));
  });

  it('handles backend error responses correctly with ApiError', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: false,
      status: 404,
      json: async () => ({
        error: {
          code: 'SCENARIO_NOT_FOUND',
          message: 'Scenario not found',
          requestId: 'req-abc',
        },
      }),
    } as Response);

    await expect(fetchScenarioById('unknown')).rejects.toThrow(ApiError);
  });
});
