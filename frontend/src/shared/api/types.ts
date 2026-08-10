// Публичный контракт API. Источник правды — docs/openapi.yaml.

export type ScenarioRole = 'buyer' | 'seller'
export type ScenarioDifficulty = 'easy' | 'medium' | 'hard'
export type NodeType = 'message' | 'decision' | 'terminal'

/**
 * Кто прислал сообщение узла.
 *
 * `buyer` и `seller` — стороны сделки. `system` — не участник переписки, а факт
 * обстановки: пришла SMS, в приложении нет заказа, денег на счёте нет. В ленте
 * диалога такому тексту не место, он показывается отдельной хроникой событий.
 *
 * Роль игрока здесь не встречается: его слова приходят полем `playerReply`
 * сделанного решения.
 */
export type MessageSender = 'buyer' | 'seller' | 'system'
export type Severity = 'safe' | 'warning' | 'dangerous'
export type OutcomeType = 'safe' | 'unsafe'
export type AttemptStatus = 'in_progress' | 'completed'

export interface Scenario {
  id: string
  version: number
  slug: string
  role: ScenarioRole
  title: string
  description: string
  difficulty: ScenarioDifficulty
  estimatedMinutes: number
  /** Число решений в самой длинной ветке — знаменатель шкалы прохождения. */
  maxDecisions: number
}

export interface ScenarioListResponse {
  scenarios: Scenario[]
}

export interface Choice {
  id: string
  label: string
}

export interface AttemptNodeOutcome {
  type: OutcomeType
  title: string
  explanation: string
}

export interface AttemptNode {
  id: string
  type: NodeType
  /** Заполняется только у узлов типа `message`. */
  sender?: MessageSender
  text?: string
  prompt?: string
  choices?: Choice[]
  outcome?: AttemptNodeOutcome
}

export interface Consequence {
  severity: Severity
  title: string
  explanation: string
  realWorldRule?: string
}

export interface AttemptScenarioRef {
  id: string
  version: number
  slug: string
  role: ScenarioRole
  title: string
}

/** Уже сделанный выбор вместе с раскрытым объяснением. */
export interface AttemptDecision {
  nodeId: string
  choiceId: string
  /** Название действия — то же, что было на кнопке. */
  label: string
  /** Закреплённая за действием реплика игрока: именно она идёт в переписку. */
  playerReply: string
  consequence: Consequence
}

export interface Attempt {
  attemptId: string
  scenario: AttemptScenarioRef
  status: AttemptStatus
  score: number
  outcome?: OutcomeType
  currentNodeId: string
  /** Полный список раскрытых узлов: по нему восстанавливается экран. */
  revealedNodes: AttemptNode[]
  decisions: AttemptDecision[]
  startedAt: string
  updatedAt: string
  completedAt?: string
}

export interface AcceptedChoice {
  nodeId: string
  choiceId: string
  label: string
  playerReply: string
}

export interface Transition {
  attemptId: string
  status: AttemptStatus
  score: number
  outcome?: OutcomeType
  acceptedChoice: AcceptedChoice
  consequence: Consequence
  /** Только узлы, раскрытые этим шагом, а не вся история попытки. */
  revealedNodes: AttemptNode[]
  currentNodeId: string
  completedAt?: string
}

export interface ProgressSummary {
  completedAttempts: number
  completedScenarios: number
  averageScore: number
  bestScore: number
  latestScore: number
}

export interface ProgressActiveAttempt {
  attemptId: string
  scenarioId: string
  updatedAt: string
}

export interface ProgressScenario {
  scenarioId: string
  role?: ScenarioRole
  title?: string
  attempts: number
  lastScore: number
  bestScore: number
  previousScore?: number
  improvement?: number
  lastCompletedAt?: string
  activeAttemptId?: string
}

export interface ProgressSkill {
  code: string
  value: number
  positive: number
  negative: number
  decisions: number
}

export interface ProgressWeakRiskTag {
  code: string
  mistakes: number
  weightedSeverity: number
  lastOccurredAt: string
}

export interface ProgressHistoryEntry {
  attemptId: string
  scenarioId: string
  role?: ScenarioRole
  title?: string
  score: number
  outcome?: OutcomeType
  startedAt: string
  completedAt: string
  decisions: number
}

export type RecommendationReason = 'not_completed' | 'improve_score'

export interface ProgressRecommendation {
  scenarioId: string
  role: ScenarioRole
  title: string
  difficulty: ScenarioDifficulty
  estimatedMinutes: number
  reason: RecommendationReason
}

export interface Progress {
  summary: ProgressSummary
  activeAttempts: ProgressActiveAttempt[]
  scenarioProgress: ProgressScenario[]
  skills: ProgressSkill[]
  weakRiskTags: ProgressWeakRiskTag[]
  recentAttempts: ProgressHistoryEntry[]
  recommendedScenarios: ProgressRecommendation[]
}

export interface HealthResponse {
  status: string
}

export interface ApiErrorDetail {
  code: string
  message: string
  requestId?: string
}

export interface ApiErrorResponse {
  error: ApiErrorDetail
}

export interface SubmitChoiceParams {
  attemptId: string
  nodeId: string
  choiceId: string
  /** Повтор запроса с тем же ключом возвращает тот же результат шага. */
  idempotencyKey: string
}

export type WeeklyTestSource = 'groq' | 'grok' | 'fallback'

export interface WeeklyTestQuestion {
  id: string
  prompt: string
  options: string[]
  riskTag: string
  difficulty: ScenarioDifficulty
}

export interface WeeklyTestReview {
  questionId: string
  selectedIndex: number
  correctIndex: number
  correct: boolean
  explanation: string
}

export interface WeeklyTestResult {
  correct: number
  total: number
  score: number
  passed: boolean
  timedOut: boolean
  livesLeft: number
  earnedXp: number
  completedAt: string
  review: WeeklyTestReview[]
}

export interface WeeklyTestRules {
  questionCount: number
  maxLives: number
  passingCorrect: number
  rewardXp: number
  durationSeconds: number
}

export interface WeeklyTest {
  testId: string
  weekStart: string
  title: string
  intro: string
  source: WeeklyTestSource
  completed: boolean
  expiresAt: string
  nextAvailableAt?: string
  rules: WeeklyTestRules
  questions: WeeklyTestQuestion[]
  result?: WeeklyTestResult
}

export interface WeeklyTestAnswer {
  questionId: string
  optionIndex: number
}

export interface SubmitWeeklyTestParams {
  testId: string
  answers: WeeklyTestAnswer[]
}

export interface CheckWeeklyTestAnswerParams extends WeeklyTestAnswer {
  testId: string
}

export interface CheckWeeklyTestAnswerResult {
  correct: boolean
}
