import React from 'react'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  queryKeys,
  useAttemptQuery,
  useProgressQuery,
  useScenarioQuery,
  useScenariosQuery,
  useStartAttemptMutation,
  useSubmitChoiceMutation,
} from '../queries'
import * as client from '../client'
import type { Attempt, Progress, Scenario, Transition } from '../types'

vi.mock('../client', async () => {
  const actual = await vi.importActual<typeof import('../client')>('../client')

  return {
    ...actual,
    fetchScenarios: vi.fn(),
    fetchScenarioById: vi.fn(),
    fetchProgress: vi.fn(),
    startAttempt: vi.fn(),
    restoreAttempt: vi.fn(),
    submitChoice: vi.fn(),
  }
})

const SCENARIO: Scenario = {
  id: 'buyer-fake-delivery',
  version: 1,
  slug: 'buyer-fake-delivery',
  role: 'buyer',
  title: 'Ссылка на доставку',
  description: 'Продавец предлагает оплату по ссылке.',
  difficulty: 'medium',
  estimatedMinutes: 4,
  maxDecisions: 3,
}

const ATTEMPT: Attempt = {
  attemptId: 'att-1',
  scenario: {
    id: SCENARIO.id,
    version: 1,
    slug: SCENARIO.slug,
    role: 'buyer',
    title: SCENARIO.title,
  },
  status: 'in_progress',
  score: 100,
  currentNodeId: 'channel-decision',
  revealedNodes: [
    { id: 'greeting', type: 'message', sender: 'seller', text: 'Товар ещё доступен.' },
    {
      id: 'channel-decision',
      type: 'decision',
      prompt: 'Что вы сделаете?',
      choices: [{ id: 'stay-on-platform', label: 'Продолжить в приложении' }],
    },
  ],
  decisions: [],
  startedAt: '2026-08-08T10:00:00Z',
  updatedAt: '2026-08-08T10:00:00Z',
}

const TRANSITION: Transition = {
  attemptId: 'att-1',
  status: 'in_progress',
  score: 90,
  acceptedChoice: {
    nodeId: 'channel-decision',
    choiceId: 'stay-on-platform',
    label: 'Продолжить в приложении',
    playerReply: 'Давайте общаться здесь, в чате объявления.',
  },
  consequence: {
    severity: 'safe',
    title: 'Переписка сохранена',
    explanation: 'История общения остаётся внутри площадки.',
  },
  revealedNodes: [{ id: 'safe-reply', type: 'message', sender: 'seller', text: 'Хорошо.' }],
  currentNodeId: 'safe-reply',
}

function createClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function wrapperFor(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('Хуки запросов к API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('загружает каталог сценариев с фильтром по роли', async () => {
    vi.mocked(client.fetchScenarios).mockResolvedValueOnce([SCENARIO])

    const { result } = renderHook(() => useScenariosQuery('buyer'), {
      wrapper: wrapperFor(createClient()),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual([SCENARIO])
    expect(client.fetchScenarios).toHaveBeenCalledWith('buyer')
  })

  it('загружает описание одного сценария', async () => {
    vi.mocked(client.fetchScenarioById).mockResolvedValueOnce(SCENARIO)

    const { result } = renderHook(() => useScenarioQuery(SCENARIO.id), {
      wrapper: wrapperFor(createClient()),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.maxDecisions).toBe(3)
  })

  it('восстанавливает попытку по сохранённому идентификатору', async () => {
    vi.mocked(client.restoreAttempt).mockResolvedValueOnce(ATTEMPT)

    const { result } = renderHook(() => useAttemptQuery('att-1'), {
      wrapper: wrapperFor(createClient()),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(client.restoreAttempt).toHaveBeenCalledWith('att-1')
  })

  it('не запрашивает попытку, пока идентификатор неизвестен', () => {
    renderHook(() => useAttemptQuery(undefined), { wrapper: wrapperFor(createClient()) })

    expect(client.restoreAttempt).not.toHaveBeenCalled()
  })

  it('кладёт начатую попытку в кэш', async () => {
    vi.mocked(client.startAttempt).mockResolvedValueOnce(ATTEMPT)

    const queryClient = createClient()
    const { result } = renderHook(() => useStartAttemptMutation(), {
      wrapper: wrapperFor(queryClient),
    })

    result.current.mutate(SCENARIO.id)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(queryClient.getQueryData(queryKeys.attempts.detail('att-1'))).toEqual(ATTEMPT)
  })

  it('дополняет попытку результатом шага, а не заменяет её', async () => {
    vi.mocked(client.submitChoice).mockResolvedValueOnce(TRANSITION)

    const queryClient = createClient()
    queryClient.setQueryData(queryKeys.attempts.detail('att-1'), ATTEMPT)

    const { result } = renderHook(() => useSubmitChoiceMutation(), {
      wrapper: wrapperFor(queryClient),
    })

    result.current.mutate({
      attemptId: 'att-1',
      nodeId: 'channel-decision',
      choiceId: 'stay-on-platform',
      idempotencyKey: 'key-1',
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    const updated = queryClient.getQueryData<Attempt>(queryKeys.attempts.detail('att-1'))

    expect(updated?.score).toBe(90)
    expect(updated?.currentNodeId).toBe('safe-reply')
    expect(updated?.revealedNodes.map((node) => node.id)).toEqual([
      'greeting',
      'channel-decision',
      'safe-reply',
    ])
    expect(updated?.decisions).toEqual([
      {
        nodeId: 'channel-decision',
        choiceId: 'stay-on-platform',
        label: 'Продолжить в приложении',
        playerReply: TRANSITION.acceptedChoice.playerReply,
        consequence: TRANSITION.consequence,
      },
    ])
  })

  it('перечитывает прогресс после завершения сценария', async () => {
    vi.mocked(client.submitChoice).mockResolvedValueOnce({
      ...TRANSITION,
      status: 'completed',
      outcome: 'safe',
      completedAt: '2026-08-08T10:05:00Z',
    })

    const queryClient = createClient()
    queryClient.setQueryData(queryKeys.attempts.detail('att-1'), ATTEMPT)
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries')

    const { result } = renderHook(() => useSubmitChoiceMutation(), {
      wrapper: wrapperFor(queryClient),
    })

    result.current.mutate({
      attemptId: 'att-1',
      nodeId: 'channel-decision',
      choiceId: 'stay-on-platform',
      idempotencyKey: 'key-1',
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.progress })
  })

  it('загружает прогресс профиля', async () => {
    const progress: Progress = {
      summary: {
        completedAttempts: 2,
        completedScenarios: 1,
        averageScore: 85,
        bestScore: 95,
        latestScore: 75,
      },
      activeAttempts: [],
      scenarioProgress: [],
      skills: [],
      weakRiskTags: [],
      recentAttempts: [],
      recommendedScenarios: [],
    }

    vi.mocked(client.fetchProgress).mockResolvedValueOnce(progress)

    const { result } = renderHook(() => useProgressQuery(), {
      wrapper: wrapperFor(createClient()),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.summary.completedScenarios).toBe(1)
  })
})
