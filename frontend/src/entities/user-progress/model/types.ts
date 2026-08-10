export interface CompletedMissionRecord {
  bestScore: number
  xpEarned: number
  completedAt: number
  attemptsCount: number
}

export interface CompletedWeeklyTestRecord {
  score: number
  xpEarned: number
  completedAt: number
}

export interface UserProgressState {
  totalXp: number
  level: number
  completedMissions: Record<string, CompletedMissionRecord>
  streakDays: number
  lastActiveDate: string
  unlockedAchievements: string[]
  completedWeeklyTests: Record<string, CompletedWeeklyTestRecord>
  lastDailyRewardDate: string | null

  // Actions
  recordMissionCompletion: (missionId: string, score: number, baseXp?: number) => void
  recordWeeklyTestCompletion: (weekStart: string, score: number, earnedXp: number) => void
  claimDailyReward: (rewardXp?: number) => boolean
  checkAndUpdateStreak: () => void
  resetProgress: () => void
}
