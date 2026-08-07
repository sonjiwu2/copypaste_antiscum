import type {
  Attempt,
  ApiErrorDetail,
  ApiErrorResponse,
  HealthResponse,
  Scenario,
  ScenarioListResponse,
  ScenarioRole,
  SubmitChoiceParams,
  Transition,
} from './types'

export class ApiError extends Error {
  public readonly status: number
  public readonly code: string
  public readonly requestId?: string

  constructor(status: number, detail: ApiErrorDetail) {
    super(detail.message || `API Error HTTP ${status}`)
    this.name = 'ApiError'
    this.status = status
    this.code = detail.code || 'UNKNOWN_ERROR'
    this.requestId = detail.requestId
  }
}

const BASE_URL = ''

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    Accept: 'application/json',
    ...(options.headers as Record<string, string>),
  }

  if (options.body && !headers['Content-Type']) {
    headers['Content-Type'] = 'application/json'
  }

  const response = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    let errorDetail: ApiErrorDetail = {
      code: `HTTP_${response.status}`,
      message: `Request failed with status ${response.status}`,
    }

    try {
      const errorJson: ApiErrorResponse = await response.json()
      if (errorJson?.error) {
        errorDetail = errorJson.error
      }
    } catch {
      // Fallback if response body is not JSON
    }

    throw new ApiError(response.status, errorDetail)
  }

  return response.json() as Promise<T>
}

/**
 * 1. Проверка доступности бэкенда (Healthcheck)
 * GET /healthz
 */
export async function checkHealth(): Promise<HealthResponse> {
  return request<HealthResponse>('/healthz')
}

/**
 * 2. Получить каталог сценариев
 * GET /api/v1/scenarios?role={role}
 */
export async function fetchScenarios(role?: ScenarioRole): Promise<Scenario[]> {
  const query = role ? `?role=${encodeURIComponent(role)}` : ''
  const data = await request<ScenarioListResponse>(`/api/v1/scenarios${query}`)
  return data.scenarios
}

/**
 * 3. Получить описание одного сценария
 * GET /api/v1/scenarios/{scenarioId}
 */
export async function fetchScenarioById(scenarioId: string): Promise<Scenario> {
  return request<Scenario>(`/api/v1/scenarios/${encodeURIComponent(scenarioId)}`)
}

/**
 * 4. Начать прохождение сценария (Старт попытки)
 * POST /api/v1/attempts
 */
export async function startAttempt(scenarioId: string): Promise<Attempt> {
  return request<Attempt>('/api/v1/attempts', {
    method: 'POST',
    body: JSON.stringify({ scenarioId }),
  })
}

/**
 * 5. Восстановить состояние попытки (при F5 или рассинхронизации)
 * GET /api/v1/attempts/{attemptId}
 */
export async function restoreAttempt(attemptId: string): Promise<Attempt> {
  return request<Attempt>(`/api/v1/attempts/${encodeURIComponent(attemptId)}`)
}

/**
 * 6. Отправить выбор пользователя
 * POST /api/v1/attempts/{attemptId}/choices
 */

export async function submitChoice({
  attemptId,
  nodeId,
  choiceId,
  idempotencyKey = crypto.randomUUID(),
}: SubmitChoiceParams): Promise<Transition> {
  try {
    return await request<Transition>(`/api/v1/attempts/${encodeURIComponent(attemptId)}/choices`, {
      method: 'POST',
      body: JSON.stringify({
        nodeId,
        choiceId,
        idempotencyKey,
      }),
    })
  } catch (err) {
    if (
      err instanceof ApiError &&
      (err.code === 'STALE_NODE' || err.code === 'CONCURRENT_TRANSITION')
    ) {
      // Если пришла ошибка о рассинхронизации состояния, возвращаем актуальное состояние попытки
      // вызывающий код или API модуль может перепроверить попытку.
    }
    throw err
  }
}
