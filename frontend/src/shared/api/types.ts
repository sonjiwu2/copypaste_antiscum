export type ScenarioRole = 'buyer' | 'seller';
export type ScenarioDifficulty = 'easy' | 'medium' | 'hard';
export type NodeType = 'message' | 'decision' | 'terminal';
export type Severity = 'safe' | 'warning' | 'dangerous';
export type OutcomeType = 'safe' | 'unsafe';
export type AttemptStatus = 'in_progress' | 'completed';

export interface Scenario {
  id: string;
  version: number;
  slug: string;
  role: ScenarioRole;
  title: string;
  description: string;
  difficulty: ScenarioDifficulty;
  estimatedMinutes: number;
}

export interface ScenarioListResponse {
  scenarios: Scenario[];
}

export interface Choice {
  id: string;
  label: string;
}

export interface AttemptNodeOutcome {
  type: OutcomeType;
  title: string;
  explanation: string;
}

export interface AttemptNode {
  id: string;
  type: NodeType;
  sender?: string;
  text?: string;
  prompt?: string;
  choices?: Choice[];
  outcome?: AttemptNodeOutcome;
}

export interface Consequence {
  severity: Severity;
  title: string;
  explanation: string;
  realWorldRule?: string;
}

export interface Attempt {
  attemptId: string;
  status: AttemptStatus;
  score: number;
  currentNodeId: string;
  revealedNodes: AttemptNode[];
  startedAt: string;
  updatedAt: string;
}

export interface Transition {
  attemptId: string;
  status: AttemptStatus;
  score: number;
  consequence: Consequence;
  revealedNodes: AttemptNode[];
  currentNodeId: string;
  acceptedChoice?: Choice;
}

export interface HealthResponse {
  status: string;
}

export interface ApiErrorDetail {
  code: string;
  message: string;
  requestId?: string;
}

export interface ApiErrorResponse {
  error: ApiErrorDetail;
}

export interface SubmitChoiceParams {
  attemptId: string;
  nodeId: string;
  choiceId: string;
  idempotencyKey?: string;
}
