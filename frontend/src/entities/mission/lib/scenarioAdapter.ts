import type { Scenario } from '../../../shared/api'
import type { Mission, Difficulty } from '../model/types'
import { MISSIONS } from '../model/missionsData'
import { useUserProgressStore } from '../../user-progress'

const CATEGORY_MAP: Record<string, string> = {
  'buyer-fake-delivery': 'Shipping Scams',
  'buyer-iphone-deposit': 'Payment Scams',
  'buyer-overpayment-scam': 'Payment Scams',
  'buyer-ps5-delivery': 'Shipping Scams',
  'seller-gpu-return-swap': 'Shipping Scams',
  'seller-laptop-courier': 'Payment Scams',
  'seller-payment-already-sent': 'Payment Scams',
}

const GLYPH_MAP: Record<string, string> = {
  'buyer-fake-delivery': '🚚',
  'buyer-iphone-deposit': '📱',
  'buyer-overpayment-scam': '💳',
  'buyer-ps5-delivery': '🎮',
  'seller-gpu-return-swap': '💻',
  'seller-laptop-courier': '🛵',
  'seller-payment-already-sent': '💰',
}

const TONES: Array<'purple' | 'orange' | 'blue' | 'green' | 'red'> = [
  'purple',
  'orange',
  'blue',
  'green',
  'red',
]

export function scenarioToMission(scenario: Scenario, index: number = 0): Mission {
  const difficultyMap: Record<string, Difficulty> = {
    easy: 'Easy',
    medium: 'Medium',
    hard: 'Hard',
  }

  const roleMap: Record<string, 'Buyer' | 'Seller' | 'Both'> = {
    buyer: 'Buyer',
    seller: 'Seller',
  }

  const completedMissions = useUserProgressStore.getState().completedMissions
  const isCompleted = Boolean(completedMissions[scenario.id])

  const diff = difficultyMap[scenario.difficulty] ?? 'Medium'

  return {
    id: scenario.id,
    title: scenario.title,
    description: scenario.description,
    difficulty: diff,
    role: roleMap[scenario.role] ?? 'Both',
    category:
      CATEGORY_MAP[scenario.id] ??
      (scenario.role === 'buyer' ? 'Payment Scams' : 'Shipping Scams'),
    xp: scenario.difficulty === 'hard' ? 20 : scenario.difficulty === 'medium' ? 15 : 10,
    section: index < 4 ? 'featured' : 'new',
    status: isCompleted ? 'completed' : 'not-started',
    glyph: GLYPH_MAP[scenario.id] ?? (scenario.role === 'buyer' ? '🛒' : '📦'),
    tone: TONES[index % TONES.length],
  }
}

export function scenariosToMissions(scenarios: Scenario[]): Mission[] {
  if (!scenarios || scenarios.length === 0) return []

  return scenarios.map((sc, index) => ({
    ...scenarioToMission(sc, index),
    section: index < 4 ? 'featured' : index < 8 ? 'new' : index < 10 ? 'progress' : 'archive',
  }))
}

