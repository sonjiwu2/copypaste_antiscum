import type { Scenario } from '../../../shared/api'
import type { PixelIconName } from '../../../shared/ui/PixelIcon'
import type { Mission, MissionSection, MissionTone, ScamCategory } from '../model/types'

/**
 * Тема и иконка сценария. Бэкенд их не присылает: это оформление каталога,
 * а не часть контракта. Незнакомый сценарий получает значения по роли.
 */
const SCENARIO_PRESENTATION: Record<string, { category: ScamCategory; icon: PixelIconName }> = {
  'buyer-airpods-counterfeit': { category: 'fake-listings', icon: 'earbuds' },
  'buyer-fake-delivery': { category: 'shipping', icon: 'delivery' },
  'buyer-gpu-hidden-repair': { category: 'fake-listings', icon: 'gpu' },
  'buyer-iphone-deposit': { category: 'payment', icon: 'phone' },
  'buyer-macbook-corporate-lock': { category: 'account-takeover', icon: 'laptop-lock' },
  'buyer-overpayment-scam': { category: 'payment', icon: 'listing' },
  'buyer-ps5-delivery': { category: 'shipping', icon: 'console' },
  'buyer-switch-prepayment': { category: 'payment', icon: 'handheld' },
  'seller-fake-payment-email': { category: 'phishing', icon: 'email-hook' },
  'seller-gpu-return-swap': { category: 'shipping', icon: 'monitor' },
  'seller-laptop-courier': { category: 'payment', icon: 'courier' },
  'seller-payment-already-sent': { category: 'payment', icon: 'payment' },
  'seller-qr-payment-trap': { category: 'payment', icon: 'qr' },
  'seller-sms-code-payment': { category: 'account-takeover', icon: 'sms' },
  'seller-third-party-overpayment': { category: 'payment', icon: 'camera' },
}

const TONES: MissionTone[] = ['purple', 'orange', 'blue', 'green', 'red']

const XP_BY_DIFFICULTY = {
  easy: 10,
  medium: 15,
  hard: 20,
} as const

/** Награда за сценарий указанной сложности при идеальном прохождении. */
export function missionXpFor(difficulty: string): number {
  return XP_BY_DIFFICULTY[difficulty as keyof typeof XP_BY_DIFFICULTY] ?? XP_BY_DIFFICULTY.medium
}

/**
 * Переводит сценарий каталога в карточку миссии.
 *
 * Список пройденных сценариев передаётся аргументом, а не читается из store:
 * иначе результат зависел бы от невидимого состояния и не пересчитывался бы
 * при его изменении.
 */
export function scenarioToMission(
  scenario: Scenario,
  index: number = 0,
  completedIds: ReadonlySet<string> = new Set()
): Mission {
  const presentation = SCENARIO_PRESENTATION[scenario.id]
  const isBuyer = scenario.role === 'buyer'

  return {
    id: scenario.id,
    title: scenario.title,
    description: scenario.description,
    difficulty: scenario.difficulty,
    role: scenario.role,
    category: presentation?.category ?? (isBuyer ? 'payment' : 'shipping'),
    xp: missionXpFor(scenario.difficulty),
    section: sectionFor(index),
    status: completedIds.has(scenario.id) ? 'completed' : 'not-started',
    tone: TONES[index % TONES.length],
    icon: presentation?.icon ?? (isBuyer ? 'role-buyer' : 'role-seller'),
  }
}

export function scenariosToMissions(
  scenarios: Scenario[],
  completedIds: ReadonlySet<string> = new Set()
): Mission[] {
  return scenarios.map((scenario, index) => scenarioToMission(scenario, index, completedIds))
}

/** Сколько сценариев каталога попадает в блок «рекомендуемые». */
const FEATURED_SECTION_SIZE = 5

/**
 * Раскладывает каталог по блокам страницы: первые сценарии попадают в
 * «рекомендуемые», остальные — в «новые».
 *
 * Блоки «в процессе» и «заблокированные» по позиции не заполняются: они
 * описывают состояние прохождения, а не место в каталоге. Иначе рост каталога
 * отправлял бы доступные сценарии в раздел «заблокированные».
 */
function sectionFor(index: number): MissionSection {
  return index < FEATURED_SECTION_SIZE ? 'featured' : 'new'
}
