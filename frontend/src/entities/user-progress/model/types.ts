export interface CompletedMissionRecord {
  bestScore: number
  xpEarned: number
  completedAt: number
  attemptsCount: number
}

export interface UserProgressState {
  totalXp: number
  level: number
  completedMissions: Record<string, CompletedMissionRecord>
  streakDays: number
  lastActiveDate: string
  unlockedAchievements: string[]

  // Actions
  recordMissionCompletion: (missionId: string, score: number, baseXp?: number) => void
  checkAndUpdateStreak: () => void
  resetProgress: () => void
}
