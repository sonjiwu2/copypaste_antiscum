import type { Attempt, AttemptDecision, ScenarioRole, Severity } from '../../../shared/api'

export type ChoiceTone = 'safe' | 'neutral' | 'risky'

export type ChatMessage = {
  id: string
  /** Кто говорит: покупатель или продавец. Система репликами не говорит. */
  role: ScenarioRole
  /** Реплика самого игрока, а не собеседника. */
  own: boolean
  text: string
  time: string
  /** Заполняется только у реплик игрока: показывает разбор сделанного выбора. */
  tone?: ChoiceTone
}

/** Факт обстановки: пришла SMS, заказа нет в приложении, деньги не поступили. */
export type WorldEvent = {
  id: string
  text: string
}

export const PRACTICAL_TIPS: Array<{ title: string; advice: string }> = [
  { title: 'Проверка продавца', advice: 'Проверяйте дату регистрации аккаунта и реальные отзывы.' },
  { title: 'Безопасные платежи', advice: 'Платите только через безопасную сделку площадки.' },
  { title: 'Анализ цены', advice: 'Сравнивайте предложение с рыночными ценами.' },
  { title: 'Жалоба и защита', advice: 'Блокируйте профиль при подозрительных просьбах.' },
]

/** Ключ, по которому попытка сценария переживает перезагрузку страницы. */
export function attemptStorageKey(scenarioId: string): string {
  return `antiscam.attempt.${scenarioId}`
}

export function readStoredAttemptId(scenarioId: string): string | null {
  try {
    return localStorage.getItem(attemptStorageKey(scenarioId))
  } catch {
    // Приватный режим браузера запрещает хранилище: попытка просто начнётся заново.
    return null
  }
}

export function writeStoredAttemptId(scenarioId: string, attemptId: string | null): void {
  try {
    if (attemptId === null) {
      localStorage.removeItem(attemptStorageKey(scenarioId))
    } else {
      localStorage.setItem(attemptStorageKey(scenarioId), attemptId)
    }
  } catch {
    // Без хранилища попытка живёт до перезагрузки страницы.
  }
}

/** Нужно ли заменить сохранённую попытку новой версией сценария. */
export function isAttemptVersionOutdated(
  attempt: Attempt | undefined,
  currentScenarioVersion: number | undefined
): boolean {
  return Boolean(
    attempt &&
      attempt.status === 'in_progress' &&
      currentScenarioVersion &&
      attempt.scenario.version < currentScenarioVersion
  )
}

/**
 * Время сообщения в ленте. Метки декоративные: сервер время реплик не хранит,
 * поэтому диалог отсчитывается по минуте на сообщение от условного начала.
 */
export function clockTime(offset: number): string {
  const startMinutes = 10 * 60 + 32
  const total = startMinutes + offset
  const hours = Math.floor(total / 60) % 24

  return `${String(hours).padStart(2, '0')}:${String(total % 60).padStart(2, '0')}`
}

const TONE_BY_SEVERITY: Record<Severity, ChoiceTone> = {
  safe: 'safe',
  warning: 'neutral',
  dangerous: 'risky',
}

export function toneOf(severity: Severity): ChoiceTone {
  return TONE_BY_SEVERITY[severity] ?? 'neutral'
}

/**
 * Собирает ленту переписки из состояния попытки.
 *
 * В ленту попадают только слова двух сторон сделки. Заранее подготовленные
 * вопросы и связующие фразы игрока приходят обычными message-узлами. Его ход
 * в точке решения показывается закреплённой за выбором репликой `playerReply`,
 * а не подписью кнопки. Факты обстановки и итог сделки в переписке не
 * участвуют — их возвращает buildEvents и итоговое окно.
 */
export function buildMessages(attempt: Attempt): ChatMessage[] {
  const decisionsByNode = new Map<string, AttemptDecision>(
    attempt.decisions.map((decision) => [decision.nodeId, decision])
  )
  const playerRole = attempt.scenario.role
  const opponentRole: ScenarioRole = playerRole === 'buyer' ? 'seller' : 'buyer'
  const messages: ChatMessage[] = []

  for (const node of attempt.revealedNodes) {
    if (node.type === 'message' && node.sender !== 'system' && node.text) {
      const role = node.sender ?? opponentRole

      messages.push({
        id: node.id,
        role,
        // Message-узел может быть заранее подготовленной репликой любой
        // стороны; собственная роль всегда рисуется справа.
        own: role === playerRole,
        text: node.text,
        time: clockTime(messages.length),
      })

      continue
    }

    const decision = decisionsByNode.get(node.id)
    if (node.type === 'decision' && decision?.playerReply) {
      messages.push({
        id: `decision-${decision.nodeId}`,
        role: playerRole,
        own: true,
        text: decision.playerReply,
        time: clockTime(messages.length),
        tone: toneOf(decision.consequence.severity),
      })
    }
  }

  return messages
}

/**
 * Собирает хронику событий: то, что произошло вокруг сделки.
 *
 * Система не участник переписки и ничего собеседнику не говорит, поэтому её
 * текст показывается отдельно от диалога.
 */
export function buildEvents(attempt: Attempt): WorldEvent[] {
  const events: WorldEvent[] = []

  for (const node of attempt.revealedNodes) {
    if (node.type === 'message' && node.sender === 'system' && node.text) {
      events.push({ id: node.id, text: node.text })
    }
  }

  return events
}

export type RiskGrade = 'low' | 'medium' | 'high'

export const RISK_LABELS: Record<RiskGrade, string> = {
  low: 'НИЗКИЙ РИСК',
  medium: 'СРЕДНИЙ РИСК',
  high: 'ВЫСОКИЙ РИСК',
}

export const RISK_HINTS: Record<RiskGrade, string> = {
  low: 'Разговор идёт в безопасном русле.',
  medium: 'Собеседник использует подозрительные приёмы.',
  high: 'Прекратите сделку и пожалуйтесь на профиль.',
}

/** Уровень риска по текущему результату попытки. */
export function riskGradeOf(score: number): RiskGrade {
  if (score >= 80) {
    return 'low'
  }

  return score >= 50 ? 'medium' : 'high'
}

/**
 * Положение стрелки на шкале риска в процентах: 0 — безопасно, 100 — опасно.
 *
 * Верхняя граница чуть меньше 100: иначе стрелка выходит за правый край шкалы.
 */
export function riskMarkerPercent(score: number): number {
  const clampedScore = Math.min(100, Math.max(0, score))

  return Math.min(96, 100 - clampedScore)
}
