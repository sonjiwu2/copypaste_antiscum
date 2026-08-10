import type { PixelIconName } from '../../../shared/ui/PixelIcon'

/**
 * Карточка миссии — представление сценария в каталоге.
 *
 * Все перечисления хранят код на английском, а не подпись: код совпадает с
 * контрактом API, участвует в фильтрах и попадает в имена CSS-классов.
 * Русские подписи лежат отдельно, в `labels.ts`.
 */

/** Совпадает с `ScenarioDifficulty` из контракта API. */
export type Difficulty = 'easy' | 'medium' | 'hard'

/** Роль игрока. `both` — сценарий подходит обеим сторонам сделки. */
export type MissionRole = 'buyer' | 'seller' | 'both'

export type ScamCategory =
  | 'fake-listings'
  | 'payment'
  | 'phishing'
  | 'account-takeover'
  | 'shipping'
  | 'impersonation'

export type MissionStatus = 'not-started' | 'in-progress' | 'completed' | 'locked'

/** Блок каталога, в котором карточка показывается по умолчанию. */
export type MissionSection = 'featured' | 'new' | 'progress' | 'archive'

/** Цвет плитки миссии. Только оформление, на логику не влияет. */
export type MissionTone = 'purple' | 'orange' | 'blue' | 'green' | 'red' | 'slate'

export type Mission = {
  id: string
  title: string
  description: string
  difficulty: Difficulty
  role: MissionRole
  category: ScamCategory
  xp: number
  section: MissionSection
  status: MissionStatus
  tone: MissionTone
  /** Иконка на плитке миссии. */
  icon: PixelIconName
  /** Причина блокировки. Заполняется, только если `status === 'locked'`. */
  unlock?: string
}
