import type {
  ApiErrorDetail,
  ApiErrorResponse,
  Attempt,
  HealthResponse,
  Progress,
  Scenario,
  ScenarioListResponse,
  ScenarioRole,
  SubmitChoiceParams,
  Transition,
  WeeklyTest,
  SubmitWeeklyTestParams,
  CheckWeeklyTestAnswerParams,
  CheckWeeklyTestAnswerResult,
} from './types'

/** Коды ошибок API. Клиент опирается на код, а не на текст сообщения. */
export const ApiErrorCode = {
  attemptNotFound: 'ATTEMPT_NOT_FOUND',
  attemptForbidden: 'ATTEMPT_FORBIDDEN',
  attemptAlreadyCompleted: 'ATTEMPT_ALREADY_COMPLETED',
  staleNode: 'STALE_NODE',
  concurrentTransition: 'CONCURRENT_TRANSITION',
} as const

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

/** Попытка потеряна на сервере: её нужно начать заново, а не восстанавливать. */
export function isAttemptGone(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    (error.code === ApiErrorCode.attemptNotFound || error.code === ApiErrorCode.attemptForbidden)
  )
}

/** Состояние клиента разошлось с сервером: экран нужно перечитать. */
export function isAttemptOutOfSync(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    (error.code === ApiErrorCode.staleNode ||
      error.code === ApiErrorCode.concurrentTransition ||
      error.code === ApiErrorCode.attemptAlreadyCompleted)
  )
}

/**
 * Ключ идемпотентности шага. Повтор запроса с тем же ключом возвращает
 * прежний результат вместо второго перехода по сценарию.
 *
 * crypto.randomUUID недоступен вне защищённого контекста, поэтому нужен
 * запасной вариант: демо может открываться по http.
 */
export function newIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }

  return `key-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    Accept: 'application/json',
    ...(options.headers as Record<string, string>),
  }

  if (options.body && !headers['Content-Type']) {
    headers['Content-Type'] = 'application/json'
  }

  const response = await fetch(path, {
    ...options,
    headers,
    // Cookie анонимного профиля обязательна: без неё каждый запрос
    // заводил бы нового пользователя и терял историю попыток.
    credentials: 'same-origin',
  })

  if (!response.ok) {
    throw new ApiError(response.status, await errorDetailOf(response))
  }

  return response.json() as Promise<T>
}

async function errorDetailOf(response: Response): Promise<ApiErrorDetail> {
  try {
    const body = (await response.json()) as ApiErrorResponse
    if (body?.error) {
      return body.error
    }
  } catch {
    // Тело ошибки не JSON: остаётся описать ответ по статусу.
  }

  return {
    code: `HTTP_${response.status}`,
    message: `Request failed with status ${response.status}`,
  }
}

/** GET /healthz — проверка доступности бэкенда. */
export async function checkHealth(): Promise<HealthResponse> {
  return request<HealthResponse>('/healthz')
}

/** GET /api/v1/scenarios — каталог сценариев. */
export async function fetchScenarios(role?: ScenarioRole): Promise<Scenario[]> {
  const query = role ? `?role=${encodeURIComponent(role)}` : ''
  const data = await request<ScenarioListResponse>(`/api/v1/scenarios${query}`)

  return data.scenarios
}

/** GET /api/v1/scenarios/{scenarioId} — описание одного сценария. */
export async function fetchScenarioById(scenarioId: string): Promise<Scenario> {
  return request<Scenario>(`/api/v1/scenarios/${encodeURIComponent(scenarioId)}`)
}

/** POST /api/v1/attempts — начало прохождения сценария. */
export async function startAttempt(scenarioId: string): Promise<Attempt> {
  return request<Attempt>('/api/v1/attempts', {
    method: 'POST',
    body: JSON.stringify({ scenarioId }),
  })
}

/** GET /api/v1/attempts/{attemptId} — восстановление состояния попытки. */
export async function restoreAttempt(attemptId: string): Promise<Attempt> {
  return request<Attempt>(`/api/v1/attempts/${encodeURIComponent(attemptId)}`)
}

/** POST /api/v1/attempts/{attemptId}/choices — отправка выбора. */
export async function submitChoice({
  attemptId,
  nodeId,
  choiceId,
  idempotencyKey,
}: SubmitChoiceParams): Promise<Transition> {
  return request<Transition>(`/api/v1/attempts/${encodeURIComponent(attemptId)}/choices`, {
    method: 'POST',
    body: JSON.stringify({ nodeId, choiceId, idempotencyKey }),
  })
}

/** GET /api/v1/progress — прогресс анонимного профиля. */
export async function fetchProgress(): Promise<Progress> {
  return request<Progress>('/api/v1/progress')
}

/** Создаёт тест недели один раз или возвращает уже сохранённый. */
export async function ensureWeeklyTest(): Promise<WeeklyTest> {
  return request<WeeklyTest>('/api/v1/weekly-tests/current', { method: 'POST' })
}

export async function checkWeeklyTestAnswer({
  testId,
  questionId,
  optionIndex,
}: CheckWeeklyTestAnswerParams): Promise<CheckWeeklyTestAnswerResult> {
  return request<CheckWeeklyTestAnswerResult>(
    `/api/v1/weekly-tests/${encodeURIComponent(testId)}/answer`,
    { method: 'POST', body: JSON.stringify({ questionId, optionIndex }) }
  )
}

/** Проверяет ответы на сервере; правильные варианты до этого запроса скрыты. */
export async function submitWeeklyTest({
  testId,
  answers,
}: SubmitWeeklyTestParams): Promise<WeeklyTest> {
  return request<WeeklyTest>(`/api/v1/weekly-tests/${encodeURIComponent(testId)}/submit`, {
    method: 'POST',
    body: JSON.stringify({ answers }),
  })
}

/** Удаляет серверные данные профиля и аннулирует HttpOnly cookie. */
export async function resetProfileData(): Promise<void> {
  await request<{ status: string }>('/api/v1/profile', { method: 'DELETE' })
}
