export type Difficulty = 'Easy' | 'Medium' | 'Hard'
export type MissionStatus = 'not-started' | 'in-progress' | 'completed' | 'locked'
export type MissionSection = 'featured' | 'new' | 'progress' | 'archive'

export type Mission = {
  id: string
  title: string
  description: string
  difficulty: Difficulty
  role: 'Buyer' | 'Seller' | 'Both'
  category: string
  xp: number
  section: MissionSection
  status: MissionStatus
  asset?: string
  glyph?: string
  tone: 'purple' | 'orange' | 'blue' | 'green' | 'red' | 'slate'
  unlock?: string
}
